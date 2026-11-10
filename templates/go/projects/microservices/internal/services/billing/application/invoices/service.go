package invoices

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/billing/domain/invoice"
	"sample-app/internal/services/billing/domain/shared"
)

const (
	OrderPlacedType = "orders.order-placed"
	maxListLimit    = 200
)

var ErrNotFound = fmt.Errorf("Fatura %w.", shared.ErrNotFound)

type Store interface {
	Issue(ctx context.Context, eventID string, i *invoice.Invoice) error
	ByOrder(ctx context.Context, orderID shared.ID) (*invoice.Invoice, error)
	List(ctx context.Context, limit int) ([]*invoice.Invoice, error)
}

type orderPlaced struct {
	OrderID    string `json:"orderId"`
	TotalCents int64  `json:"totalCents"`
	Currency   string `json:"currency"`
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Handle(ctx context.Context, e messaging.Envelope) error {
	if e.Type != OrderPlacedType {
		return nil
	}
	var placed orderPlaced
	if err := json.Unmarshal(e.Payload, &placed); err != nil {
		return messaging.Reject(err)
	}
	i, err := invoice.Issue(placed.OrderID, placed.TotalCents, placed.Currency, s.now())
	if err != nil {
		return messaging.Reject(err)
	}
	return s.store.Issue(ctx, e.ID, i)
}

func (s *Service) ByOrder(ctx context.Context, orderID shared.ID) (View, error) {
	i, err := s.store.ByOrder(ctx, orderID)
	if err != nil {
		return View{}, err
	}
	return viewOf(i), nil
}

func (s *Service) List(ctx context.Context, limit int) ([]View, error) {
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	list, err := s.store.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(list))
	for _, i := range list {
		views = append(views, viewOf(i))
	}
	return views, nil
}
