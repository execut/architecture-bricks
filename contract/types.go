package contract

type ProductEvent string

type ProductApproveReason string

type ProductRejectionReason string

type ProductStatus string

const (
	EventProductCreated      ProductEvent = "product_created"
	EventProductUpdated      ProductEvent = "product_updated"
	EventProductApproved     ProductEvent = "product_approved"
	EventProductRejected     ProductEvent = "product_rejected"
	EventProductAutoApproved ProductEvent = "product_auto_approved"
)

const (
	ProductApproveReasonModerator ProductApproveReason = "Moderator"
	ProductApproveReasonAuto      ProductApproveReason = "Auto"
)

const (
	ProductRejectionReasonModerator ProductRejectionReason = "Moderator"
)

const (
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusApproved ProductStatus = "approved"
	ProductStatusRejected ProductStatus = "rejected"
)

type CreateProduct struct {
	ProductID string
	UserID    string
	Name      string
}

type UpdateProduct struct {
	ProductID string
	UserID    string
	Name      string
}

type ApproveProduct struct {
	ProductID   string
	ModeratorID string
}

type RejectProduct struct {
	ProductID   string
	ModeratorID string
}

type GetProduct struct {
	ProductID string
}

type ProductHistory struct {
	ProductID string
}

type Product struct {
	ID              string
	Name            string
	UserID          string
	Status          ProductStatus
	ApproveReason   ProductApproveReason
	RejectionReason ProductRejectionReason
}

type ProductHistoryRow struct {
	ID              string
	Event           ProductEvent
	ModeratorID     string
	Name            string
	ApproveReason   ProductApproveReason
	RejectionReason ProductRejectionReason
}
