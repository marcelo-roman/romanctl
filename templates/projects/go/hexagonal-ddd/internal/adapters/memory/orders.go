package memory

import (
	"context"
	"slices"
	"sync"

	"sample-app/internal/application/orders"
	"sample-app/internal/domain/order"
	"sample-app/internal/domain/shared"
)

type OrderRepository struct {
	mu     sync.RWMutex
	orders map[shared.ID]*order.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{orders: map[shared.ID]*order.Order{}}
}

func (r *OrderRepository) Save(_ context.Context, o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID()] = clone(o)
	return nil
}

func (r *OrderRepository) Get(_ context.Context, id shared.ID) (*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id]
	if !ok {
		return nil, orders.ErrNotFound
	}
	return clone(o), nil
}

func (r *OrderRepository) List(_ context.Context, limit int) ([]*order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*order.Order, 0, len(r.orders))
	for _, o := range r.orders {
		list = append(list, clone(o))
	}
	slices.SortFunc(list, func(a, b *order.Order) int { return b.CreatedAt().Compare(a.CreatedAt()) })
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func clone(o *order.Order) *order.Order {
	return order.Restore(o.ID(), o.CustomerName(), o.Status(), o.CreatedAt(), o.Items())
}
