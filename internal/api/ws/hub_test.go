package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create a mock client channel
	ch := make(chan []byte, 10)
	client := &Client{send: ch, hub: hub}
	hub.Register(client)

	// Give registration time to process
	time.Sleep(50 * time.Millisecond)

	event := models.Event{
		Type:      models.EventDeviceDiscovered,
		Payload:   map[string]string{"id": "test-1"},
		Timestamp: time.Now(),
	}
	hub.Broadcast(event)

	select {
	case msg := <-ch:
		var got models.Event
		if err := json.Unmarshal(msg, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Type != models.EventDeviceDiscovered {
			t.Errorf("expected %s, got %s", models.EventDeviceDiscovered, got.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}
