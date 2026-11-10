package orders

import (
	"context"
	"fmt"
	"time"

	"sample-app/internal/domain/order"
	"sample-app/internal/domain/shared"
)

var ErrNotFound = fmt.Errorf("Pedido %w.", shared.ErrNotFound)

type Repository interface {
	Save(ctx context.Context, o *order.Order) error
	Get(ctx context.Context, id shared.ID) (*order.Order, error)
	List(ctx context.Context, limit int) ([]*order.Order, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, events []order.Event)
}

type Clock func() time.Time
