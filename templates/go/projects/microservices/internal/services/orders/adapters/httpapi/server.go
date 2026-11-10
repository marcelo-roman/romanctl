package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"sample-app/internal/platform/httpx"
	"sample-app/internal/services/orders/application/orders"
	// romanctl:imports
	"sample-app/internal/services/orders/domain/shared"
)

type Services struct {
	Orders *orders.Service
	// romanctl:services
}

func NewHandler(s Services) *http.ServeMux {
	mux := http.NewServeMux()
	httpx.Health(mux)
	registerOrders(mux, s.Orders)
	// romanctl:routes
	return mux
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool { return httpx.Decode(w, r, into) }

func writeJSON(w http.ResponseWriter, status int, body any) { httpx.WriteJSON(w, status, body) }

func writeError(w http.ResponseWriter, err error) {
	var ruleErr *shared.RuleError
	switch {
	case errors.Is(err, shared.ErrNotFound):
		httpx.WriteProblem(w, http.StatusNotFound, "Não encontrado", err.Error())
	case errors.As(err, &ruleErr):
		httpx.WriteProblem(w, http.StatusUnprocessableEntity, "Regra de negócio violada", err.Error())
	default:
		slog.Error("request failed", "error", err)
		httpx.WriteProblem(w, http.StatusInternalServerError, "Erro interno", "")
	}
}
