package domain

// ProductApproveReason — причина одобрения продукта.
type ProductApproveReason string

const (
	ProductApproveReasonModerator ProductApproveReason = "Moderator"
	ProductApproveReasonAuto      ProductApproveReason = "Auto"
)
