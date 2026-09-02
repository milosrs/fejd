package handler

import (
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SSEHandler struct {
	hub      *sse.Hub
	business *store.BusinessStore
}

func NewSSEHandler(hub *sse.Hub, business *store.BusinessStore) *SSEHandler {
	return &SSEHandler{hub: hub, business: business}
}

func (h *SSEHandler) StreamSlots(w http.ResponseWriter, r *http.Request) {
	businessSlug := chi.URLParam(r, "slug")

	b, err := h.business.GetBySlug(r.Context(), businessSlug)
	if err != nil {
		http.Error(w, "business not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	businessID := b.ID.String()
	ch := h.hub.Subscribe(businessID)
	defer h.hub.Unsubscribe(businessID, ch)

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: slots_updated\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
