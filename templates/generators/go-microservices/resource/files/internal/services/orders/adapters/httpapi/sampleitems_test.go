package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sample-app/internal/services/orders/adapters/memory"
	"sample-app/internal/services/orders/application/sampleitems"
)

func TestSampleItemCrud(t *testing.T) {
	srv := httptest.NewServer(NewHandler(Services{SampleItems: sampleitems.NewService(memory.NewSampleItemRepository(), time.Now)}))
	t.Cleanup(srv.Close)

	res := send(t, "POST", srv.URL+"/sample-items", `{"name":"Primeiro"}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", res.StatusCode)
	}
	var created struct{ ID string }
	_ = json.NewDecoder(res.Body).Decode(&created)

	if res := send(t, "PUT", srv.URL+"/sample-items/"+created.ID, `{"name":"Segundo"}`); res.StatusCode != http.StatusNoContent {
		t.Fatalf("rename: %d", res.StatusCode)
	}
	res = send(t, "GET", srv.URL+"/sample-items/"+created.ID, ``)
	var got sampleItemResponse
	_ = json.NewDecoder(res.Body).Decode(&got)
	if got.Name != "Segundo" {
		t.Fatalf("got %+v", got)
	}
	if res := send(t, "DELETE", srv.URL+"/sample-items/"+created.ID, ``); res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", res.StatusCode)
	}
	if res := send(t, "GET", srv.URL+"/sample-items/"+created.ID, ``); res.StatusCode != http.StatusNotFound {
		t.Fatalf("after delete: %d", res.StatusCode)
	}
	if res := send(t, "POST", srv.URL+"/sample-items", `{"name":" "}`); res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty name: %d", res.StatusCode)
	}
}
