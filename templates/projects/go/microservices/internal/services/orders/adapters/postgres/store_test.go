package postgres

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/orders/application/orders"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

func newOrder(t *testing.T) *order.Order {
	t.Helper()
	o, _ := order.Create("Ana", time.Now().UTC().Truncate(time.Microsecond))
	price, _ := order.NewMoney(350, "BRL")
	if err := o.AddItem("Caneta", 2, price); err != nil {
		t.Fatal(err)
	}
	return o
}

func pendingIDs(t *testing.T, store *Store) []string {
	t.Helper()
	pending, err := store.Pending(context.Background(), 1000)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(pending))
	for _, e := range pending {
		ids = append(ids, e.ID)
	}
	return ids
}

func TestOrderRoundTrip(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()
	o := newOrder(t)
	if err := store.Save(ctx, o, nil); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, o.ID())
	if err != nil {
		t.Fatal(err)
	}
	if got.Total().Cents() != 700 || len(got.Items()) != 1 || got.CustomerName() != "Ana" {
		t.Fatalf("loaded %+v", got)
	}
	if _, err := store.Get(ctx, shared.NewID()); !errors.Is(err, orders.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestOutboxIsWrittenInTheSameTransaction(t *testing.T) {
	store := NewStore(openTestDB(t))
	ctx := context.Background()
	message, _ := messaging.NewEnvelope(string(shared.NewID()), orders.OrderPlacedType, time.Now(), map[string]int{"totalCents": 700})

	placed := newOrder(t)
	if err := store.Save(ctx, placed, []messaging.Envelope{message}); err != nil {
		t.Fatal(err)
	}
	pending, _ := store.Pending(ctx, 1000)
	i := slices.IndexFunc(pending, func(e messaging.Envelope) bool { return e.ID == message.ID })
	if i < 0 || pending[i].Type != orders.OrderPlacedType || string(pending[i].Payload) != `{"totalCents": 700}` {
		t.Fatalf("message not pending: %+v", pending)
	}
	if err := store.MarkSent(ctx, message.ID); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(pendingIDs(t, store), message.ID) {
		t.Fatal("still pending after being marked as sent")
	}

	rolledBack := newOrder(t)
	if err := store.Save(ctx, rolledBack, []messaging.Envelope{message}); err == nil {
		t.Fatal("saving a duplicated outbox message must fail")
	}
	if _, err := store.Get(ctx, rolledBack.ID()); !errors.Is(err, orders.ErrNotFound) {
		t.Fatalf("the order was saved without its outbox message: %v", err)
	}
}
