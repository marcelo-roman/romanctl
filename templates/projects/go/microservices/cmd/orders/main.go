package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sample-app/internal/platform/httpx"
	"sample-app/internal/platform/messaging/outbox"
	"sample-app/internal/services/orders/adapters/httpapi"
	"sample-app/internal/services/orders/application/orders"
	// romanctl:imports
)

type storage struct {
	orders interface {
		orders.Store
		outbox.Store
	}
	// romanctl:storage
	close func()
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "orders")
	if err := run(log); err != nil {
		log.Error("stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := openStorage(ctx)
	if err != nil {
		return err
	}
	defer st.close()

	publisher, closePublisher := openPublisher()
	defer closePublisher()
	go outbox.NewRelay(st.orders, publisher, log, 500*time.Millisecond).Run(ctx)

	now := func() time.Time { return time.Now().UTC() }
	services := httpapi.Services{
		Orders: orders.NewService(st.orders, now),
		// romanctl:services
	}
	return httpx.Serve(ctx, log, ":"+envOr("PORT", "8091"), httpapi.NewHandler(services))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
