package eventbus

import (
	"sync"

	"github.com/netmap/netmap/internal/core/models"
)

type Handler func(models.Event)

type EventBus struct {
	mu       sync.RWMutex
	wg       sync.WaitGroup
	handlers map[models.EventType][]Handler
	closed   bool
}

func New() *EventBus {
	return &EventBus{
		handlers: make(map[models.EventType][]Handler),
	}
}

func (b *EventBus) Subscribe(eventType models.EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *EventBus) Publish(event models.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	for _, handler := range b.handlers[event.Type] {
		h := handler
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			h(event)
		}()
	}
}

func (b *EventBus) Close() {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
	b.wg.Wait() // wait for all in-flight handlers to finish
}
