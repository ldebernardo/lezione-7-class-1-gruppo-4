package entities

import "regexp"

var locationPattern = regexp.MustCompile(`^[A-Z]{2}-[A-Z0-9]{3,8}$`)

type WarehouseLocation struct {
	code string
}

func NewWarehouseLocation(code string) (WarehouseLocation, error) {
	if !locationPattern.MatchString(code) {
		return WarehouseLocation{}, ErrInvalidLocation
	}
	return WarehouseLocation{code: code}, nil
}

func (l WarehouseLocation) Code() string {
	return l.code
}

func (l WarehouseLocation) Equal(other WarehouseLocation) bool {
	return l.code == other.code
}
