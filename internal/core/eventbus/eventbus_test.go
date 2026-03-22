package eventbus

import (
	"testing"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

func TestPublishSubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	received := make(chan models.Event, 1)
	bus.Subscribe(models.EventDeviceDiscovered, func(e models.Event) {
		received <- e
	})

	bus.Publish(models.Event{
		Type:      models.EventDeviceDiscovered,
		Payload:   "test-device",
		Timestamp: time.Now(),
	})

	select {
	case e := <-received:
		if e.Type != models.EventDeviceDiscovered {
			t.Errorf("expected %s, got %s", models.EventDeviceDiscovered, e.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := New()
	defer bus.Close()

	count := make(chan struct{}, 2)
	handler := func(e models.Event) { count <- struct{}{} }

	bus.Subscribe(models.EventDeviceDiscovered, handler)
	bus.Subscribe(models.EventDeviceDiscovered, handler)

	bus.Publish(models.Event{Type: models.EventDeviceDiscovered, Timestamp: time.Now()})

	for i := 0; i < 2; i++ {
		select {
		case <-count:
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for subscriber %d", i+1)
		}
	}
}

func TestUnsubscribedEventIgnored(t *testing.T) {
	bus := New()
	defer bus.Close()

	called := false
	bus.Subscribe(models.EventDeviceDiscovered, func(e models.Event) {
		called = true
	})

	bus.Publish(models.Event{Type: models.EventDeviceLost, Timestamp: time.Now()})
	time.Sleep(50 * time.Millisecond)

	if called {
		t.Error("handler should not have been called for unsubscribed event")
	}
}
