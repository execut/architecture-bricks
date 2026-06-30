package products

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"architecture-bricks/contract"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context) (*PostgresRepository, error) {
	pool, err := newDBConnection(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{pool: pool}, nil
}

func newDBConnection(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	const attempts = 20
	const delay = 500 * time.Millisecond

	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		var pool *pgxpool.Pool
		pool, err = pgxpool.New(ctx, databaseURL)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				return pool, nil
			}

			pool.Close()
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	if err == nil {
		err = errors.New("could not connect to database")
	}

	return nil, err
}

func (r *PostgresRepository) CreateProduct(
	ctx context.Context,
	product contract.Product,
	events ...contract.ProductHistoryRow,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	var approveReason, rejectionReason *string
	if product.ApproveReason != "" {
		reason := string(product.ApproveReason)
		approveReason = &reason
	}
	if product.RejectionReason != "" {
		reason := string(product.RejectionReason)
		rejectionReason = &reason
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO product (id, name, user_id, status, approve_reason, rejection_reason)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, product.ID, product.Name, product.UserID, string(product.Status), approveReason, rejectionReason)
	if err != nil {
		return mapCreateProductError(err)
	}

	for _, event := range events {
		payload, marshalErr := marshalEventPayload(event)
		if marshalErr != nil {
			return marshalErr
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO event (id, entry_id, event, payload)
            VALUES ($1, $2, $3, $4::jsonb)
        `, event.ID, product.ID, event.Event, string(payload))
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *PostgresRepository) UpdateProduct(
	ctx context.Context,
	product contract.Product,
	version int,
	events ...contract.ProductHistoryRow,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if product.Status == contract.ProductStatusApproved {
		tag, updateErr := tx.Exec(ctx, `
            UPDATE product
            SET name = $1, status = $2, approve_reason = $3, rejection_reason = NULL, version = version + 1
            WHERE id = $4 AND version = $5
        `, product.Name, string(product.Status), string(product.ApproveReason), product.ID, version)
		err = updateErr
		if err == nil && tag.RowsAffected() == 0 {
			return contract.ErrProductAlreadyChanged
		}
	} else {
		tag, updateErr := tx.Exec(ctx, `
            UPDATE product
            SET name = $1, version = version + 1
            WHERE id = $2 AND version = $3
        `, product.Name, product.ID, version)
		err = updateErr
		if err == nil && tag.RowsAffected() == 0 {
			return contract.ErrProductAlreadyChanged
		}
	}
	if err != nil {
		return err
	}

	for _, event := range events {
		payload, marshalErr := marshalEventPayload(event)
		if marshalErr != nil {
			return marshalErr
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO event (id, entry_id, event, payload)
            VALUES ($1, $2, $3, $4::jsonb)
        `, event.ID, product.ID, event.Event, string(payload))
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *PostgresRepository) GetProduct(ctx context.Context, productID string) (contract.Product, int, error) {
	product := contract.Product{}
	var status string
	var approveReason, rejectionReason *string
	var version int
	err := r.pool.QueryRow(ctx, `
        SELECT id::text, name, user_id::text, status, approve_reason, rejection_reason, version
        FROM product
        WHERE id = $1
    `, productID).Scan(&product.ID, &product.Name, &product.UserID, &status, &approveReason, &rejectionReason, &version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contract.Product{}, 0, contract.ErrProductNotFound
		}

		return contract.Product{}, 0, err
	}

	product.Status = contract.ProductStatus(status)
	if approveReason != nil {
		product.ApproveReason = contract.ProductApproveReason(*approveReason)
	}
	if rejectionReason != nil {
		product.RejectionReason = contract.ProductRejectionReason(*rejectionReason)
	}

	return product, version, nil
}

func (r *PostgresRepository) ApproveProduct(
	ctx context.Context,
	productID string,
	moderatorID string,
	version int,
	event contract.ProductHistoryRow,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	tag, err := tx.Exec(ctx, `
        UPDATE product
        SET status = $1, approve_reason = $2, rejection_reason = NULL, version = version + 1
        WHERE id = $3 AND version = $4
    `, string(contract.ProductStatusApproved), string(contract.ProductApproveReasonModerator), productID, version)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return contract.ErrProductAlreadyChanged
	}

	payload, err := marshalEventPayload(event)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO event (id, entry_id, event, payload)
        VALUES ($1, $2, $3, $4::jsonb)
    `, event.ID, productID, event.Event, string(payload))
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *PostgresRepository) RejectProduct(
	ctx context.Context,
	productID string,
	moderatorID string,
	version int,
	event contract.ProductHistoryRow,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	tag, err := tx.Exec(ctx, `
        UPDATE product
        SET status = $1, rejection_reason = $2, approve_reason = NULL, version = version + 1
        WHERE id = $3 AND version = $4
    `, string(contract.ProductStatusRejected), string(contract.ProductRejectionReasonModerator), productID, version)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return contract.ErrProductAlreadyChanged
	}

	payload, err := marshalEventPayload(event)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO event (id, entry_id, event, payload)
        VALUES ($1, $2, $3, $4::jsonb)
    `, event.ID, productID, event.Event, string(payload))
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *PostgresRepository) ProductHistory(ctx context.Context, productID string) ([]contract.ProductHistoryRow, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id::text, event, payload
        FROM event
        WHERE entry_id = $1
        ORDER BY created_at
    `, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]contract.ProductHistoryRow, 0)
	for rows.Next() {
		row := contract.ProductHistoryRow{}
		var payload []byte

		if err := rows.Scan(&row.ID, &row.Event, &payload); err != nil {
			return nil, err
		}

		var payloadData map[string]any
		if err := json.Unmarshal(payload, &payloadData); err != nil {
			return nil, err
		}

		if name, ok := payloadData["name"].(string); ok {
			row.Name = name
		}
		if moderatorID, ok := payloadData["moderator_id"].(string); ok {
			row.ModeratorID = moderatorID
		}
		if approveReason, ok := payloadData["approve_reason"].(string); ok {
			row.ApproveReason = contract.ProductApproveReason(approveReason)
		}
		if rejectionReason, ok := payloadData["rejection_reason"].(string); ok {
			row.RejectionReason = contract.ProductRejectionReason(rejectionReason)
		}

		history = append(history, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func marshalEventPayload(event contract.ProductHistoryRow) ([]byte, error) {
	payloadMap := map[string]any{"name": event.Name}
	if event.ModeratorID != "" {
		payloadMap["moderator_id"] = event.ModeratorID
	}
	if event.ApproveReason != "" {
		payloadMap["approve_reason"] = string(event.ApproveReason)
	}
	if event.RejectionReason != "" {
		payloadMap["rejection_reason"] = string(event.RejectionReason)
	}
	return json.Marshal(payloadMap)
}
