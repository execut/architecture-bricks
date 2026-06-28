package domain

import "strings"

// ProductName — VO имени продукта. Инкапсулирует имя с триммингом и валидацией.
type ProductName struct {
	value string
}

func NewProductName(name string) (ProductName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ProductName{}, ErrProductNameRequired
	}

	return ProductName{value: name}, nil
}

func (n ProductName) Value() string {
	return n.value
}
