package domain

import "context"

type ProductHistoryRow struct {
	ID              string
	Event           string
	Name            string
	ModeratorID     string
	ApproveReason   string
	RejectionReason string
}

type Repository interface {
	Create(ctx context.Context, product *Product, events ...ProductHistoryRow) error
	Update(ctx context.Context, product *Product, events ...ProductHistoryRow) error
	GetByID(ctx context.Context, id string) (*Product, error)
	GetHistory(ctx context.Context, id string) ([]ProductHistoryRow, error)
}
