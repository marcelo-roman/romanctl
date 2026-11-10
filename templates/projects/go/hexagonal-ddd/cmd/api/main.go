package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sample-app/internal/adapters/httpapi"
	"sample-app/internal/adapters/logging"

	"sample-app/internal/adapters/postgres"

	"sample-app/internal/application/orders"
	// romanctl:imports
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(ctx, envOr("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/sample_app"))
	if err != nil {
		return err
	}
	defer db.Close()

	events := logging.NewPublisher(log)
	now := func() time.Time { return time.Now().UTC() }
	services := httpapi.Services{

		Orders: orders.NewService(postgres.NewOrderRepository(db), events, now),

		// romanctl:services
	}

	srv := &http.Server{Addr: ":" + envOr("PORT", "8090"), Handler: httpapi.NewHandler(services), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	log.Info("listening", "addr", "http://localhost"+srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
