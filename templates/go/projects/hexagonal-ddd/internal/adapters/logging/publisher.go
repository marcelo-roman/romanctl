package logging

import (
	"context"
	"log/slog"

	"sample-app/internal/domain/order"
)

type Publisher struct {
	log *slog.Logger
}

func NewPublisher(log *slog.Logger) *Publisher { return &Publisher{log: log} }

func (p *Publisher) Publish(ctx context.Context, events []order.Event) {
	for _, e := range events {
		p.log.InfoContext(ctx, "domain event", "name", e.EventName(), "event", e)
	}
}
