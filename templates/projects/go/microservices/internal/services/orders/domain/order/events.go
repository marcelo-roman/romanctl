package order

import (
	"time"

	"sample-app/internal/services/orders/domain/shared"
)

type Event interface {
	EventName() string
}

type PlacedEvent struct {
	OrderID      shared.ID
	CustomerName string
	Total        int64
	Currency     string
	At           time.Time
}

func (PlacedEvent) EventName() string { return "order.placed" }
