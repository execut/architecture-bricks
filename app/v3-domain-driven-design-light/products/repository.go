package products

import (
    "context"
    "encoding/json"
    "errors"
    "os"
    "time"

    "architecture-bricks/app/v3-domain-driven-design-light/products/domain"

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

func (r *Repository) Create(
    ctx context.Context,
    product *domain.Product,
    events ...domain.ProductHistoryRow,
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
	`, product.ID(), product.Name(), product.UserID(), string(product.Status()), approveReason, rejectionReason)
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
		`, event.ID, product.ID(), event.Event, string(payload))
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

func (r *Repository) Update(
    ctx context.Context,
    product *domain.Product,
    events ...domain.ProductHistoryRow,
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
    if product.ApproveReason() != "" {
        reason := string(product.ApproveReason())
        approveReason = &reason
    }
    if product.RejectionReason() != "" {
        reason := string(product.RejectionReason())
        rejectionReason = &reason
    }

    _, err = tx.Exec(ctx, `
		UPDATE product
		SET name = $1, status = $2, approve_reason = $3, rejection_reason = $4
		WHERE id = $5
	`, product.Name(), string(product.Status()), approveReason, rejectionReason, product.ID())
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
		`, event.ID, product.ID(), event.Event, string(payload))
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

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
    var productID string
    var productName string
    var userID string
    var status string
    var approveReason, rejectionReason *string

    err := r.pool.QueryRow(ctx, `
		SELECT id::text, name, user_id::text, status, approve_reason, rejection_reason
		FROM product
		WHERE id = $1
	`, id).Scan(&productID, &productName, &userID, &status, &approveReason, &rejectionReason)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrProductNotFound
        }

        return nil, err
    }

    var approveReasonVal domain.ProductApproveReason
    if approveReason != nil {
        approveReasonVal = domain.ProductApproveReason(*approveReason)
    }
    var rejectionReasonVal domain.ProductRejectionReason
    if rejectionReason != nil {
        rejectionReasonVal = domain.ProductRejectionReason(*rejectionReason)
    }

    return domain.LoadProduct(
        productID, userID, productName,
        domain.ProductStatus(status),
        approveReasonVal,
        rejectionReasonVal,
    ), nil
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
