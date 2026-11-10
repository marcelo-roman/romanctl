package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sample-app/internal/services/orders/adapters/memory"
	"sample-app/internal/services/orders/application/orders"
)

func newServer(t *testing.T) (*httptest.Server, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	srv := httptest.NewServer(NewHandler(Services{Orders: orders.NewService(store, time.Now)}))
	t.Cleanup(srv.Close)
	return srv, store
}

func send(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func TestOrderLifecycle(t *testing.T) {
	srv, store := newServer(t)
	res := send(t, "POST", srv.URL+"/orders", `{"customerName":"Ana"}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", res.StatusCode)
	}
	var created struct{ ID string }
	_ = json.NewDecoder(res.Body).Decode(&created)

	if res := send(t, "POST", srv.URL+"/orders/"+created.ID+"/items", `{"productName":"Caneta","quantity":2,"unitPrice":3.5}`); res.StatusCode != http.StatusNoContent {
		t.Fatalf("add item: %d", res.StatusCode)
	}
	if res := send(t, "POST", srv.URL+"/orders/"+created.ID+"/place", ``); res.StatusCode != http.StatusNoContent {
		t.Fatalf("place: %d", res.StatusCode)
	}
	res = send(t, "GET", srv.URL+"/orders/"+created.ID, ``)
	var got orderResponse
	_ = json.NewDecoder(res.Body).Decode(&got)
	if got.Status != "Placed" || got.Total != 7 || len(got.Items) != 1 {
		t.Fatalf("order %+v", got)
	}
	pending, _ := store.Pending(context.Background(), 10)
	if len(pending) != 1 || pending[0].Type != orders.OrderPlacedType {
		t.Fatalf("outbox %+v", pending)
	}
	if res := send(t, "GET", srv.URL+"/orders", ``); res.StatusCode != http.StatusOK {
		t.Fatalf("list: %d", res.StatusCode)
	}
}

func TestErrorsBecomeProblems(t *testing.T) {
	srv, _ := newServer(t)
	if res := send(t, "POST", srv.URL+"/orders", `{"customerName":""}`); res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty customer: %d", res.StatusCode)
	}
	if res := send(t, "GET", srv.URL+"/orders/00000000-0000-4000-8000-000000000000", ``); res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing order: %d", res.StatusCode)
	}
	res := send(t, "POST", srv.URL+"/orders", `{`)
	if res.StatusCode != http.StatusBadRequest || res.Header.Get("Content-Type") != "application/problem+json" {
		t.Fatalf("bad json: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	if res := send(t, "GET", srv.URL+"/health", ``); res.StatusCode != http.StatusOK {
		t.Fatalf("health: %d", res.StatusCode)
	}
}
