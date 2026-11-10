package invoices_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/billing/adapters/memory"
	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/shared"
)

func orderPlaced(t *testing.T, eventID string, orderID shared.ID) messaging.Envelope {
	t.Helper()
	e, err := messaging.NewEnvelope(eventID, invoices.OrderPlacedType, time.Now(),
		map[string]any{"orderId": orderID, "customerName": "Ana", "totalCents": 750, "currency": "BRL"})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func newService() *invoices.Service { return invoices.NewService(memory.NewStore(), time.Now) }

func TestSameEventTwiceIssuesOneInvoice(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	order := shared.NewID()
	e := orderPlaced(t, string(shared.NewID()), order)
	for range 2 {
		if err := svc.Handle(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := svc.List(ctx, 0)
	if len(list) != 1 {
		t.Fatalf("invoices %+v", list)
	}
	got, err := svc.ByOrder(ctx, order)
	if err != nil {
		t.Fatal(err)
	}
	if got.Amount != 750 || got.Currency != "BRL" || got.Status != "Issued" || got.OrderID != string(order) {
		t.Fatalf("invoice %+v", got)
	}
}

func TestAnotherEventForTheSameOrderDoesNotIssueAgain(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	order := shared.NewID()
	_ = svc.Handle(ctx, orderPlaced(t, string(shared.NewID()), order))
	_ = svc.Handle(ctx, orderPlaced(t, string(shared.NewID()), order))
	if list, _ := svc.List(ctx, 0); len(list) != 1 {
		t.Fatalf("invoices %+v", list)
	}
}

func TestOtherEventsAreIgnoredAndBadPayloadsRejected(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	if err := svc.Handle(ctx, messaging.Envelope{ID: "1", Type: "orders.order-cancelled", Payload: []byte(`{}`)}); err != nil {
		t.Fatal(err)
	}
	bad := messaging.Envelope{ID: "2", Type: invoices.OrderPlacedType, Payload: []byte(`{"orderId":"x","totalCents":1,"currency":"BRL"}`)}
	if err := svc.Handle(ctx, bad); !errors.Is(err, messaging.ErrRejected) {
		t.Fatalf("expected rejection, got %v", err)
	}
	if _, err := svc.ByOrder(ctx, shared.NewID()); !errors.Is(err, invoices.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
