package contract

import "context"

type Application interface {
	CreateProduct(ctx context.Context, params CreateProduct) error
	UpdateProduct(ctx context.Context, params UpdateProduct) error
	ApproveProduct(ctx context.Context, params ApproveProduct) error
	RejectProduct(ctx context.Context, params RejectProduct) error
	GetProduct(ctx context.Context, params GetProduct) (Product, error)
	ProductHistory(ctx context.Context, params ProductHistory) ([]ProductHistoryRow, error)
}
