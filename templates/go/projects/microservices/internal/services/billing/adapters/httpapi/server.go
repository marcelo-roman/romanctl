package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"sample-app/internal/platform/httpx"
	"sample-app/internal/services/billing/application/invoices"
	"sample-app/internal/services/billing/domain/shared"
)

type Services struct {
	Invoices *invoices.Service
}

type invoiceResponse struct {
	ID       string  `json:"id"`
	OrderID  string  `json:"orderId"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	IssuedAt string  `json:"issuedAt"`
}

type invoiceHandler struct {
	svc *invoices.Service
}

func NewHandler(s Services) *http.ServeMux {
	mux := http.NewServeMux()
	httpx.Health(mux)
	h := invoiceHandler{svc: s.Invoices}
	mux.HandleFunc("GET /invoices", h.list)
	mux.HandleFunc("GET /invoices/{orderId}", h.byOrder)
	return mux
}

func (h invoiceHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	views, err := h.svc.List(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]invoiceResponse, 0, len(views))
	for _, v := range views {
		out = append(out, responseOf(v))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h invoiceHandler) byOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := shared.ParseID(r.PathValue("orderId"))
	if err != nil {
		writeError(w, err)
		return
	}
	view, err := h.svc.ByOrder(r.Context(), orderID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, responseOf(view))
}

func responseOf(v invoices.View) invoiceResponse {
	return invoiceResponse{ID: v.ID, OrderID: v.OrderID, Amount: float64(v.Amount) / 100, Currency: v.Currency, Status: v.Status,
		IssuedAt: v.IssuedAt.Format("2006-01-02T15:04:05Z07:00")}
}

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
