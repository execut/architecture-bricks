package products

import (
    "context"
    "encoding/json"
    "errors"
    "os"
    "strings"
    "time"

    "architecture-bricks/contract"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var _ contract.Application = (*Service)(nil)

type Service struct {
    pool *pgxpool.Pool
}

func NewService(ctx context.Context) (*Service, error) {
    pool, err := newDBConnection(ctx)
    if err != nil {
        return nil, err
    }

    return &Service{pool: pool}, nil
}

func (s *Service) ApproveProduct(ctx context.Context, params contract.ApproveProduct) error {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return err
    }

    _, err = parseUUID(params.ModeratorID, contract.ErrInvalidUserID)
    if err != nil {
        return err
    }

    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return err
    }

    committed := false
    defer func() {
        if !committed {
            _ = tx.Rollback(ctx)
        }
    }()

    // Get current product name before updating
    var currentName string
    err = tx.QueryRow(ctx, `SELECT name FROM product WHERE id = $1 FOR UPDATE`, parsedProductID).Scan(&currentName)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return contract.ErrProductNotFound
        }
        return err
    }

    _, err = tx.Exec(ctx, `
		UPDATE product
		SET status = $1, approve_reason = $2, rejection_reason = NULL
		WHERE id = $3
	`, string(contract.ProductStatusApproved), string(contract.ProductApproveReasonModerator), parsedProductID)
    if err != nil {
        return err
    }

    payload, err := json.Marshal(map[string]string{
        "name":           currentName,
        "moderator_id":   params.ModeratorID,
        "approve_reason": string(contract.ProductApproveReasonModerator),
    })
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `
		INSERT INTO event (id, entry_id, event, payload)
		VALUES ($1, $2, $3, $4::jsonb)
	`, uuid.NewString(), parsedProductID, contract.EventProductApproved, string(payload))
    if err != nil {
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return err
    }

    committed = true

    return nil
}

func (s *Service) RejectProduct(ctx context.Context, params contract.RejectProduct) error {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return err
    }

    _, err = parseUUID(params.ModeratorID, contract.ErrInvalidUserID)
    if err != nil {
        return err
    }

    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return err
    }

    committed := false
    defer func() {
        if !committed {
            _ = tx.Rollback(ctx)
        }
    }()

    // Get current product name before updating
    var currentName string
    err = tx.QueryRow(ctx, `SELECT name FROM product WHERE id = $1 FOR UPDATE`, parsedProductID).Scan(&currentName)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return contract.ErrProductNotFound
        }
        return err
    }

    _, err = tx.Exec(ctx, `
		UPDATE product
		SET status = $1, rejection_reason = $2, approve_reason = NULL
		WHERE id = $3
	`, string(contract.ProductStatusRejected), string(contract.ProductRejectionReasonModerator), parsedProductID)
    if err != nil {
        return err
    }

    payload, err := json.Marshal(map[string]string{
        "name":             currentName,
        "moderator_id":     params.ModeratorID,
        "rejection_reason": string(contract.ProductRejectionReasonModerator),
    })
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `
		INSERT INTO event (id, entry_id, event, payload)
		VALUES ($1, $2, $3, $4::jsonb)
	`, uuid.NewString(), parsedProductID, contract.EventProductRejected, string(payload))
    if err != nil {
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return err
    }

    committed = true

    return nil
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

