package sse

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[chan []byte]struct{}),
	}
}

func (h *Hub) Subscribe(businessID string) chan []byte {
	ch := make(chan []byte, 64)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[businessID]; !ok {
		h.subscribers[businessID] = make(map[chan []byte]struct{})
	}
	h.subscribers[businessID][ch] = struct{}{}

	return ch
}

func (h *Hub) Unsubscribe(businessID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.subscribers[businessID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(h.subscribers, businessID)
		}
	}
}

func (h *Hub) Publish(businessID string, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("SSE hub: failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.subscribers[businessID]; ok {
		for ch := range subs {
			select {
			case ch <- data:
			default:
			}
		}
	}
}
