package products

import (
	"context"
	"errors"
	"strings"

	"architecture-bricks/contract"

	"github.com/google/uuid"
)

var _ contract.Application = (*Service)(nil)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ApproveProduct(ctx context.Context, params contract.ApproveProduct) error {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return err
	}

	_, err = parseUUID(params.ModeratorID, contract.ErrInvalidUserID)
	if err != nil {
		return err
	}

	product, version, err := s.repository.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	event := contract.ProductHistoryRow{
		ID:            uuid.NewString(),
		Event:         contract.EventProductApproved,
		Name:          product.Name,
		ModeratorID:   params.ModeratorID,
		ApproveReason: contract.ProductApproveReasonModerator,
	}

	return s.repository.ApproveProduct(ctx, productID, params.ModeratorID, version, event)
}

func (s *Service) RejectProduct(ctx context.Context, params contract.RejectProduct) error {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return err
	}

	_, err = parseUUID(params.ModeratorID, contract.ErrInvalidUserID)
	if err != nil {
		return err
	}

	product, version, err := s.repository.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	event := contract.ProductHistoryRow{
		ID:              uuid.NewString(),
		Event:           contract.EventProductRejected,
		Name:            product.Name,
		ModeratorID:     params.ModeratorID,
		RejectionReason: contract.ProductRejectionReasonModerator,
	}

	return s.repository.RejectProduct(ctx, productID, params.ModeratorID, version, event)
}

func (s *Service) CreateProduct(ctx context.Context, params contract.CreateProduct) error {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return err
	}

	_, err = parseUUID(params.UserID, contract.ErrInvalidUserID)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(params.Name)
	if name == "" {
		return contract.ErrProductNameRequired
	}

	autoApprove := isAutoApprovable(name)

	product := contract.Product{
		ID:     productID,
		Name:   name,
		UserID: params.UserID,
		Status: contract.ProductStatusPending,
	}

	eventList := []contract.ProductHistoryRow{
		{
			ID:    uuid.NewString(),
			Event: contract.EventProductCreated,
			Name:  name,
		},
	}

	if autoApprove {
		product.Status = contract.ProductStatusApproved
		product.ApproveReason = contract.ProductApproveReasonAuto

		eventList = append(eventList, contract.ProductHistoryRow{
			ID:            uuid.NewString(),
			Event:         contract.EventProductAutoApproved,
			Name:          name,
			ApproveReason: contract.ProductApproveReasonAuto,
		})
	}

	if err = s.repository.CreateProduct(ctx, product, eventList...); err != nil {
		if errors.Is(err, contract.ErrProductAlreadyExists) {
			return contract.ErrProductAlreadyExists
		}

		return err
	}

	return nil
}

func (s *Service) UpdateProduct(ctx context.Context, params contract.UpdateProduct) error {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return err
	}

	_, err = parseUUID(params.UserID, contract.ErrInvalidUserID)
	if err != nil {
		return err
	}

	existingProduct, version, err := s.repository.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	if existingProduct.Name == params.Name {
		return contract.ErrProductNameNotChanged
	}

	newName := strings.TrimSpace(params.Name)
	if newName == "" {
		return contract.ErrProductNameRequired
	}

	autoApprove := isAutoApprovable(newName)

	product := contract.Product{
		ID:   productID,
		Name: newName,
	}

	eventList := []contract.ProductHistoryRow{
		{
			ID:    uuid.NewString(),
			Event: contract.EventProductUpdated,
			Name:  newName,
		},
	}

	if autoApprove {
		product.Status = contract.ProductStatusApproved
		product.ApproveReason = contract.ProductApproveReasonAuto

		eventList = append(eventList, contract.ProductHistoryRow{
			ID:            uuid.NewString(),
			Event:         contract.EventProductAutoApproved,
			Name:          newName,
			ApproveReason: contract.ProductApproveReasonAuto,
		})
	}

	if err = s.repository.UpdateProduct(ctx, product, version, eventList...); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProduct(ctx context.Context, params contract.GetProduct) (contract.Product, error) {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return contract.Product{}, err
	}

	product, _, err := s.repository.GetProduct(ctx, productID)
	if err != nil {
		return contract.Product{}, err
	}

	return product, nil
}

func (s *Service) ProductHistory(ctx context.Context, params contract.ProductHistory) ([]contract.ProductHistoryRow, error) {
	productID, err := parseUUID(params.ProductID, contract.ErrInvalidProductID)
	if err != nil {
		return nil, err
	}

	return s.repository.ProductHistory(ctx, productID)
}

func isAutoApprovable(name string) bool {
	return name == "Кот" || name == "Собака"
}

func parseUUID(value string, err error) (string, error) {
	parsed, parseErr := uuid.Parse(value)
	if parseErr != nil {
		return "", err
	}

	return parsed.String(), nil
}
