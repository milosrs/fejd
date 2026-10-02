package handler

import (
	"encoding/json"
	"fejd-backend/auth"
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

	// Realm-admin-created salons are hidden from non-realm-administrators, so
	// their slot stream must not leak to those callers either.
	if b.RealmAdminCreated && !auth.IsRealmAdmin(auth.GetClaimsFromRequest(r)) {
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
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", sseEventName(data), data)
			flusher.Flush()
		}
	}
}

// sseEventName derives the SSE event name from the message's "type" field so
// subscribers can listen to dedicated events (e.g. closures_updated) rather
// than a single generic stream. Falls back to slots_updated for legacy payloads.
func sseEventName(data []byte) string {
	var v struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &v); err == nil && v.Type != "" {
		return v.Type
	}
	return "slots_updated"
}
