package sampleitem

import (
	"strings"
	"time"

	"sample-app/internal/domain/shared"
)

type SampleItem struct {
	id        shared.ID
	name      string
	createdAt time.Time
}

func Create(name string, now time.Time) (*SampleItem, error) {
	e := &SampleItem{id: shared.NewID(), createdAt: now}
	if err := e.Rename(name); err != nil {
		return nil, err
	}
	return e, nil
}

func Restore(id shared.ID, name string, createdAt time.Time) *SampleItem {
	return &SampleItem{id: id, name: name, createdAt: createdAt}
}

func (e *SampleItem) ID() shared.ID        { return e.id }
func (e *SampleItem) Name() string         { return e.name }
func (e *SampleItem) CreatedAt() time.Time { return e.createdAt }

func (e *SampleItem) Rename(name string) error {
	if strings.TrimSpace(name) == "" {
		return shared.Rule("O nome é obrigatório.")
	}
	e.name = strings.TrimSpace(name)
	return nil
}
