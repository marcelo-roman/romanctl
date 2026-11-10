package httppush

import (
	"errors"
	"log/slog"
	"net/http"

	"sample-app/internal/platform/httpx"
	"sample-app/internal/platform/messaging"
)

func Receiver(handler messaging.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var e messaging.Envelope
		if !httpx.Decode(w, r, &e) {
			return
		}
		if err := e.Validate(); err != nil {
			httpx.WriteProblem(w, http.StatusBadRequest, "Evento inválido", err.Error())
			return
		}
		err := handler(r.Context(), e)
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, messaging.ErrRejected):
			httpx.WriteProblem(w, http.StatusUnprocessableEntity, "Evento recusado", err.Error())
		default:
			slog.Error("event handling failed", "id", e.ID, "type", e.Type, "error", err)
			httpx.WriteProblem(w, http.StatusServiceUnavailable, "Evento não processado", "")
		}
	})
}
