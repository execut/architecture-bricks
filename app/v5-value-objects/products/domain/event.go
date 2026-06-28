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
    User User
    Name ProductName
}

type ProductRenamed struct {
    User    User
    OldName ProductName
    NewName ProductName
}

type ProductApproved struct {
    Moderator     Moderator
    ApproveReason ProductApproveReason
}

type ProductRejected struct {
    Moderator       Moderator
    RejectionReason ProductRejectionReason
}

type ProductAutoApproved struct {
    ApproveReason ProductApproveReason
}
