package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/billing/adapters/memory"
	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/shared"
)

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func TestInvoicesAreListedAndFoundByOrder(t *testing.T) {
	svc := invoices.NewService(memory.NewStore(), time.Now)
	srv := httptest.NewServer(NewHandler(Services{Invoices: svc}))
	t.Cleanup(srv.Close)

	order := shared.NewID()
	e, _ := messaging.NewEnvelope(string(shared.NewID()), invoices.OrderPlacedType, time.Now(),
		map[string]any{"orderId": order, "totalCents": 1250, "currency": "BRL"})
	if err := svc.Handle(context.Background(), e); err != nil {
		t.Fatal(err)
	}

	var list []invoiceResponse
	_ = json.NewDecoder(get(t, srv.URL+"/invoices").Body).Decode(&list)
	if len(list) != 1 || list[0].OrderID != string(order) {
		t.Fatalf("list %+v", list)
	}
	res := get(t, srv.URL+"/invoices/"+string(order))
	var got invoiceResponse
	_ = json.NewDecoder(res.Body).Decode(&got)
	if res.StatusCode != http.StatusOK || got.Amount != 12.5 || got.Status != "Issued" {
		t.Fatalf("%d %+v", res.StatusCode, got)
	}
}

func TestInvoiceErrorsBecomeProblems(t *testing.T) {
	srv := httptest.NewServer(NewHandler(Services{Invoices: invoices.NewService(memory.NewStore(), time.Now)}))
	t.Cleanup(srv.Close)
	if res := get(t, srv.URL+"/invoices/00000000-0000-4000-8000-000000000000"); res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing invoice: %d", res.StatusCode)
	}
	if res := get(t, srv.URL+"/invoices/x"); res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("invalid id: %d", res.StatusCode)
	}
	if res := get(t, srv.URL+"/health"); res.StatusCode != http.StatusOK {
		t.Fatalf("health: %d", res.StatusCode)
	}
}
