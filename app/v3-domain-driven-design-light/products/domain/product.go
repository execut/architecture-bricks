package domain

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
	version         int
}

func NewProduct(id string, userID string, name string) (*Product, error) {
	if name == "" {
		return nil, ErrProductNameRequired
	}

	return &Product{id: id, userID: userID, name: name, status: ProductStatusPending}, nil
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

func (p *Product) Version() int {
	return p.version
}

func (p *Product) Rename(name string) error {
	if p.name == name {
		return ErrProductNameNotChanged
	}

	if name == "" {
		return ErrProductNameRequired
	}

	p.name = name

	return nil
}

func (p *Product) Approve(moderatorID string) {
	p.status = ProductStatusApproved
	p.approveReason = ProductApproveReasonModerator
	p.rejectionReason = ""
}

func (p *Product) Reject(moderatorID string) {
	p.status = ProductStatusRejected
	p.rejectionReason = ProductRejectionReasonModerator
	p.approveReason = ""
}

func (p *Product) AutoApprove() {
	p.status = ProductStatusApproved
	p.approveReason = ProductApproveReasonAuto
	p.rejectionReason = ""
}

// AutoApproveIfEligible проверяет, подлежит ли продукт авто-одобрению,
// и если да — применяет его. Возвращает true, если авто-одобрение было применено.
func (p *Product) AutoApproveIfEligible() bool {
	if p.name != "Кот" && p.name != "Собака" {
		return false
	}

	p.AutoApprove()

	return true
}

// LoadProduct creates a Product from persisted data without validation.
// Used by repository to reconstruct domain objects from DB rows.
func LoadProduct(
	id string,
	userID string,
	name string,
	status ProductStatus,
	approveReason ProductApproveReason,
	rejectionReason ProductRejectionReason,
	version int,
) *Product {
	return &Product{
		id:              id,
		userID:          userID,
		name:            name,
		status:          status,
		approveReason:   approveReason,
		rejectionReason: rejectionReason,
		version:         version,
	}
}
