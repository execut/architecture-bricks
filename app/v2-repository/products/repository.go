package products

import (
	"context"

	"architecture-bricks/contract"
)

type Repository interface {
	CreateProduct(ctx context.Context, product contract.Product, events ...contract.ProductHistoryRow) error
	UpdateProduct(ctx context.Context, product contract.Product, version int, events ...contract.ProductHistoryRow) error
	ApproveProduct(ctx context.Context, productID string, moderatorID string, version int, event contract.ProductHistoryRow) error
	RejectProduct(ctx context.Context, productID string, moderatorID string, version int, event contract.ProductHistoryRow) error
	GetProduct(ctx context.Context, productID string) (contract.Product, int, error)
	ProductHistory(ctx context.Context, productID string) ([]contract.ProductHistoryRow, error)
}
