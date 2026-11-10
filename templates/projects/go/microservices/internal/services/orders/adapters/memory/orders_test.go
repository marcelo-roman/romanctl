package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/outbox"
	"sample-app/internal/services/orders/application/orders"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

var (
	_ orders.Store = (*Store)(nil)
	_ outbox.Store = (*Store)(nil)
)

func TestOutboxKeepsMessagesUntilSent(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	o, _ := order.Create("Ana", time.Now())
	if err := store.Save(ctx, o, []messaging.Envelope{{ID: "1", Type: "x"}, {ID: "2", Type: "x"}}); err != nil {
		t.Fatal(err)
	}
	pending, _ := store.Pending(ctx, 1)
	if len(pending) != 1 || pending[0].ID != "1" {
		t.Fatalf("pending %+v", pending)
	}
	_ = store.MarkSent(ctx, "1")
	pending, _ = store.Pending(ctx, 10)
	if len(pending) != 1 || pending[0].ID != "2" {
		t.Fatalf("pending %+v", pending)
	}
	if _, err := store.Get(ctx, shared.NewID()); !errors.Is(err, orders.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
