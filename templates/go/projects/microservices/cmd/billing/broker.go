package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/rabbitmq"
)

const (
	exchange    = "sample-app.events"
	queue       = "billing.order-placed"
	orderPlaced = "orders.order-placed"
)

func startConsumer(ctx context.Context, log *slog.Logger, _ *http.ServeMux, handler messaging.Handler) {
	consumer := rabbitmq.NewConsumer(envOr("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672"),
		rabbitmq.Topology{Exchange: exchange, Queue: queue, Bindings: []string{orderPlaced}})
	go func() {
		for ctx.Err() == nil {
			err := consumer.Run(ctx, handler)
			if ctx.Err() != nil {
				return
			}
			log.Warn("rabbitmq consumer stopped, reconnecting", "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}()
	log.Info("consuming", "queue", queue, "exchange", exchange)
}
