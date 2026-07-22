package sse

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestHubSubscribeUnsubscribe(t *testing.T) {
	hub := NewHub()

	ch := hub.Subscribe("business-1")
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	hub.Unsubscribe("business-1", ch)

	hub.mu.RLock()
	defer hub.mu.RUnlock()
	if _, ok := hub.subscribers["business-1"]; ok {
		t.Error("expected subscribers for business-1 to be removed")
	}
}

func TestHubPublish(t *testing.T) {
	hub := NewHub()
	ch := hub.Subscribe("business-1")

	event := map[string]string{"type": "slots_updated"}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case data := <-ch:
			var decoded map[string]string
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("failed to unmarshal: %v", err)
				return
			}
			if decoded["type"] != "slots_updated" {
				t.Errorf("expected type 'slots_updated', got '%s'", decoded["type"])
			}
		case <-time.After(1 * time.Second):
			t.Error("timed out waiting for event")
		}
	}()

	hub.Publish("business-1", event)
	wg.Wait()

	hub.Unsubscribe("business-1", ch)
}

func TestHubPublishToMultipleSubscribers(t *testing.T) {
	hub := NewHub()
	ch1 := hub.Subscribe("business-1")
	ch2 := hub.Subscribe("business-1")

	event := map[string]string{"type": "test"}

	var wg sync.WaitGroup
	wg.Add(2)

	receive := func(ch chan []byte) {
		defer wg.Done()
		select {
		case <-ch:
		case <-time.After(1 * time.Second):
			t.Error("timed out waiting for event")
		}
	}

	go receive(ch1)
	go receive(ch2)

	hub.Publish("business-1", event)
	wg.Wait()

	hub.Unsubscribe("business-1", ch1)
	hub.Unsubscribe("business-1", ch2)
}

func TestHubPublishDifferentBusinesses(t *testing.T) {
	hub := NewHub()
	ch1 := hub.Subscribe("business-1")
	ch2 := hub.Subscribe("business-2")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case <-ch1:
		case <-time.After(500 * time.Millisecond):
			t.Error("subscriber 1 should have received event")
		}
	}()

	hub.Publish("business-1", map[string]string{"type": "for_business_1"})
	wg.Wait()

	select {
	case <-ch2:
		t.Error("subscriber 2 should NOT have received event for business-1")
	case <-time.After(200 * time.Millisecond):
	}

	hub.Unsubscribe("business-1", ch1)
	hub.Unsubscribe("business-2", ch2)
}
