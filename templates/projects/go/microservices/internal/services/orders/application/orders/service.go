package orders

import (
	"context"

	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

const maxListLimit = 200

type AddItem struct {
	ProductName string
	Quantity    int
	UnitPrice   int64
	Currency    string
}

type Service struct {
	store Store
	now   Clock
}

func NewService(store Store, now Clock) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Create(ctx context.Context, customerName string) (shared.ID, error) {
	o, err := order.Create(customerName, s.now())
	if err != nil {
		return "", err
	}
	if err := s.store.Save(ctx, o, nil); err != nil {
		return "", err
	}
	return o.ID(), nil
}

func (s *Service) AddItem(ctx context.Context, id shared.ID, in AddItem) error {
	currency := in.Currency
	if currency == "" {
		currency = order.DefaultCurrency
	}
	price, err := order.NewMoney(in.UnitPrice, currency)
	if err != nil {
		return err
	}
	return s.change(ctx, id, func(o *order.Order) error { return o.AddItem(in.ProductName, in.Quantity, price) })
}

func (s *Service) Place(ctx context.Context, id shared.ID) error {
	return s.change(ctx, id, func(o *order.Order) error { return o.Place(s.now()) })
}

func (s *Service) Get(ctx context.Context, id shared.ID) (OrderView, error) {
	o, err := s.store.Get(ctx, id)
	if err != nil {
		return OrderView{}, err
	}
	return viewOf(o), nil
}

func (s *Service) List(ctx context.Context, limit int) ([]OrderView, error) {
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	list, err := s.store.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	views := make([]OrderView, 0, len(list))
	for _, o := range list {
		views = append(views, viewOf(o))
	}
	return views, nil
}

func (s *Service) change(ctx context.Context, id shared.ID, apply func(*order.Order) error) error {
	o, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := apply(o); err != nil {
		return err
	}
	outbox, err := integrationEvents(o.PullEvents())
	if err != nil {
		return err
	}
	return s.store.Save(ctx, o, outbox)
}
