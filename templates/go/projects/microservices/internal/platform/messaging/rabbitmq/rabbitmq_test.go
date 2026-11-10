package rabbitmq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sample-app/internal/platform/messaging"
)

func TestEnvelopeFallsBackToTheRoutingKey(t *testing.T) {
	e := envelopeOf(amqp.Delivery{MessageId: "1", RoutingKey: "orders.order-placed", Body: []byte(`{}`)})
	if e.ID != "1" || e.Type != "orders.order-placed" || string(e.Payload) != `{}` {
		t.Fatalf("envelope %+v", e)
	}
}

func brokerURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		t.Skip("RABBITMQ_URL not set: romanctl up --no-run starts RabbitMQ and writes the .env")
	}
	waitForBroker(t, url)
	return url
}

func waitForBroker(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for {
		conn, err := amqp.Dial(url)
		if err == nil {
			conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("RabbitMQ did not accept connections: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func uniqueTopology(t *testing.T, url string) Topology {
	t.Helper()
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	suffix := hex.EncodeToString(b)
	topology := Topology{Exchange: "sample-app.test-" + suffix, Queue: "test-" + suffix, Bindings: []string{"orders.order-placed"}}
	t.Cleanup(func() {
		conn, err := amqp.Dial(url)
		if err != nil {
			return
		}
		defer conn.Close()
		ch, err := conn.Channel()
		if err != nil {
			return
		}
		_, _ = ch.QueueDelete(topology.Queue, false, false, false)
		_ = ch.ExchangeDelete(topology.Exchange, false, false)
	})
	return topology
}

type recorder struct {
	mu       sync.Mutex
	received []messaging.Envelope
	failures int
	arrived  chan struct{}
}

func (r *recorder) handle(_ context.Context, e messaging.Envelope) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.received = append(r.received, e)
	r.arrived <- struct{}{}
	if r.failures > 0 {
		r.failures--
		return errors.New("falha passageira")
	}
	return nil
}

func TestPublishAndConsumeRoundTrip(t *testing.T) {
	url := brokerURL(t)
	topology := uniqueTopology(t, url)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	publisher := NewPublisher(url, topology.Exchange)
	defer publisher.Close()
	e, _ := messaging.NewEnvelope("00000000-0000-4000-8000-000000000001", "orders.order-placed", time.Now().Truncate(time.Second), map[string]int{"totalCents": 750})

	if err := publisher.Publish(ctx, e); err == nil {
		t.Fatal("published with no queue bound: the message would be lost")
	}

	rec := &recorder{failures: 1, arrived: make(chan struct{}, 4)}
	consumer := NewConsumer(url, topology)
	consumer.RetryDelay = 10 * time.Millisecond
	stopped := make(chan error, 1)
	go func() { stopped <- consumer.Run(ctx, rec.handle) }()

	deadline := time.Now().Add(10 * time.Second)
	for err := publisher.Publish(ctx, e); err != nil; err = publisher.Publish(ctx, e) {
		if time.Now().After(deadline) {
			t.Fatalf("queue never bound: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	for range 2 {
		select {
		case <-rec.arrived:
		case <-ctx.Done():
			t.Fatal("message not delivered and redelivered")
		}
	}
	cancel()
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	for _, got := range rec.received {
		if got.ID != e.ID || got.Type != e.Type || string(got.Payload) != string(e.Payload) || !got.OccurredAt.Equal(e.OccurredAt) {
			t.Fatalf("received %+v, want %+v", got, e)
		}
	}
}
