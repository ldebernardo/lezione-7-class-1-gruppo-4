package entities

import "regexp"

var skuPattern = regexp.MustCompile(`^[A-Z0-9-]{3,32}$`)

type SKU struct {
	code string
}

func NewSKU(code string) (SKU, error) {
	if !skuPattern.MatchString(code) {
		return SKU{}, ErrInvalidSKU
	}
	return SKU{code: code}, nil
}

func (s SKU) Code() string {
	return s.code
}

func (s SKU) Equal(other SKU) bool {
	return s.code == other.code
}
