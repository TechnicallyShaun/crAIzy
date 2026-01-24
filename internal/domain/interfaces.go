package domain

import "os/exec"

// ITmuxClient defines the interface for tmux operations.
type ITmuxClient interface {
	// CreateSession creates a new detached tmux session.
	CreateSession(id, command, workDir string) error

	// KillSession terminates a tmux session.
	KillSession(id string) error

	// ListSessions returns all tmux session names.
	ListSessions() ([]string, error)

	// AttachCmd returns an exec.Cmd that can be used to attach to a session.
	AttachCmd(id string) *exec.Cmd

	// SessionExists checks if a tmux session exists.
	SessionExists(id string) bool
}

// IAgentStore defines the interface for agent persistence.
type IAgentStore interface {
	// Add stores a new agent.
	Add(agent *Agent) error

	// Remove deletes an agent by ID.
	Remove(id string) error

	// List returns all stored agents.
	List() []*Agent

	// Get retrieves an agent by ID.
	Get(id string) *Agent

	// Exists checks if an agent with the given ID exists.
	Exists(id string) bool

	// UpdateStatus updates the status of an agent.
	UpdateStatus(id string, status AgentStatus) error
}
