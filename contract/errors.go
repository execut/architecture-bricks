package contract

import "errors"

var (
    ErrInvalidProductID      = errors.New("invalid product id")
    ErrInvalidUserID         = errors.New("invalid user id")
    ErrProductNameRequired   = errors.New("product name is required")
    ErrProductAlreadyExists  = errors.New("product already exists")
    ErrProductNotFound       = errors.New("product not found")
    ErrProductNameNotChanged = errors.New("product name has not changed")
    ErrProductAlreadyChanged = errors.New("product already changed by other request")
)
