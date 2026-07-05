package products

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"architecture-bricks/app/v5-value-objects/products/domain"
	"architecture-bricks/contract"
	biz "architecture-bricks/pkg/optimistic-locking/business-events/value-objects/domain"
	vo "architecture-bricks/pkg/optimistic-locking/value-objects/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ domain.Repository = (*Repository)(nil)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(ctx context.Context) (*Repository, error) {
	pool, err := newDBConnection(ctx)
	if err != nil {
		return nil, err
	}

	return &Repository{pool: pool}, nil
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

func (r *Repository) Load(ctx context.Context, product *domain.Product) error {
	var productID string
	var productName string
	var userID string
	var status string
	var approveReason, rejectionReason *string
	var dbVersion int

	err := r.pool.QueryRow(ctx, `
		SELECT id::text, name, user_id::text, status, approve_reason, rejection_reason, version
		FROM product
		WHERE id = $1
	`, product.ID()).Scan(&productID, &productName, &userID, &status, &approveReason, &rejectionReason, &dbVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrProductNotFound
		}

		return err
	}

	version, err := vo.NewVersion(dbVersion)
	if err != nil {
		return err
	}

	product.AddAndApplyEvent(biz.NewEvent(version, domain.ProductCreatedPayload{
		Name: productName,
		User: userID,
	}))

	if status == string(domain.ProductStatusApproved) && approveReason != nil {
		if *approveReason == string(domain.ProductApproveReasonAuto) {
			product.AddAndApplyEvent(biz.NewEvent(version, domain.ProductAutoApprovedPayload{
				ApproveReason: *approveReason,
			}))
		} else {
			product.AddAndApplyEvent(biz.NewEvent(version, domain.ProductApprovedPayload{
				ApproveReason: *approveReason,
			}))
		}
	} else if status == string(domain.ProductStatusRejected) && rejectionReason != nil {
		product.AddAndApplyEvent(biz.NewEvent(version, domain.ProductRejectedPayload{
			RejectionReason: *rejectionReason,
		}))
	}

	product.CleanEventList()

	return nil
}

func (r *Repository) Save(ctx context.Context, product *domain.Product) error {
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

	productID := product.ID()

	for _, event := range product.EventList() {
		switch payload := event.Payload().(type) {
		case domain.ProductCreatedPayload:
			var approveReason, rejectionReason *string
			if product.ApproveReason() != "" {
				reason := string(product.ApproveReason())
				approveReason = &reason
			}
			if product.RejectionReason() != "" {
				reason := string(product.RejectionReason())
				rejectionReason = &reason
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO product (id, name, user_id, status, approve_reason, rejection_reason)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, productID, product.Name(), product.UserID(), string(product.Status()), approveReason, rejectionReason)
			if err != nil {
				return mapCreateProductError(err)
			}

			eventPayload, marshalErr := marshalEventPayload(domain.ProductHistoryRow{Name: payload.Name})
			if marshalErr != nil {
				return marshalErr
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO event (id, entry_id, event, payload)
				VALUES ($1, $2, $3, $4::jsonb)
			`, uuid.NewString(), productID, "product_created", string(eventPayload))
			if err != nil {
				return err
			}

		case domain.ProductRenamedPayload:
			tag, err := tx.Exec(ctx, `
				UPDATE product
				SET name = $1, version = version + 1
				WHERE id = $2 AND version = $3
			`, product.Name(), productID, event.Version().Value()-1)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return contract.ErrProductAlreadyChanged
			}

			eventPayload, marshalErr := marshalEventPayload(domain.ProductHistoryRow{Name: product.Name()})
			if marshalErr != nil {
				return marshalErr
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO event (id, entry_id, event, payload)
				VALUES ($1, $2, $3, $4::jsonb)
			`, uuid.NewString(), productID, "product_updated", string(eventPayload))
			if err != nil {
				return err
			}

		case domain.ProductApprovedPayload:
			tag, err := tx.Exec(ctx, `
				UPDATE product
				SET status = $1, approve_reason = $2, rejection_reason = NULL, version = version + 1
				WHERE id = $3 AND version = $4
			`, string(product.Status()), string(product.ApproveReason()), productID, event.Version().Value()-1)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return contract.ErrProductAlreadyChanged
			}

			eventPayload, marshalErr := marshalEventPayload(domain.ProductHistoryRow{
				Name:          product.Name(),
				ModeratorID:   payload.ModeratorID,
				ApproveReason: payload.ApproveReason,
			})
			if marshalErr != nil {
				return marshalErr
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO event (id, entry_id, event, payload)
				VALUES ($1, $2, $3, $4::jsonb)
			`, uuid.NewString(), productID, "product_approved", string(eventPayload))
			if err != nil {
				return err
			}

		case domain.ProductRejectedPayload:
			tag, err := tx.Exec(ctx, `
				UPDATE product
				SET status = $1, rejection_reason = $2, approve_reason = NULL, version = version + 1
				WHERE id = $3 AND version = $4
			`, string(product.Status()), string(product.RejectionReason()), productID, event.Version().Value()-1)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return contract.ErrProductAlreadyChanged
			}

			eventPayload, marshalErr := marshalEventPayload(domain.ProductHistoryRow{
				Name:            product.Name(),
				ModeratorID:     payload.ModeratorID,
				RejectionReason: payload.RejectionReason,
			})
			if marshalErr != nil {
				return marshalErr
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO event (id, entry_id, event, payload)
				VALUES ($1, $2, $3, $4::jsonb)
			`, uuid.NewString(), productID, "product_rejected", string(eventPayload))
			if err != nil {
				return err
			}

		case domain.ProductAutoApprovedPayload:
			tag, err := tx.Exec(ctx, `
				UPDATE product
				SET status = $1, approve_reason = $2, rejection_reason = NULL, version = version + 1
				WHERE id = $3 AND version = $4
			`, string(product.Status()), string(product.ApproveReason()), productID, event.Version().Value()-1)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return contract.ErrProductAlreadyChanged
			}

			eventPayload, marshalErr := marshalEventPayload(domain.ProductHistoryRow{
				Name:          product.Name(),
				ApproveReason: payload.ApproveReason,
			})
			if marshalErr != nil {
				return marshalErr
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO event (id, entry_id, event, payload)
				VALUES ($1, $2, $3, $4::jsonb)
			`, uuid.NewString(), productID, "product_auto_approved", string(eventPayload))
			if err != nil {
				return err
			}
		}
	}

	product.CleanEventList()

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (r *Repository) GetHistory(
	ctx context.Context,
	id string,
) ([]domain.ProductHistoryRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, event, payload
		FROM event
		WHERE entry_id = $1
		ORDER BY created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]domain.ProductHistoryRow, 0)
	for rows.Next() {
		row := domain.ProductHistoryRow{}
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
			row.ApproveReason = approveReason
		}
		if rejectionReason, ok := payloadData["rejection_reason"].(string); ok {
			row.RejectionReason = rejectionReason
		}

		history = append(history, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func marshalEventPayload(event domain.ProductHistoryRow) ([]byte, error) {
	payloadMap := map[string]any{"name": event.Name}
	if event.ModeratorID != "" {
		payloadMap["moderator_id"] = event.ModeratorID
	}
	if event.ApproveReason != "" {
		payloadMap["approve_reason"] = event.ApproveReason
	}
	if event.RejectionReason != "" {
		payloadMap["rejection_reason"] = event.RejectionReason
	}
	return json.Marshal(payloadMap)
}

func mapCreateProductError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrProductAlreadyExists
	}

	return err
}
