package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"sample-app/internal/application/orders"
	// romanctl:imports
	"sample-app/internal/domain/shared"
)

type problem struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type Services struct {
	Orders *orders.Service
	// romanctl:services
}

func NewHandler(s Services) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	registerOrders(mux, s.Orders)
	// romanctl:routes
	return mux
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		writeJSON(w, http.StatusBadRequest, problem{Status: http.StatusBadRequest, Title: "Requisição inválida", Detail: err.Error()})
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, err error) {
	var ruleErr *shared.RuleError
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeJSON(w, http.StatusNotFound, problem{Status: http.StatusNotFound, Title: "Não encontrado", Detail: err.Error()})
	case errors.As(err, &ruleErr):
		writeJSON(w, http.StatusUnprocessableEntity, problem{Status: http.StatusUnprocessableEntity, Title: "Regra de negócio violada", Detail: err.Error()})
	default:
		slog.Error("request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, problem{Status: http.StatusInternalServerError, Title: "Erro interno"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
