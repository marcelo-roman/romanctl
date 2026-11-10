package orders

import (
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

const OrderPlacedType = "orders.order-placed"

type OrderPlaced struct {
	OrderID      string    `json:"orderId"`
	CustomerName string    `json:"customerName"`
	TotalCents   int64     `json:"totalCents"`
	Currency     string    `json:"currency"`
	PlacedAt     time.Time `json:"placedAt"`
}

func integrationEvents(events []order.Event) ([]messaging.Envelope, error) {
	var out []messaging.Envelope
	for _, e := range events {
		placed, ok := e.(order.PlacedEvent)
		if !ok {
			continue
		}
		envelope, err := messaging.NewEnvelope(string(shared.NewID()), OrderPlacedType, placed.At, OrderPlaced{
			OrderID:      string(placed.OrderID),
			CustomerName: placed.CustomerName,
			TotalCents:   placed.Total,
			Currency:     placed.Currency,
			PlacedAt:     placed.At,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, envelope)
	}
	return out, nil
}
