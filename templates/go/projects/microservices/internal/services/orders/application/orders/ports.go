package orders

import (
	"context"
	"fmt"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/services/orders/domain/order"
	"sample-app/internal/services/orders/domain/shared"
)

var ErrNotFound = fmt.Errorf("Pedido %w.", shared.ErrNotFound)

type Store interface {
	Save(ctx context.Context, o *order.Order, outbox []messaging.Envelope) error
	Get(ctx context.Context, id shared.ID) (*order.Order, error)
	List(ctx context.Context, limit int) ([]*order.Order, error)
}

type Clock func() time.Time
