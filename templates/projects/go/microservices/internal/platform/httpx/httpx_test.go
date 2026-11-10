package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBadJSONBecomesProblem(t *testing.T) {
	rec := httptest.NewRecorder()
	var into map[string]any
	if Decode(rec, httptest.NewRequest("POST", "/", strings.NewReader("{")), &into) {
		t.Fatal("decoded invalid JSON")
	}
	var p Problem
	_ = json.NewDecoder(rec.Body).Decode(&p)
	if rec.Code != http.StatusBadRequest || p.Status != http.StatusBadRequest || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("%d %+v", rec.Code, p)
	}
}

func TestHealth(t *testing.T) {
	mux := http.NewServeMux()
	Health(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
