package entities

import "errors"

var (
	ErrInvalidArticleID       = errors.New("article id must not be empty")
	ErrInvalidName            = errors.New("article name must not be empty")
	ErrInvalidPrice           = errors.New("article price must be greater than zero")
	ErrCurrencyChange         = errors.New("article price currency cannot change after creation")
	ErrInvalidSKU             = errors.New("sku must match ^[A-Z0-9-]{3,32}$")
	ErrInvalidMoney           = errors.New("money must have non-negative cents and a three-letter uppercase currency")
	ErrInvalidInventoryID     = errors.New("inventory level id must not be empty")
	ErrInvalidLocation        = errors.New("warehouse location must match ^[A-Z]{2}-[A-Z0-9]{3,8}$")
	ErrInvalidQuantity        = errors.New("quantity must not be negative")
	ErrInvalidReserved        = errors.New("reserved quantity must not be negative")
	ErrReservedExceedsStock   = errors.New("reserved quantity cannot exceed quantity")
	ErrInventoryLevelNotFound = errors.New("inventory level not found")
	ErrInvalidReservation     = errors.New("reservation id and order id must not be empty")
)
