package infra

import (
	"sync"
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

type testEvent struct {
	eventType string
	timestamp time.Time
}

func (e testEvent) EventType() string     { return e.eventType }
func (e testEvent) OccurredAt() time.Time { return e.timestamp }

func TestEventDispatcher_SubscribePublish(t *testing.T) {
	t.Run("single handler", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		var received domain.Event

		dispatcher.Subscribe("test.event", func(e domain.Event) {
			received = e
		})

		event := testEvent{eventType: "test.event", timestamp: time.Now()}
		dispatcher.Publish(event)

		if received == nil {
			t.Fatal("handler not called")
		}
		if received.EventType() != "test.event" {
			t.Errorf("event type = %q, want %q", received.EventType(), "test.event")
		}
	})

	t.Run("multiple handlers same event", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		callCount := 0

		dispatcher.Subscribe("test.event", func(e domain.Event) { callCount++ })
		dispatcher.Subscribe("test.event", func(e domain.Event) { callCount++ })

		dispatcher.Publish(testEvent{eventType: "test.event"})

		if callCount != 2 {
			t.Errorf("call count = %d, want 2", callCount)
		}
	})

	t.Run("no handlers for event type", func(t *testing.T) {
		dispatcher := NewEventDispatcher()

		// Should not panic
		dispatcher.Publish(testEvent{eventType: "unsubscribed.event"})
	})

	t.Run("different event types isolated", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		var calledA, calledB bool

		dispatcher.Subscribe("type.a", func(e domain.Event) { calledA = true })
		dispatcher.Subscribe("type.b", func(e domain.Event) { calledB = true })

		dispatcher.Publish(testEvent{eventType: "type.a"})

		if !calledA {
			t.Error("handler A should have been called")
		}
		if calledB {
			t.Error("handler B should not have been called")
		}
	})
}

func TestEventDispatcher_Concurrency(t *testing.T) {
	dispatcher := NewEventDispatcher()
	var wg sync.WaitGroup
	n := 100

	// Concurrent subscribes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			dispatcher.Subscribe("test.event", func(e domain.Event) {})
		}(i)
	}

	// Concurrent publishes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dispatcher.Publish(testEvent{eventType: "test.event"})
		}()
	}

	wg.Wait()
	// Test passes if no race condition panics
}
