package httpapi

import (
	"math"
	"net/http"
	"strconv"

	"sample-app/internal/services/orders/application/orders"
	"sample-app/internal/services/orders/domain/shared"
)

type createOrderRequest struct {
	CustomerName string `json:"customerName"`
}

type addItemRequest struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Currency    string  `json:"currency"`
}

type itemResponse struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Subtotal    float64 `json:"subtotal"`
}

type orderResponse struct {
	ID           string         `json:"id"`
	CustomerName string         `json:"customerName"`
	Status       string         `json:"status"`
	CreatedAt    string         `json:"createdAt"`
	Total        float64        `json:"total"`
	Currency     string         `json:"currency"`
	Items        []itemResponse `json:"items"`
}

type orderHandler struct {
	svc *orders.Service
}

func registerOrders(mux *http.ServeMux, svc *orders.Service) {
	h := orderHandler{svc: svc}
	mux.HandleFunc("POST /orders", h.create)
	mux.HandleFunc("POST /orders/{id}/items", h.addItem)
	mux.HandleFunc("POST /orders/{id}/place", h.place)
	mux.HandleFunc("GET /orders/{id}", h.get)
	mux.HandleFunc("GET /orders", h.list)
}

func (h orderHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if !decode(w, r, &req) {
		return
	}
	id, err := h.svc.Create(r.Context(), req.CustomerName)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Location", "/orders/"+string(id))
	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

func (h orderHandler) addItem(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req addItemRequest
	if !decode(w, r, &req) {
		return
	}
	in := orders.AddItem{ProductName: req.ProductName, Quantity: req.Quantity, UnitPrice: toCents(req.UnitPrice), Currency: req.Currency}
	if err := h.svc.AddItem(r.Context(), id, in); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h orderHandler) place(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Place(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h orderHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	view, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, responseOf(view))
}

func (h orderHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	views, err := h.svc.List(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]orderResponse, 0, len(views))
	for _, v := range views {
		out = append(out, responseOf(v))
	}
	writeJSON(w, http.StatusOK, out)
}

func responseOf(v orders.OrderView) orderResponse {
	out := orderResponse{ID: v.ID, CustomerName: v.CustomerName, Status: v.Status, CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Total: fromCents(v.Total), Currency: v.Currency, Items: []itemResponse{}}
	for _, item := range v.Items {
		out.Items = append(out.Items, itemResponse{ProductName: item.ProductName, Quantity: item.Quantity,
			UnitPrice: fromCents(item.UnitPrice), Subtotal: fromCents(item.Subtotal)})
	}
	return out
}

func toCents(v float64) int64   { return int64(math.Round(v * 100)) }
func fromCents(c int64) float64 { return float64(c) / 100 }
