package domain

import (
    biz "architecture-bricks/pkg/optimistic-locking/business-events/value-objects/domain"
    vo "architecture-bricks/pkg/optimistic-locking/value-objects/domain"
)

var _ biz.Entity = (*Product)(nil)

type Product struct {
    id              ProductID
    user            User
    name            ProductName
    status          ProductStatus
    approveReason   ProductApproveReason
    rejectionReason ProductRejectionReason
    version         vo.Version
    events          []biz.Event
}

func NewProduct(id ProductID) *Product {
    return &Product{id: id, version: vo.NewInitialVersion()}
}

func (p *Product) Version() vo.Version {
    return p.version
}

func (p *Product) Create(name string, user User) error {
    if p.name.Value() != "" {
        return ErrProductAlreadyExists
    }

    productName, err := NewProductName(name)
    if err != nil {
        return err
    }

    p.AddAndApplyEvent(biz.NewEvent(p.version.Next(), ProductCreatedPayload{
        Name: productName.Value(),
        User: user.Value(),
    }))

    p.AutoApproveIfEligible()

    return nil
}

func (p *Product) Rename(name string, user User) error {
    newName, err := NewProductName(name)
    if err != nil {
        return err
    }

    if p.name.Value() == newName.Value() {
        return ErrProductNameNotChanged
    }

    p.AddAndApplyEvent(biz.NewEvent(p.version.Next(), ProductRenamedPayload{
        OldName: p.name.Value(),
        NewName: newName.Value(),
        User:    user.Value(),
    }))

    p.AutoApproveIfEligible()

    return nil
}

func (p *Product) Approve(moderator Moderator) {
    p.AddAndApplyEvent(biz.NewEvent(p.version.Next(), ProductApprovedPayload{
        ModeratorID:   moderator.Value(),
        ApproveReason: string(ProductApproveReasonModerator),
    }))
}

func (p *Product) Reject(moderator Moderator) {
    p.AddAndApplyEvent(biz.NewEvent(p.version.Next(), ProductRejectedPayload{
        ModeratorID:     moderator.Value(),
        RejectionReason: string(ProductRejectionReasonModerator),
    }))
}

func (p *Product) AutoApproveIfEligible() bool {
    if p.name.Value() != "Кот" && p.name.Value() != "Собака" {
        return false
    }

    p.AddAndApplyEvent(biz.NewEvent(p.version.Next(), ProductAutoApprovedPayload{
        ApproveReason: string(ProductApproveReasonAuto),
    }))

    return true
}

func (p *Product) EventList() []biz.Event {
    return p.events
}

func (p *Product) AddAndApplyEvent(event biz.Event) {
    p.events = append(p.events, event)
    p.version = event.Version()

    switch payload := event.Payload().(type) {
    case ProductCreatedPayload:
        name, _ := NewProductName(payload.Name)
        p.name = name
        p.user, _ = NewUser(payload.User)
        p.status = ProductStatusPending
        p.approveReason = ""
        p.rejectionReason = ""
    case ProductRenamedPayload:
        name, _ := NewProductName(payload.NewName)
        p.name = name
    case ProductApprovedPayload:
        p.status = ProductStatusApproved
        p.approveReason = ProductApproveReason(payload.ApproveReason)
        p.rejectionReason = ""
    case ProductRejectedPayload:
        p.status = ProductStatusRejected
        p.rejectionReason = ProductRejectionReason(payload.RejectionReason)
        p.approveReason = ""
    case ProductAutoApprovedPayload:
        p.status = ProductStatusApproved
        p.approveReason = ProductApproveReason(payload.ApproveReason)
        p.rejectionReason = ""
    }
}

func (p *Product) CleanEventList() {
    p.events = nil
}

func (p *Product) ID() string {
    return p.id.Value()
}

func (p *Product) UserID() string {
    return p.user.Value()
}

func (p *Product) Name() string {
    return p.name.Value()
}

func (p *Product) Status() ProductStatus {
    return p.status
}

func (p *Product) ApproveReason() ProductApproveReason {
    return p.approveReason
}

func (p *Product) RejectionReason() ProductRejectionReason {
    return p.rejectionReason
}
