package handler

import (
	"fejd-backend/internal/sse"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SSEHandler struct {
	hub *sse.Hub
}

func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) StreamSlots(w http.ResponseWriter, r *http.Request) {
	businessSlug := chi.URLParam(r, "slug")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.hub.Subscribe(businessSlug)
	defer h.hub.Unsubscribe(businessSlug, ch)

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
