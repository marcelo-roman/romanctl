package memory

import (
	"context"
	"slices"
	"sync"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/orders/application/orders"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

type outboxEntry struct {
	message messaging.Envelope
	sent    bool
}

type Store struct {
	mu     sync.RWMutex
	orders map[shared.ID]*order.Order
	outbox []outboxEntry
}

func NewStore() *Store {
	return &Store{orders: map[shared.ID]*order.Order{}}
}

func (s *Store) Save(_ context.Context, o *order.Order, outbox []messaging.Envelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID()] = clone(o)
	for _, m := range outbox {
		s.outbox = append(s.outbox, outboxEntry{message: m})
	}
	return nil
}

func (s *Store) Get(_ context.Context, id shared.ID) (*order.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, orders.ErrNotFound
	}
	return clone(o), nil
}

func (s *Store) List(_ context.Context, limit int) ([]*order.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*order.Order, 0, len(s.orders))
	for _, o := range s.orders {
		list = append(list, clone(o))
	}
	slices.SortFunc(list, func(a, b *order.Order) int { return b.CreatedAt().Compare(a.CreatedAt()) })
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (s *Store) Pending(_ context.Context, limit int) ([]messaging.Envelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []messaging.Envelope
	for _, entry := range s.outbox {
		if len(out) == limit {
			break
		}
		if !entry.sent {
			out = append(out, entry.message)
		}
	}
	return out, nil
}

func (s *Store) MarkSent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.outbox {
		if s.outbox[i].message.ID == id {
			s.outbox[i].sent = true
		}
	}
	return nil
}

func clone(o *order.Order) *order.Order {
	return order.Restore(o.ID(), o.CustomerName(), o.Status(), o.CreatedAt(), o.Items())
}
