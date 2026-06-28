package domain

import ddd "architecture-bricks/pkg/domain-events/domain"

var (
    _ ddd.Event = ProductCreated{}
    _ ddd.Event = ProductRenamed{}
    _ ddd.Event = ProductApproved{}
    _ ddd.Event = ProductRejected{}
    _ ddd.Event = ProductAutoApproved{}
)

type ProductCreated struct {
    UserID string
    Name   string
}

type ProductRenamed struct {
    UserID  string
    OldName string
    NewName string
}

type ProductApproved struct {
    ModeratorID   string
    ApproveReason ProductApproveReason
}

type ProductRejected struct {
    ModeratorID     string
    RejectionReason ProductRejectionReason
}

type ProductAutoApproved struct {
    ApproveReason ProductApproveReason
}
