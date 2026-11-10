package messaging

import (
	"errors"
	"testing"
	"time"
)

func TestEnvelopeCarriesPayload(t *testing.T) {
	e, err := NewEnvelope("1", "orders.order-placed", time.Date(2026, 1, 1, 9, 0, 0, 0, time.FixedZone("", -3*3600)), map[string]int{"totalCents": 750})
	if err != nil {
		t.Fatal(err)
	}
	if string(e.Payload) != `{"totalCents":750}` || e.OccurredAt.Location() != time.UTC {
		t.Fatalf("envelope %+v", e)
	}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeWithoutIDIsRejected(t *testing.T) {
	if err := (Envelope{Type: "x"}).Validate(); !errors.Is(err, ErrRejected) {
		t.Fatalf("expected rejection, got %v", err)
	}
}
