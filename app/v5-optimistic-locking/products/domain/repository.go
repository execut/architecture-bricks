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
	Load(ctx context.Context, product *Product) error
	Save(ctx context.Context, product *Product) error
	GetHistory(ctx context.Context, id string) ([]ProductHistoryRow, error)
}
