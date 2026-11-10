package orders

import (
	"context"
	"errors"
	"testing"
	"time"

	"sample-app/internal/domain/order"
	"sample-app/internal/domain/shared"
)

type fakeRepo struct{ saved map[shared.ID]*order.Order }

func (r *fakeRepo) Save(_ context.Context, o *order.Order) error { r.saved[o.ID()] = o; return nil }

func (r *fakeRepo) Get(_ context.Context, id shared.ID) (*order.Order, error) {
	o, ok := r.saved[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

func (r *fakeRepo) List(context.Context, int) ([]*order.Order, error) { return nil, nil }

type fakeEvents struct{ published []order.Event }

func (f *fakeEvents) Publish(_ context.Context, events []order.Event) {
	f.published = append(f.published, events...)
}

func newService() (*Service, *fakeEvents) {
	events := &fakeEvents{}
	now := func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	return NewService(&fakeRepo{saved: map[shared.ID]*order.Order{}}, events, now), events
}

func TestPlacingPublishesEvent(t *testing.T) {
	ctx := context.Background()
	svc, events := newService()
	id, err := svc.Create(ctx, "Ana")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddItem(ctx, id, AddItem{ProductName: "Caneta", Quantity: 3, UnitPrice: 250}); err != nil {
		t.Fatal(err)
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
	if len(events.published) != 1 {
		t.Fatalf("events %+v", events.published)
	}
}

func TestUnknownOrderIsNotFound(t *testing.T) {
	svc, _ := newService()
	if err := svc.Place(context.Background(), shared.NewID()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
