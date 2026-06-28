package domain

import ddd "architecture-bricks/pkg/domain-events/domain"

var _ ddd.Entity = (*Product)(nil)

type Product struct {
    id              ProductID
    user            User
    name            ProductName
    status          ProductStatus
    approveReason   ProductApproveReason
    rejectionReason ProductRejectionReason
    events          []ddd.Event
}

func NewProduct(id ProductID) *Product {
    return &Product{id: id}
}

func (p *Product) Create(name string, user User) error {
    if p.name.Value() != "" {
        return ErrProductAlreadyExists
    }

    productName, err := NewProductName(name)
    if err != nil {
        return err
    }

    p.AddAndApplyEvent(ProductCreated{User: user, Name: productName})
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

    p.AddAndApplyEvent(ProductRenamed{
        User:    user,
        OldName: p.name,
        NewName: newName,
    })

    p.AutoApproveIfEligible()

    return nil
}

func (p *Product) Approve(moderator Moderator) {
    p.AddAndApplyEvent(ProductApproved{
        Moderator:     moderator,
        ApproveReason: ProductApproveReasonModerator,
    })
}

func (p *Product) Reject(moderator Moderator) {
    p.AddAndApplyEvent(ProductRejected{
        Moderator:       moderator,
        RejectionReason: ProductRejectionReasonModerator,
    })
}

// AutoApproveIfEligible проверяет, подлежит ли продукт авто-одобрению,
// и если да — генерирует событие ProductAutoApproved.
func (p *Product) AutoApproveIfEligible() bool {
    if p.name.Value() != "Кот" && p.name.Value() != "Собака" {
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
        p.user = e.User
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
