package products

import (
    "context"
    "errors"
    "strings"

    "architecture-bricks/app/v3-domain-driven-design-light/products/domain"
    "architecture-bricks/contract"

    "github.com/google/uuid"
)

var _ contract.Application = (*Service)(nil)

type Service struct {
    repo domain.Repository
}

func NewService(ctx context.Context) (*Service, error) {
    repo, err := NewRepository(ctx)
    if err != nil {
        return nil, err
    }

    return &Service{repo: repo}, nil
}

func (s *Service) ApproveProduct(ctx context.Context, params contract.ApproveProduct) error {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    _, err = uuid.Parse(params.ModeratorID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product, err := s.repo.GetByID(ctx, params.ProductID)
    if err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Approve(params.ModeratorID)

    event := domain.ProductHistoryRow{
        ID:            uuid.NewString(),
        Event:         string(contract.EventProductApproved),
        Name:          product.Name(),
        ModeratorID:   params.ModeratorID,
        ApproveReason: string(domain.ProductApproveReasonModerator),
    }

    return s.repo.Update(ctx, product, event)
}

func (s *Service) RejectProduct(ctx context.Context, params contract.RejectProduct) error {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    _, err = uuid.Parse(params.ModeratorID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product, err := s.repo.GetByID(ctx, params.ProductID)
    if err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Reject(params.ModeratorID)

    event := domain.ProductHistoryRow{
        ID:              uuid.NewString(),
        Event:           string(contract.EventProductRejected),
        Name:            product.Name(),
        ModeratorID:     params.ModeratorID,
        RejectionReason: string(domain.ProductRejectionReasonModerator),
    }

    return s.repo.Update(ctx, product, event)
}

func (s *Service) CreateProduct(ctx context.Context, params contract.CreateProduct) error {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    _, err = uuid.Parse(params.UserID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    name := strings.TrimSpace(params.Name)

    product, err := domain.NewProduct(params.ProductID, params.UserID, name)
    if err != nil {
        if errors.Is(err, domain.ErrProductNameRequired) {
            return contract.ErrProductNameRequired
        }

        return err
    }

    events := make([]domain.ProductHistoryRow, 0, 2)

    events = append(events, domain.ProductHistoryRow{
        ID:    uuid.NewString(),
        Event: string(contract.EventProductCreated),
        Name:  product.Name(),
    })

    if product.AutoApproveIfEligible() {
        events = append(events, domain.ProductHistoryRow{
            ID:            uuid.NewString(),
            Event:         string(contract.EventProductAutoApproved),
            Name:          product.Name(),
            ApproveReason: string(domain.ProductApproveReasonAuto),
        })
    }

    err = s.repo.Create(ctx, product, events...)
    if errors.Is(err, domain.ErrProductAlreadyExists) {
        return contract.ErrProductAlreadyExists
    }

    return err
}

func (s *Service) UpdateProduct(ctx context.Context, params contract.UpdateProduct) error {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    _, err = uuid.Parse(params.UserID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product, err := s.repo.GetByID(ctx, params.ProductID)
    if err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    err = product.Rename(strings.TrimSpace(params.Name))
    if err != nil {
        if errors.Is(err, domain.ErrProductNameNotChanged) {
            return contract.ErrProductNameNotChanged
        }

        if errors.Is(err, domain.ErrProductNameRequired) {
            return contract.ErrProductNameRequired
        }

        return err
    }

    events := make([]domain.ProductHistoryRow, 0, 2)

    events = append(events, domain.ProductHistoryRow{
        ID:    uuid.NewString(),
        Event: string(contract.EventProductUpdated),
        Name:  product.Name(),
    })

    if product.AutoApproveIfEligible() {
        events = append(events, domain.ProductHistoryRow{
            ID:            uuid.NewString(),
            Event:         string(contract.EventProductAutoApproved),
            Name:          product.Name(),
            ApproveReason: string(domain.ProductApproveReasonAuto),
        })
    }

    return s.repo.Update(ctx, product, events...)
}

func (s *Service) GetProduct(ctx context.Context, params contract.GetProduct) (contract.Product, error) {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.Product{}, contract.ErrInvalidProductID
    }

    product, err := s.repo.GetByID(ctx, params.ProductID)
    if err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.Product{}, contract.ErrProductNotFound
        }

        return contract.Product{}, err
    }

    return contract.Product{
        ID:              product.ID(),
        Name:            product.Name(),
        UserID:          product.UserID(),
        Status:          contract.ProductStatus(product.Status()),
        ApproveReason:   contract.ProductApproveReason(product.ApproveReason()),
        RejectionReason: contract.ProductRejectionReason(product.RejectionReason()),
    }, nil
}

func (s *Service) ProductHistory(
    ctx context.Context,
    params contract.ProductHistory,
) ([]contract.ProductHistoryRow, error) {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return nil, contract.ErrInvalidProductID
    }

    history, err := s.repo.GetHistory(ctx, params.ProductID)
    if err != nil {
        return nil, err
    }

    result := make([]contract.ProductHistoryRow, len(history))
    for i, row := range history {
        result[i] = contract.ProductHistoryRow{
            ID:              row.ID,
            Event:           contract.ProductEvent(row.Event),
            Name:            row.Name,
            ModeratorID:     row.ModeratorID,
            ApproveReason:   contract.ProductApproveReason(row.ApproveReason),
            RejectionReason: contract.ProductRejectionReason(row.RejectionReason),
        }
    }

    return result, nil
}
