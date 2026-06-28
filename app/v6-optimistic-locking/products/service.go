package products

import (
    "context"
    "errors"

    "architecture-bricks/app/v6-optimistic-locking/products/domain"
    "architecture-bricks/contract"
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
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    moderator, err := domain.NewModerator(params.ModeratorID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product := domain.NewProduct(productID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Approve(moderator)

    err = s.repo.Save(ctx, product)
    if errors.Is(err, contract.ErrProductAlreadyChanged) {
        return contract.ErrProductAlreadyChanged
    }

    return err
}

func (s *Service) RejectProduct(ctx context.Context, params contract.RejectProduct) error {
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    moderator, err := domain.NewModerator(params.ModeratorID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product := domain.NewProduct(productID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    product.Reject(moderator)

    err = s.repo.Save(ctx, product)
    if errors.Is(err, contract.ErrProductAlreadyChanged) {
        return contract.ErrProductAlreadyChanged
    }

    return err
}

func (s *Service) CreateProduct(ctx context.Context, params contract.CreateProduct) error {
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    user, err := domain.NewUser(params.UserID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product := domain.NewProduct(productID)

    if err = product.Create(params.Name, user); err != nil {
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

    if errors.Is(err, contract.ErrProductAlreadyChanged) {
        return contract.ErrProductAlreadyChanged
    }

    return err
}

func (s *Service) UpdateProduct(ctx context.Context, params contract.UpdateProduct) error {
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return contract.ErrInvalidProductID
    }

    user, err := domain.NewUser(params.UserID)
    if err != nil {
        return contract.ErrInvalidUserID
    }

    product := domain.NewProduct(productID)

    if err = s.repo.Load(ctx, product); err != nil {
        if errors.Is(err, domain.ErrProductNotFound) {
            return contract.ErrProductNotFound
        }

        return err
    }

    if err = product.Rename(params.Name, user); err != nil {
        if errors.Is(err, domain.ErrProductNameNotChanged) {
            return contract.ErrProductNameNotChanged
        }

        if errors.Is(err, domain.ErrProductNameRequired) {
            return contract.ErrProductNameRequired
        }

        return err
    }

    err = s.repo.Save(ctx, product)
    if errors.Is(err, contract.ErrProductAlreadyChanged) {
        return contract.ErrProductAlreadyChanged
    }

    return err
}

func (s *Service) GetProduct(ctx context.Context, params contract.GetProduct) (contract.Product, error) {
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return contract.Product{}, contract.ErrInvalidProductID
    }

    product := domain.NewProduct(productID)

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
    productID, err := domain.NewProductID(params.ProductID)
    if err != nil {
        return nil, contract.ErrInvalidProductID
    }

    history, err := s.repo.GetHistory(ctx, productID.Value())
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
