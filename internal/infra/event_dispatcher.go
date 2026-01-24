package infra

import (
	"sync"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// EventDispatcher implements IEventDispatcher with synchronous event handling.
type EventDispatcher struct {
	handlers map[string][]domain.EventHandler
	mu       sync.RWMutex
}

// NewEventDispatcher creates a new EventDispatcher.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]domain.EventHandler),
	}
}

// Publish sends an event to all registered handlers for that event type.
func (d *EventDispatcher) Publish(event domain.Event) {
	d.mu.RLock()
	handlers := d.handlers[event.EventType()]
	d.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// Subscribe registers a handler for a specific event type.
func (d *EventDispatcher) Subscribe(eventType string, handler domain.EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}
