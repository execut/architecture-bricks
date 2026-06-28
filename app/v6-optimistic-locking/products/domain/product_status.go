package domain

// ProductStatus — статус продукта.
type ProductStatus string

const (
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusApproved ProductStatus = "approved"
	ProductStatusRejected ProductStatus = "rejected"
)
