package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sample-app/internal/platform/messaging"
)

const (
	batchSize = 100
	maxDelay  = 10 * time.Second
)

type Store interface {
	Pending(ctx context.Context, limit int) ([]messaging.Envelope, error)
	MarkSent(ctx context.Context, id string) error
}

type Relay struct {
	store     Store
	publisher messaging.Publisher
	log       *slog.Logger
	interval  time.Duration
}

func NewRelay(store Store, publisher messaging.Publisher, log *slog.Logger, interval time.Duration) *Relay {
	return &Relay{store: store, publisher: publisher, log: log, interval: interval}
}

func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	pending, err := r.store.Pending(ctx, batchSize)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, e := range pending {
		if err := r.publisher.Publish(ctx, e); err != nil {
			return sent, fmt.Errorf("publicar %s %s: %w", e.Type, e.ID, err)
		}
		if err := r.store.MarkSent(ctx, e.ID); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}

func (r *Relay) Run(ctx context.Context) {
	delay := r.interval
	for {
		sent, err := r.RunOnce(ctx)
		delay = r.nextDelay(delay, err)
		if sent > 0 {
			r.log.Info("outbox relay published", "events", sent)
		}
		if err != nil && ctx.Err() == nil {
			r.log.Warn("outbox relay failed", "error", err, "retryIn", delay)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (r *Relay) nextDelay(current time.Duration, err error) time.Duration {
	if err == nil {
		return r.interval
	}
	return min(current*2, maxDelay)
}