func (s *Service) CreateProduct(ctx context.Context, params contract.CreateProduct) error {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return err
    }

    parsedUserID, err := parseUUID(params.UserID, contract.ErrInvalidUserID)
    if err != nil {
        return err
    }

    if strings.TrimSpace(params.Name) == "" {
        return contract.ErrProductNameRequired
    }

    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return err
    }

    committed := false
    defer func() {
        if !committed {
            _ = tx.Rollback(ctx)
        }
    }()

    name := strings.TrimSpace(params.Name)

    autoApprove := isAutoApprovable(name)

    status := string(contract.ProductStatusPending)
    var approveReason *string
    if autoApprove {
        status = string(contract.ProductStatusApproved)
        reason := string(contract.ProductApproveReasonAuto)
        approveReason = &reason
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO product (id, name, user_id, status, approve_reason, rejection_reason)
        VALUES ($1, $2, $3, $4, $5, NULL)
    `, parsedProductID, name, parsedUserID, status, approveReason)
    if err != nil {
        return mapCreateProductError(err)
    }

    payload, err := json.Marshal(map[string]string{"name": name})
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO event (id, entry_id, event, payload)
        VALUES ($1, $2, $3, $4::jsonb)
    `, uuid.NewString(), parsedProductID, contract.EventProductCreated, string(payload))
    if err != nil {
        return err
    }

    if autoApprove {
        autoPayload, err := json.Marshal(map[string]string{
            "name":           name,
            "approve_reason": string(contract.ProductApproveReasonAuto),
        })
        if err != nil {
            return err
        }

        _, err = tx.Exec(ctx, `
			INSERT INTO event (id, entry_id, event, payload)
			VALUES ($1, $2, $3, $4::jsonb)
		`, uuid.NewString(), parsedProductID, contract.EventProductAutoApproved, string(autoPayload))
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

func isAutoApprovable(name string) bool {
    return name == "Кот" || name == "Собака"
}

func (s *Service) UpdateProduct(ctx context.Context, params contract.UpdateProduct) error {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return err
    }

    _, err = parseUUID(params.UserID, contract.ErrInvalidUserID)
    if err != nil {
        return err
    }

    var existingName string
    err = s.pool.QueryRow(ctx, `
        SELECT name
        FROM product
        WHERE id = $1
    `, parsedProductID).Scan(&existingName)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return contract.ErrProductNotFound
        }

        return err
    }

    newName := strings.TrimSpace(params.Name)

    if existingName == newName {
        return contract.ErrProductNameNotChanged
    }

    if newName == "" {
        return contract.ErrProductNameRequired
    }

    autoApprove := isAutoApprovable(newName)

    tx, err := s.pool.Begin(ctx)
    if err != nil {
        return err
    }

    committed := false
    defer func() {
        if !committed {
            _ = tx.Rollback(ctx)
        }
    }()

    if autoApprove {
        _, err = tx.Exec(ctx, `
			UPDATE product
			SET name = $1, status = $2, approve_reason = $3, rejection_reason = NULL
			WHERE id = $4
		`, newName, string(contract.ProductStatusApproved), string(contract.ProductApproveReasonAuto), parsedProductID)
    } else {
        _, err = tx.Exec(ctx, `
			UPDATE product
			SET name = $1
			WHERE id = $2
		`, newName, parsedProductID)
    }
    if err != nil {
        return err
    }

    payload, err := json.Marshal(map[string]string{"name": newName})
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO event (id, entry_id, event, payload)
        VALUES ($1, $2, $3, $4::jsonb)
    `, uuid.NewString(), parsedProductID, contract.EventProductUpdated, string(payload))
    if err != nil {
        return err
    }

    if autoApprove {
        autoPayload, err := json.Marshal(map[string]string{
            "name":           newName,
            "approve_reason": string(contract.ProductApproveReasonAuto),
        })
        if err != nil {
            return err
        }

        _, err = tx.Exec(ctx, `
			INSERT INTO event (id, entry_id, event, payload)
			VALUES ($1, $2, $3, $4::jsonb)
		`, uuid.NewString(), parsedProductID, contract.EventProductAutoApproved, string(autoPayload))
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

func (s *Service) GetProduct(ctx context.Context, params contract.GetProduct) (contract.Product, error) {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return contract.Product{}, err
    }

    product := contract.Product{}
    var status string
    var approveReason, rejectionReason *string
    err = s.pool.QueryRow(ctx, `
        SELECT id::text, name, user_id::text, status, approve_reason, rejection_reason
        FROM product
        WHERE id = $1
    `, parsedProductID).Scan(&product.ID, &product.Name, &product.UserID, &status, &approveReason, &rejectionReason)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return contract.Product{}, contract.ErrProductNotFound
        }

        return contract.Product{}, err
    }

    product.Status = contract.ProductStatus(status)
    if approveReason != nil {
        product.ApproveReason = contract.ProductApproveReason(*approveReason)
    }
    if rejectionReason != nil {
        product.RejectionReason = contract.ProductRejectionReason(*rejectionReason)
    }

    return product, nil
}

func (s *Service) ProductHistory(ctx context.Context, params contract.ProductHistory) ([]contract.ProductHistoryRow, error) {
    parsedProductID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
    if err != nil {
        return nil, err
    }

    rows, err := s.pool.Query(ctx, `
        SELECT id::text, event, payload
        FROM event
        WHERE entry_id = $1
        ORDER BY created_at
    `, parsedProductID)
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

func parseUUID(value string, err error) (string, error) {
    parsed, parseErr := uuid.Parse(value)
    if parseErr != nil {
        return "", err
    }

    return parsed.String(), nil
}
