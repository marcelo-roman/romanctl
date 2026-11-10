package httpapi

import (
	"net/http"
	"strconv"

	"sample-app/internal/application/sampleitems"
	"sample-app/internal/domain/shared"
)

type sampleItemRequest struct {
	Name string `json:"name"`
}

type sampleItemResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type sampleItemHandler struct {
	svc *sampleitems.Service
}

func registerSampleItems(mux *http.ServeMux, svc *sampleitems.Service) {
	h := sampleItemHandler{svc: svc}
	mux.HandleFunc("POST /sample-items", h.create)
	mux.HandleFunc("GET /sample-items", h.list)
	mux.HandleFunc("GET /sample-items/{id}", h.get)
	mux.HandleFunc("PUT /sample-items/{id}", h.rename)
	mux.HandleFunc("DELETE /sample-items/{id}", h.delete)
}

func (h sampleItemHandler) create(w http.ResponseWriter, r *http.Request) {
	var req sampleItemRequest
	if !decode(w, r, &req) {
		return
	}
	id, err := h.svc.Create(r.Context(), req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Location", "/sample-items/"+string(id))
	writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
}

func (h sampleItemHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	views, err := h.svc.List(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]sampleItemResponse, 0, len(views))
	for _, v := range views {
		out = append(out, sampleItemResponseOf(v))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h sampleItemHandler) get(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, sampleItemResponseOf(view))
}

func (h sampleItemHandler) rename(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req sampleItemRequest
	if !decode(w, r, &req) {
		return
	}
	if err := h.svc.Rename(r.Context(), id, req.Name); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h sampleItemHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sampleItemResponseOf(v sampleitems.View) sampleItemResponse {
	return sampleItemResponse{ID: v.ID, Name: v.Name, CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}
