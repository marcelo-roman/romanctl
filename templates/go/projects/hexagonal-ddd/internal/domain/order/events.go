package order

import (
	"time"

	"sample-app/internal/domain/shared"
)

type Event interface {
	EventName() string
}

type PlacedEvent struct {
	OrderID  shared.ID
	Total    int64
	Currency string
	At       time.Time
}

func (PlacedEvent) EventName() string { return "order.placed" }
