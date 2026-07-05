package domain

type ProductCreatedPayload struct {
	Name string
	User string
}

type ProductRenamedPayload struct {
	OldName string
	NewName string
	User    string
}

type ProductApprovedPayload struct {
	ModeratorID   string
	ApproveReason string
}

type ProductRejectedPayload struct {
	ModeratorID     string
	RejectionReason string
}

type ProductAutoApprovedPayload struct {
	ApproveReason string
}
