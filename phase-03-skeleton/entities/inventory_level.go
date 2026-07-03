package entities

type InventoryLevel struct {
	id       string
	location WarehouseLocation
	quantity int
	reserved int
}

func newInventoryLevel(id string, location WarehouseLocation, quantity, reserved int) (*InventoryLevel, error) {
	if id == "" {
		return nil, ErrInvalidInventoryID
	}
	if quantity < 0 {
		return nil, ErrInvalidQuantity
	}
	if reserved < 0 {
		return nil, ErrInvalidReserved
	}
	if reserved > quantity {
		return nil, ErrReservedExceedsStock
	}
	return &InventoryLevel{
		id:       id,
		location: location,
		quantity: quantity,
		reserved: reserved,
	}, nil
}

func (l InventoryLevel) ID() string {
	return l.id
}

func (l InventoryLevel) Location() WarehouseLocation {
	return l.location
}

func (l InventoryLevel) Quantity() int {
	return l.quantity
}

func (l InventoryLevel) Reserved() int {
	return l.reserved
}

func (l InventoryLevel) Available() int {
	return l.quantity - l.reserved
}

func (l *InventoryLevel) adjust(delta int) error {
	next := l.quantity + delta
	if next < 0 {
		return ErrInvalidQuantity
	}
	if l.reserved > next {
		return ErrReservedExceedsStock
	}
	l.quantity = next
	return nil
}

func (l *InventoryLevel) reserve(quantity int) error {
	if quantity < 0 {
		return ErrInvalidReserved
	}
	next := l.reserved + quantity
	if next > l.quantity {
		return ErrReservedExceedsStock
	}
	l.reserved = next
	return nil
}

func (l *InventoryLevel) release(quantity int) error {
	if quantity < 0 {
		return ErrInvalidReserved
	}
	next := l.reserved - quantity
	if next < 0 {
		return ErrInvalidReserved
	}
	l.reserved = next
	return nil
}
