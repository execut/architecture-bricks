package products

import (
    "context"
    "errors"
    "strings"

    "architecture-bricks/app/v4-domain-driven-design-light-with-events/products/domain"
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

    product := domain.NewProduct(params.ProductID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Approve(params.ModeratorID)

    return s.repo.Save(ctx, product)
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

    product := domain.NewProduct(params.ProductID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Reject(params.ModeratorID)

    return s.repo.Save(ctx, product)
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

    product := domain.NewProduct(params.ProductID)

    name := strings.TrimSpace(params.Name)

    if err = product.Create(name, params.UserID); err != nil {
        if errors.Is(err, domain.ErrProductNameRequired) {
            return contract.ErrProductNameRequired
        }

        if errors.Is(err, domain.ErrProductAlreadyExists) {
            return contract.ErrProductAlreadyExists
        }

        return err
    }

    err = s.repo.Save(ctx, product)
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

    product := domain.NewProduct(params.ProductID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    if err = product.Rename(strings.TrimSpace(params.Name), params.UserID); err != nil {
        if errors.Is(err, domain.ErrProductNameNotChanged) {
            return contract.ErrProductNameNotChanged
        }

        if errors.Is(err, domain.ErrProductNameRequired) {
            return contract.ErrProductNameRequired
        }

        return err
    }

    return s.repo.Save(ctx, product)
}

func (s *Service) GetProduct(ctx context.Context, params contract.GetProduct) (contract.Product, error) {
    _, err := uuid.Parse(params.ProductID)
    if err != nil {
        return contract.Product{}, contract.ErrInvalidProductID
    }

    product := domain.NewProduct(params.ProductID)

    if err = s.repo.Load(ctx, product); err != nil {
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
