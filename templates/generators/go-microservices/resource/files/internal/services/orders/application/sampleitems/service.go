package sampleitems

import (
	"context"
	"fmt"
	"time"

	"sample-app/internal/services/orders/domain/sampleitem"
	"sample-app/internal/services/orders/domain/shared"
)

const maxListLimit = 200

var ErrNotFound = fmt.Errorf("SampleItem %w.", shared.ErrNotFound)

type Repository interface {
	Save(ctx context.Context, e *sampleitem.SampleItem) error
	Get(ctx context.Context, id shared.ID) (*sampleitem.SampleItem, error)
	List(ctx context.Context, limit int) ([]*sampleitem.SampleItem, error)
	Delete(ctx context.Context, id shared.ID) error
}

type View struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) Create(ctx context.Context, name string) (shared.ID, error) {
	e, err := sampleitem.Create(name, s.now())
	if err != nil {
		return "", err
	}
	if err := s.repo.Save(ctx, e); err != nil {
		return "", err
	}
	return e.ID(), nil
}

func (s *Service) Rename(ctx context.Context, id shared.ID, name string) error {
	e, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := e.Rename(name); err != nil {
		return err
	}
	return s.repo.Save(ctx, e)
}

func (s *Service) Get(ctx context.Context, id shared.ID) (View, error) {
	e, err := s.repo.Get(ctx, id)
	if err != nil {
		return View{}, err
	}
	return viewOf(e), nil
}

func (s *Service) List(ctx context.Context, limit int) ([]View, error) {
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	list, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(list))
	for _, e := range list {
		views = append(views, viewOf(e))
	}
	return views, nil
}

func (s *Service) Delete(ctx context.Context, id shared.ID) error {
	return s.repo.Delete(ctx, id)
}

func viewOf(e *sampleitem.SampleItem) View {
	return View{ID: string(e.ID()), Name: e.Name(), CreatedAt: e.CreatedAt()}
}
