package outbox

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"time"

	"sample-app/internal/platform/messaging"
)

type fakeStore struct {
	mu       sync.Mutex
	messages []messaging.Envelope
	sent     map[string]bool
}

func newStore(ids ...string) *fakeStore {
	s := &fakeStore{sent: map[string]bool{}}
	for _, id := range ids {
		s.messages = append(s.messages, messaging.Envelope{ID: id, Type: "orders.order-placed"})
	}
	return s
}

func (s *fakeStore) Pending(_ context.Context, limit int) ([]messaging.Envelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []messaging.Envelope
	for _, m := range s.messages {
		if !s.sent[m.ID] && len(out) < limit {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *fakeStore) MarkSent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent[id] = true
	return nil
}

func (s *fakeStore) isSent(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sent[id]
}

type flakyPublisher struct {
	mu        sync.Mutex
	failures  int
	attempts  int
	published []string
}

func (p *flakyPublisher) Publish(_ context.Context, e messaging.Envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.attempts++
	if p.failures > 0 {
		p.failures--
		return errors.New("broker fora do ar")
	}
	p.published = append(p.published, e.ID)
	return nil
}

func (p *flakyPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.attempts
}

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestMarksSentOnlyAfterPublish(t *testing.T) {
	ctx := context.Background()
	store := newStore("a")
	publisher := &flakyPublisher{failures: 1}
	relay := NewRelay(store, publisher, quiet, time.Millisecond)

	if sent, err := relay.RunOnce(ctx); err == nil || sent != 0 {
		t.Fatalf("expected failure, got sent=%d err=%v", sent, err)
	}
	if store.isSent("a") {
		t.Fatal("marked as sent before the broker accepted it")
	}
	if sent, err := relay.RunOnce(ctx); err != nil || sent != 1 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	if !store.isSent("a") {
		t.Fatal("not marked as sent after publishing")
	}
}

func TestStopsAtTheFirstFailureAndKeepsTheRestPending(t *testing.T) {
	store := newStore("a", "b")
	relay := NewRelay(store, &flakyPublisher{failures: 1}, quiet, time.Millisecond)
	if _, err := relay.RunOnce(context.Background()); err == nil {
		t.Fatal("expected failure")
	}
	if store.isSent("a") || store.isSent("b") {
		t.Fatal("nothing should be sent")
	}
}

func TestRunRetriesUntilPublished(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := newStore("a", "b")
	publisher := &flakyPublisher{failures: 3}
	done := make(chan struct{})
	go func() {
		NewRelay(store, publisher, quiet, time.Millisecond).Run(ctx)
		close(done)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for !(store.isSent("a") && store.isSent("b")) {
		if time.Now().After(deadline) {
			t.Fatalf("not delivered after %d attempts", publisher.count())
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done
	if !slices.Equal(publisher.published, []string{"a", "b"}) || publisher.count() != 5 {
		t.Fatalf("published %v in %d attempts", publisher.published, publisher.count())
	}
}

func TestBackoffDoublesUpToTheLimitAndResets(t *testing.T) {
	relay := NewRelay(newStore(), &flakyPublisher{}, quiet, time.Second)
	failure := errors.New("x")
	if got := relay.nextDelay(time.Second, failure); got != 2*time.Second {
		t.Fatalf("got %s", got)
	}
	if got := relay.nextDelay(8*time.Second, failure); got != maxDelay {
		t.Fatalf("got %s", got)
	}
	if got := relay.nextDelay(maxDelay, nil); got != time.Second {
		t.Fatalf("got %s", got)
	}
}
