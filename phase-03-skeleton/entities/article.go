package entities

import (
	"strings"
	"sync"
	"time"

	"warehouse.local/core/events"
)

type Article struct {
	mu          sync.Mutex
	id          string
	sku         SKU
	name        string
	description string
	price       Money
	inventory   map[string]*InventoryLevel
	pending     []events.DomainEvent
}

func NewArticle(id string, sku SKU, name, description string, price Money, occurredAt time.Time) (*Article, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidArticleID
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}
	if price.AmountCents() <= 0 {
		return nil, ErrInvalidPrice
	}
	article := &Article{
		id:          id,
		sku:         sku,
		name:        strings.TrimSpace(name),
		description: description,
		price:       price,
		inventory:   make(map[string]*InventoryLevel),
	}
	article.record(events.NewArticleCreated(id, sku.Code(), article.name, price.AmountCents(), price.Currency(), occurredAt))
	return article, nil
}

func (a *Article) ID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.id
}

func (a *Article) SKU() SKU {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sku
}

func (a *Article) Name() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.name
}

func (a *Article) Description() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.description
}

func (a *Article) Price() Money {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.price
}

func (a *Article) ChangeName(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	a.name = strings.TrimSpace(name)
	return nil
}

func (a *Article) ChangeDescription(description string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.description = description
}

func (a *Article) ChangePrice(price Money, occurredAt time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if price.AmountCents() <= 0 {
		return ErrInvalidPrice
	}
	if price.Currency() != a.price.Currency() {
		return ErrCurrencyChange
	}
	if price.AmountCents() == a.price.AmountCents() {
		return nil
	}
	old := a.price
	a.price = price
	a.record(events.NewArticlePriceChanged(a.id, old.AmountCents(), price.AmountCents(), price.Currency(), occurredAt))
	return nil
}

func (a *Article) AddInventoryLevel(id string, location WarehouseLocation, quantity, reserved int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.inventory[location.Code()]; exists {
		return nil
	}
	level, err := newInventoryLevel(id, location, quantity, reserved)
	if err != nil {
		return err
	}
	a.inventory[location.Code()] = level
	return nil
}

func (a *Article) AdjustInventory(location WarehouseLocation, delta int, reason string, occurredAt time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	level, err := a.inventoryLevelFor(location)
	if err != nil {
		return err
	}
	if err := level.adjust(delta); err != nil {
		return err
	}
	a.record(events.NewInventoryAdjusted(a.id, location.Code(), delta, level.quantity, reason, occurredAt))
	return nil
}

func (a *Article) ReserveStock(location WarehouseLocation, quantity int, reservationID, orderID string, occurredAt time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(reservationID) == "" || strings.TrimSpace(orderID) == "" {
		return ErrInvalidReservation
	}
	level, err := a.inventoryLevelFor(location)
	if err != nil {
		return err
	}
	if err := level.reserve(quantity); err != nil {
		return err
	}
	a.record(events.NewStockReserved(a.id, location.Code(), quantity, reservationID, orderID, occurredAt))
	return nil
}

func (a *Article) ReleaseStock(location WarehouseLocation, quantity int, reservationID, reason string, occurredAt time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(reservationID) == "" {
		return ErrInvalidReservation
	}
	level, err := a.inventoryLevelFor(location)
	if err != nil {
		return err
	}
	if err := level.release(quantity); err != nil {
		return err
	}
	a.record(events.NewStockReservationReleased(a.id, location.Code(), quantity, reservationID, reason, occurredAt))
	return nil
}

func (a *Article) InventoryLevel(location WarehouseLocation) (InventoryLevel, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	level, ok := a.inventory[location.Code()]
	if !ok {
		return InventoryLevel{}, false
	}
	return *level, true
}

func (a *Article) InventoryLevels() []InventoryLevel {
	a.mu.Lock()
	defer a.mu.Unlock()
	levels := make([]InventoryLevel, 0, len(a.inventory))
	for _, level := range a.inventory {
		levels = append(levels, *level)
	}
	return levels
}

func (a *Article) PendingEvents() []events.DomainEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]events.DomainEvent(nil), a.pending...)
}

func (a *Article) PullEvents() []events.DomainEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	pending := append([]events.DomainEvent(nil), a.pending...)
	a.pending = nil
	return pending
}

func (a *Article) inventoryLevelFor(location WarehouseLocation) (*InventoryLevel, error) {
	level, ok := a.inventory[location.Code()]
	if !ok {
		return nil, ErrInventoryLevelNotFound
	}
	return level, nil
}

func (a *Article) record(event events.DomainEvent) {
	a.pending = append(a.pending, event)
}
