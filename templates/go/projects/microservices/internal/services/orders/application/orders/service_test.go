package orders

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

type fakeStore struct {
	saved  map[shared.ID]*order.Order
	outbox []messaging.Envelope
}

func (s *fakeStore) Save(_ context.Context, o *order.Order, outbox []messaging.Envelope) error {
	s.saved[o.ID()] = o
	s.outbox = append(s.outbox, outbox...)
	return nil
}

func (s *fakeStore) Get(_ context.Context, id shared.ID) (*order.Order, error) {
	o, ok := s.saved[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

func (s *fakeStore) List(context.Context, int) ([]*order.Order, error) { return nil, nil }

func newService() (*Service, *fakeStore) {
	store := &fakeStore{saved: map[shared.ID]*order.Order{}}
	now := func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	return NewService(store, now), store
}

func TestPlacingWritesOrderPlacedToTheOutbox(t *testing.T) {
	ctx := context.Background()
	svc, store := newService()
	id, err := svc.Create(ctx, "Ana")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddItem(ctx, id, AddItem{ProductName: "Caneta", Quantity: 3, UnitPrice: 250}); err != nil {
		t.Fatal(err)
	}
	if len(store.outbox) != 0 {
		t.Fatalf("draft changes must not publish: %+v", store.outbox)
	}
	if err := svc.Place(ctx, id); err != nil {
		t.Fatal(err)
	}
	view, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != "Placed" || view.Total != 750 || view.Currency != "BRL" {
		t.Fatalf("view %+v", view)
	}
	if len(store.outbox) != 1 || store.outbox[0].Type != OrderPlacedType || store.outbox[0].ID == "" {
		t.Fatalf("outbox %+v", store.outbox)
	}
	var placed OrderPlaced
	if err := json.Unmarshal(store.outbox[0].Payload, &placed); err != nil {
		t.Fatal(err)
	}
	if placed.OrderID != string(id) || placed.TotalCents != 750 || placed.Currency != "BRL" || placed.CustomerName != "Ana" {
		t.Fatalf("payload %+v", placed)
	}
}

func TestUnknownOrderIsNotFound(t *testing.T) {
	svc, _ := newService()
	if err := svc.Place(context.Background(), shared.NewID()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
