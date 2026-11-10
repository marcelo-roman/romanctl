package membus

import (
	"context"
	"sync"

	"sample-app/internal/platform/messaging"
)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]messaging.Handler
}

func New() *Bus { return &Bus{handlers: map[string][]messaging.Handler{}} }

func (b *Bus) Subscribe(eventType string, h messaging.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

func (b *Bus) Publish(ctx context.Context, e messaging.Envelope) error {
	b.mu.RLock()
	handlers := append([]messaging.Handler(nil), b.handlers[e.Type]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
