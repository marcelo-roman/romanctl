package e2e

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/membus"
	"sample-app/internal/platform/messaging/outbox"
	billinghttp "sample-app/internal/services/billing/adapters/httpapi"
	billingmemory "sample-app/internal/services/billing/adapters/memory"
	"sample-app/internal/services/billing/application/invoices"
	ordershttp "sample-app/internal/services/orders/adapters/httpapi"
	ordersmemory "sample-app/internal/services/orders/adapters/memory"
	"sample-app/internal/services/orders/application/orders"
)

type system struct {
	orders  *httptest.Server
	billing *httptest.Server
	relay   *outbox.Relay
	bus     *membus.Bus
	store   *ordersmemory.Store
}

func start(t *testing.T) system {
	t.Helper()
	store := ordersmemory.NewStore()
	orderServer := httptest.NewServer(ordershttp.NewHandler(ordershttp.Services{Orders: orders.NewService(store, time.Now)}))
	t.Cleanup(orderServer.Close)

	billing := invoices.NewService(billingmemory.NewStore(), time.Now)
	billingServer := httptest.NewServer(billinghttp.NewHandler(billinghttp.Services{Invoices: billing}))
	t.Cleanup(billingServer.Close)

	bus := membus.New()
	bus.Subscribe(invoices.OrderPlacedType, billing.Handle)
	relay := outbox.NewRelay(store, bus, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Millisecond)
	return system{orders: orderServer, billing: billingServer, relay: relay, bus: bus, store: store}
}

func call(t *testing.T, method, url, body string, into any) int {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if into != nil {
		_ = json.NewDecoder(res.Body).Decode(into)
	}
	return res.StatusCode
}

func placeOrder(t *testing.T, s system) string {
	t.Helper()
	var created struct{ ID string }
	if code := call(t, "POST", s.orders.URL+"/orders", `{"customerName":"Ana"}`, &created); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	if code := call(t, "POST", s.orders.URL+"/orders/"+created.ID+"/items", `{"productName":"Caneta","quantity":3,"unitPrice":2.5}`, nil); code != http.StatusNoContent {
		t.Fatalf("add item: %d", code)
	}
	if code := call(t, "POST", s.orders.URL+"/orders/"+created.ID+"/place", ``, nil); code != http.StatusNoContent {
		t.Fatalf("place: %d", code)
	}
	return created.ID
}

func TestPlacedOrderBecomesAnInvoice(t *testing.T) {
	s := start(t)
	orderID := placeOrder(t, s)

	if code := call(t, "GET", s.billing.URL+"/invoices/"+orderID, ``, nil); code != http.StatusNotFound {
		t.Fatalf("invoice before the relay ran: %d", code)
	}
	if sent, err := s.relay.RunOnce(context.Background()); err != nil || sent != 1 {
		t.Fatalf("relay sent=%d err=%v", sent, err)
	}
	var invoice struct {
		OrderID  string
		Amount   float64
		Currency string
		Status   string
	}
	if code := call(t, "GET", s.billing.URL+"/invoices/"+orderID, ``, &invoice); code != http.StatusOK {
		t.Fatalf("invoice: %d", code)
	}
	if invoice.OrderID != orderID || invoice.Amount != 7.5 || invoice.Currency != "BRL" || invoice.Status != "Issued" {
		t.Fatalf("invoice %+v", invoice)
	}
	if sent, _ := s.relay.RunOnce(context.Background()); sent != 0 {
		t.Fatalf("sent %d events twice", sent)
	}
}

func TestRedeliveredEventDoesNotDuplicateTheInvoice(t *testing.T) {
	s := start(t)
	placeOrder(t, s)
	pending, _ := s.store.Pending(context.Background(), 10)
	if len(pending) != 1 {
		t.Fatalf("outbox %+v", pending)
	}
	for range 3 {
		if err := s.bus.Publish(context.Background(), pending[0]); err != nil {
			t.Fatal(err)
		}
	}
	var list []json.RawMessage
	call(t, "GET", s.billing.URL+"/invoices", ``, &list)
	if len(list) != 1 {
		t.Fatalf("invoices %s", list)
	}
}

func TestBillingFailureKeepsTheEventInTheOutbox(t *testing.T) {
	s := start(t)
	failing := true
	s.bus.Subscribe(invoices.OrderPlacedType, func(context.Context, messaging.Envelope) error {
		if failing {
			return io.ErrUnexpectedEOF
		}
		return nil
	})
	placeOrder(t, s)
	if _, err := s.relay.RunOnce(context.Background()); err == nil {
		t.Fatal("expected the relay to fail")
	}
	if pending, _ := s.store.Pending(context.Background(), 10); len(pending) != 1 {
		t.Fatalf("event lost: %+v", pending)
	}
	failing = false
	if sent, err := s.relay.RunOnce(context.Background()); err != nil || sent != 1 {
		t.Fatalf("relay sent=%d err=%v", sent, err)
	}
	var list []json.RawMessage
	call(t, "GET", s.billing.URL+"/invoices", ``, &list)
	if len(list) != 1 {
		t.Fatalf("invoices %s", list)
	}
}
