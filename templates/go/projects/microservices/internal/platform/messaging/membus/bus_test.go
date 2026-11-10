package membus

import (
	"context"
	"errors"
	"testing"

	"sample-app/internal/platform/messaging"
)

func TestDeliversOnlyToSubscribersOfTheType(t *testing.T) {
	bus := New()
	var got []string
	bus.Subscribe("a", func(_ context.Context, e messaging.Envelope) error { got = append(got, "a:"+e.ID); return nil })
	bus.Subscribe("b", func(_ context.Context, e messaging.Envelope) error { got = append(got, "b:"+e.ID); return nil })
	if err := bus.Publish(context.Background(), messaging.Envelope{ID: "1", Type: "a"}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "a:1" {
		t.Fatalf("delivered %v", got)
	}
}

func TestHandlerErrorReachesThePublisher(t *testing.T) {
	bus := New()
	boom := errors.New("boom")
	bus.Subscribe("a", func(context.Context, messaging.Envelope) error { return boom })
	if err := bus.Publish(context.Background(), messaging.Envelope{ID: "1", Type: "a"}); !errors.Is(err, boom) {
		t.Fatalf("expected handler error, got %v", err)
	}
}
