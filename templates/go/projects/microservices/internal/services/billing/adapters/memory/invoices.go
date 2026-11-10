package memory

import (
	"context"
	"slices"
	"sync"

	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/invoice"
	"sample-app/internal/services/billing/domain/shared"
)

type Store struct {
	mu        sync.RWMutex
	invoices  map[shared.ID]*invoice.Invoice
	processed map[string]bool
}

func NewStore() *Store {
	return &Store{invoices: map[shared.ID]*invoice.Invoice{}, processed: map[string]bool{}}
}

func (s *Store) Issue(_ context.Context, eventID string, i *invoice.Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.processed[eventID] {
		return nil
	}
	s.processed[eventID] = true
	if _, exists := s.invoices[i.OrderID()]; exists {
		return nil
	}
	s.invoices[i.OrderID()] = clone(i)
	return nil
}

func (s *Store) ByOrder(_ context.Context, orderID shared.ID) (*invoice.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.invoices[orderID]
	if !ok {
		return nil, invoices.ErrNotFound
	}
	return clone(i), nil
}

func (s *Store) List(_ context.Context, limit int) ([]*invoice.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*invoice.Invoice, 0, len(s.invoices))
	for _, i := range s.invoices {
		list = append(list, clone(i))
	}
	slices.SortFunc(list, func(a, b *invoice.Invoice) int { return b.IssuedAt().Compare(a.IssuedAt()) })
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func clone(i *invoice.Invoice) *invoice.Invoice {
	return invoice.Restore(i.ID(), i.OrderID(), i.Amount(), i.Currency(), i.Status(), i.IssuedAt())
}
