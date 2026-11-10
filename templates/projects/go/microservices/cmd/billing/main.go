package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sample-app/internal/platform/httpx"
	"sample-app/internal/services/billing/adapters/httpapi"
	"sample-app/internal/services/billing/application/invoices"
)

type storage struct {
	invoices invoices.Store
	close    func()
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "billing")
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

	svc := invoices.NewService(st.invoices, func() time.Time { return time.Now().UTC() })
	mux := httpapi.NewHandler(httpapi.Services{Invoices: svc})
	startConsumer(ctx, log, mux, svc.Handle)
	return httpx.Serve(ctx, log, ":"+envOr("PORT", "8092"), mux)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
