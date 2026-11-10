package main

import (
	"context"
	"log/slog"
	"net/http"

	"sample-app/internal/platform/messaging"
	"sample-app/internal/platform/messaging/httppush"
)

func startConsumer(_ context.Context, log *slog.Logger, mux *http.ServeMux, handler messaging.Handler) {
	mux.Handle("POST /events", httppush.Receiver(handler))
	log.Info("receiving events", "endpoint", "POST /events")
}
