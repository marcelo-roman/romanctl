package memory

import (
	"context"
	"slices"
	"sync"

	"sample-app/internal/services/orders/application/sampleitems"
	"sample-app/internal/services/orders/domain/sampleitem"
	"sample-app/internal/services/orders/domain/shared"
)

type SampleItemRepository struct {
	mu    sync.RWMutex
	items map[shared.ID]*sampleitem.SampleItem
}

func NewSampleItemRepository() *SampleItemRepository {
	return &SampleItemRepository{items: map[shared.ID]*sampleitem.SampleItem{}}
}

func (r *SampleItemRepository) Save(_ context.Context, e *sampleitem.SampleItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[e.ID()] = sampleitem.Restore(e.ID(), e.Name(), e.CreatedAt())
	return nil
}

func (r *SampleItemRepository) Get(_ context.Context, id shared.ID) (*sampleitem.SampleItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.items[id]
	if !ok {
		return nil, sampleitems.ErrNotFound
	}
	return sampleitem.Restore(e.ID(), e.Name(), e.CreatedAt()), nil
}

func (r *SampleItemRepository) List(_ context.Context, limit int) ([]*sampleitem.SampleItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*sampleitem.SampleItem, 0, len(r.items))
	for _, e := range r.items {
		list = append(list, sampleitem.Restore(e.ID(), e.Name(), e.CreatedAt()))
	}
	slices.SortFunc(list, func(a, b *sampleitem.SampleItem) int { return b.CreatedAt().Compare(a.CreatedAt()) })
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *SampleItemRepository) Delete(_ context.Context, id shared.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return sampleitems.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
