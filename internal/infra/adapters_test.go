package infra

import (
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// mockTmuxClient is a test double for ITmuxClient
type mockTmuxClient struct {
	sessions        map[string]bool
	createErr       error
	killErr         error
	createCallCount int
	killCallCount   int
}

func newMockTmux() *mockTmuxClient {
	return &mockTmuxClient{sessions: make(map[string]bool)}
}

func (m *mockTmuxClient) CreateSession(id, command, workDir string) error {
	m.createCallCount++
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[id] = true
	return nil
}

func (m *mockTmuxClient) KillSession(id string) error {
	m.killCallCount++
	if m.killErr != nil {
		return m.killErr
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockTmuxClient) ListSessions() ([]string, error) {
	var sessions []string
	for id := range m.sessions {
		sessions = append(sessions, id)
	}
	return sessions, nil
}

func (m *mockTmuxClient) AttachCmd(id string) *exec.Cmd {
	return exec.Command("echo", "attach", id)
}

func (m *mockTmuxClient) SessionExists(id string) bool {
	return m.sessions[id]
}

func (m *mockTmuxClient) CapturePaneOutput(sessionID string, lines int) (string, error) {
	return "mock output", nil
}

func (m *mockTmuxClient) SendKeys(sessionID, text string) error {
	return nil
}

func TestWireAdapters_AgentCreated(t *testing.T) {
	t.Run("creates tmux session and stores agent", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		store := NewMemoryAgentStore()
		tmux := newMockTmux()

		WireAdapters(dispatcher, store, tmux, nil)

		agent := &domain.Agent{
			ID:        "test-agent",
			Project:   "test",
			AgentType: "claude",
			Name:      "worker",
			Command:   "echo hello",
			WorkDir:   "/tmp",
			Status:    domain.AgentStatusActive,
			CreatedAt: time.Now(),
		}

		dispatcher.Publish(domain.AgentCreated{
			Agent:     agent,
			Timestamp: time.Now(),
		})

		// Verify tmux session was created
		if !tmux.sessions["test-agent"] {
			t.Error("tmux session should have been created")
		}

		// Verify agent was stored
		if !store.Exists("test-agent") {
			t.Error("agent should have been stored")
		}
	})

	t.Run("does not store if tmux creation fails", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		store := NewMemoryAgentStore()
		tmux := newMockTmux()
		tmux.createErr = errors.New("tmux error")

		WireAdapters(dispatcher, store, tmux, nil)

		agent := &domain.Agent{
			ID:        "test-agent",
			Project:   "test",
			Status:    domain.AgentStatusActive,
			CreatedAt: time.Now(),
		}

		dispatcher.Publish(domain.AgentCreated{
			Agent:     agent,
			Timestamp: time.Now(),
		})

		// Agent should not be stored
		if store.Exists("test-agent") {
			t.Error("agent should not be stored when tmux fails")
		}
	})
}

func TestWireAdapters_AgentKilled(t *testing.T) {
	t.Run("kills tmux session and updates status", func(t *testing.T) {
		dispatcher := NewEventDispatcher()
		store := NewMemoryAgentStore()
		tmux := newMockTmux()

		WireAdapters(dispatcher, store, tmux, nil)

		// Pre-populate store and tmux
		agent := &domain.Agent{
			ID:        "test-agent",
			Status:    domain.AgentStatusActive,
			CreatedAt: time.Now(),
		}
		store.Add(agent)
		tmux.sessions["test-agent"] = true

		dispatcher.Publish(domain.AgentKilled{
			AgentID:   "test-agent",
			Timestamp: time.Now(),
		})

		// Verify tmux session was killed
		if tmux.sessions["test-agent"] {
			t.Error("tmux session should have been killed")
		}

		// Verify status was updated
		storedAgent := store.Get("test-agent")
		if storedAgent.Status != domain.AgentStatusTerminated {
			t.Errorf("status = %v, want terminated", storedAgent.Status)
		}
	})
}
