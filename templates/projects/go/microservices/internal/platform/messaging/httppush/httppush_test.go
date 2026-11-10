package httppush

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
)

func TestPublisherDeliversToReceiver(t *testing.T) {
	var got []messaging.Envelope
	srv := httptest.NewServer(Receiver(func(_ context.Context, e messaging.Envelope) error {
		got = append(got, e)
		return nil
	}))
	t.Cleanup(srv.Close)

	e, _ := messaging.NewEnvelope("1", "orders.order-placed", time.Now(), map[string]int{"totalCents": 750})
	if err := NewPublisher(srv.URL).Publish(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "1" || string(got[0].Payload) != `{"totalCents":750}` {
		t.Fatalf("received %+v", got)
	}
}

func TestHandlerFailureIsAPublishFailure(t *testing.T) {
	for _, failure := range []error{errors.New("banco fora do ar"), messaging.Reject(errors.New("payload inválido"))} {
		srv := httptest.NewServer(Receiver(func(context.Context, messaging.Envelope) error { return failure }))
		err := NewPublisher(srv.URL).Publish(context.Background(), messaging.Envelope{ID: "1", Type: "x"})
		srv.Close()
		if err == nil {
			t.Fatalf("%v: expected the publisher to fail so the relay retries", failure)
		}
	}
}

func TestReceiverRejectsInvalidEvents(t *testing.T) {
	srv := httptest.NewServer(Receiver(func(context.Context, messaging.Envelope) error { return nil }))
	t.Cleanup(srv.Close)
	for _, body := range []string{`{`, `{"type":"x"}`} {
		res, err := http.Post(srv.URL, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s: %d", body, res.StatusCode)
		}
	}
}
