package domain

import ddd "architecture-bricks/pkg/domain-events/domain"

var _ ddd.Entity = (*Product)(nil)

type ProductStatus string

const (
    ProductStatusPending  ProductStatus = "pending"
    ProductStatusApproved ProductStatus = "approved"
    ProductStatusRejected ProductStatus = "rejected"
)

type ProductApproveReason string

const (
    ProductApproveReasonModerator ProductApproveReason = "Moderator"
    ProductApproveReasonAuto      ProductApproveReason = "Auto"
)

type ProductRejectionReason string

const (
    ProductRejectionReasonModerator ProductRejectionReason = "Moderator"
)

type Product struct {
    id              string
    userID          string
    name            string
    status          ProductStatus
    approveReason   ProductApproveReason
    rejectionReason ProductRejectionReason
    events          []ddd.Event
}

func NewProduct(id string) *Product {
    return &Product{id: id}
}

func (p *Product) Create(name string, userID string) error {
    if p.name != "" {
        return ErrProductAlreadyExists
    }

    if name == "" {
        return ErrProductNameRequired
    }

    p.AddAndApplyEvent(ProductCreated{UserID: userID, Name: name})
    p.AutoApproveIfEligible()

    return nil
}

func (p *Product) Rename(name string, userID string) error {
    if p.name == name {
        return ErrProductNameNotChanged
    }

    if name == "" {
        return ErrProductNameRequired
    }

    p.AddAndApplyEvent(ProductRenamed{UserID: userID, OldName: p.name, NewName: name})
    p.AutoApproveIfEligible()

    return nil
}

func (p *Product) Approve(moderatorID string) {
    p.AddAndApplyEvent(ProductApproved{
        ModeratorID:   moderatorID,
        ApproveReason: ProductApproveReasonModerator,
    })
}

func (p *Product) Reject(moderatorID string) {
    p.AddAndApplyEvent(ProductRejected{
        ModeratorID:     moderatorID,
        RejectionReason: ProductRejectionReasonModerator,
    })
}

// AutoApproveIfEligible проверяет, подлежит ли продукт авто-одобрению,
// и если да — генерирует событие ProductAutoApproved.
// Возвращает true, если авто-одобрение было применено.
func (p *Product) AutoApproveIfEligible() bool {
    if p.name != "Кот" && p.name != "Собака" {
        return false
    }

    p.AddAndApplyEvent(ProductAutoApproved{
        ApproveReason: ProductApproveReasonAuto,
    })

    return true
}

func (p *Product) EventList() []ddd.Event {
    return p.events
}

func (p *Product) AddAndApplyEvent(event ddd.Event) {
    p.events = append(p.events, event)

    switch e := event.(type) {
    case ProductCreated:
        p.name = e.Name
        p.userID = e.UserID
        p.status = ProductStatusPending
        p.approveReason = ""
        p.rejectionReason = ""
    case ProductRenamed:
        p.name = e.NewName
    case ProductApproved:
        p.status = ProductStatusApproved
        p.approveReason = e.ApproveReason
        p.rejectionReason = ""
    case ProductRejected:
        p.status = ProductStatusRejected
        p.rejectionReason = e.RejectionReason
        p.approveReason = ""
    case ProductAutoApproved:
        p.status = ProductStatusApproved
        p.approveReason = e.ApproveReason
        p.rejectionReason = ""
    }
}

func (p *Product) CleanEventList() {
    p.events = nil
}

func (p *Product) ID() string {
    return p.id
}

func (p *Product) UserID() string {
    return p.userID
}

func (p *Product) Name() string {
    return p.name
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
