package domain

import "errors"

var (
	ErrProductNameRequired   = errors.New("product name is required")
	ErrProductAlreadyExists  = errors.New("product already exists")
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidProductID      = errors.New("invalid product id")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrProductNameNotChanged = errors.New("product name has not changed")
)
