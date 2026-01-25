package domain

import "time"

// Event represents a domain event that can be published and subscribed to.
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// EventHandler is a function that handles domain events.
type EventHandler func(event Event)

// IEventDispatcher defines the interface for publishing and subscribing to events.
type IEventDispatcher interface {
	// Publish sends an event to all registered handlers.
	Publish(event Event)

	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType string, handler EventHandler)
}

// AgentCreated is published when a new agent is created.
type AgentCreated struct {
	Agent     *Agent
	Timestamp time.Time
}

func (e AgentCreated) EventType() string     { return "agent.created" }
func (e AgentCreated) OccurredAt() time.Time { return e.Timestamp }

// AgentKilled is published when an agent is terminated.
type AgentKilled struct {
	AgentID   string
	Timestamp time.Time
}

func (e AgentKilled) EventType() string     { return "agent.killed" }
func (e AgentKilled) OccurredAt() time.Time { return e.Timestamp }

// AgentStatusChanged is published when an agent's status changes.
type AgentStatusChanged struct {
	AgentID   string
	OldStatus AgentStatus
	NewStatus AgentStatus
	Timestamp time.Time
}

func (e AgentStatusChanged) EventType() string     { return "agent.status_changed" }
func (e AgentStatusChanged) OccurredAt() time.Time { return e.Timestamp }
