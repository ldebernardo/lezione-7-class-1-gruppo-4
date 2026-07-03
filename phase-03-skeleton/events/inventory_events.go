package events

import "time"

type InventoryAdjusted struct {
	baseEvent
	articleID    string
	locationCode string
	delta        int
	newQuantity  int
	reason       string
}

func NewInventoryAdjusted(articleID, locationCode string, delta, newQuantity int, reason string, at time.Time) InventoryAdjusted {
	return InventoryAdjusted{
		baseEvent:    baseEvent{occurredAt: occurredAt(at)},
		articleID:    articleID,
		locationCode: locationCode,
		delta:        delta,
		newQuantity:  newQuantity,
		reason:       reason,
	}
}

func (InventoryAdjusted) EventName() string      { return "InventoryAdjusted" }
func (e InventoryAdjusted) ArticleID() string    { return e.articleID }
func (e InventoryAdjusted) LocationCode() string { return e.locationCode }
func (e InventoryAdjusted) Delta() int           { return e.delta }
func (e InventoryAdjusted) NewQuantity() int     { return e.newQuantity }
func (e InventoryAdjusted) Reason() string       { return e.reason }

type StockReserved struct {
	baseEvent
	articleID     string
	locationCode  string
	quantity      int
	reservationID string
	orderID       string
}

func NewStockReserved(articleID, locationCode string, quantity int, reservationID, orderID string, at time.Time) StockReserved {
	return StockReserved{
		baseEvent:     baseEvent{occurredAt: occurredAt(at)},
		articleID:     articleID,
		locationCode:  locationCode,
		quantity:      quantity,
		reservationID: reservationID,
		orderID:       orderID,
	}
}

func (StockReserved) EventName() string       { return "StockReserved" }
func (e StockReserved) ArticleID() string     { return e.articleID }
func (e StockReserved) LocationCode() string  { return e.locationCode }
func (e StockReserved) Quantity() int         { return e.quantity }
func (e StockReserved) ReservationID() string { return e.reservationID }
func (e StockReserved) OrderID() string       { return e.orderID }

type StockReservationReleased struct {
	baseEvent
	articleID     string
	locationCode  string
	quantity      int
	reservationID string
	reason        string
}

func NewStockReservationReleased(articleID, locationCode string, quantity int, reservationID, reason string, at time.Time) StockReservationReleased {
	return StockReservationReleased{
		baseEvent:     baseEvent{occurredAt: occurredAt(at)},
		articleID:     articleID,
		locationCode:  locationCode,
		quantity:      quantity,
		reservationID: reservationID,
		reason:        reason,
	}
}

func (StockReservationReleased) EventName() string       { return "StockReservationReleased" }
func (e StockReservationReleased) ArticleID() string     { return e.articleID }
func (e StockReservationReleased) LocationCode() string  { return e.locationCode }
func (e StockReservationReleased) Quantity() int         { return e.quantity }
func (e StockReservationReleased) ReservationID() string { return e.reservationID }
func (e StockReservationReleased) Reason() string        { return e.reason }
