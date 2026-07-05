package domain

import "github.com/google/uuid"

// ProductID — VO идентификатора продукта. Инкапсулирует строковое представление.
type ProductID struct {
	value string
}

func NewProductID(id string) (ProductID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return ProductID{}, ErrInvalidProductID
	}

	return ProductID{value: id}, nil
}

func (id ProductID) Value() string {
	return id.value
}
