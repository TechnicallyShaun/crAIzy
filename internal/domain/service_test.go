package domain

import (
	"os/exec"
	"testing"
)

// Mock implementations

type mockTmuxClient struct {
	sessions       map[string]bool
	createErr      error
	killErr        error
	listErr        error
	capturedOutput string
	captureErr     error
}

func (m *mockTmuxClient) CreateSession(id, command, workDir string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[id] = true
	return nil
}

func (m *mockTmuxClient) KillSession(id string) error {
	if m.killErr != nil {
		return m.killErr
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockTmuxClient) ListSessions() ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var sessions []string
	for id := range m.sessions {
		sessions = append(sessions, id)
	}
	return sessions, nil
}

func (m *mockTmuxClient) AttachCmd(id string) *exec.Cmd {
	return exec.Command("echo", "attached")
}

func (m *mockTmuxClient) SessionExists(id string) bool {
	_, exists := m.sessions[id]
	return exists
}

func (m *mockTmuxClient) CapturePaneOutput(sessionID string, lines int) (string, error) {
	return m.capturedOutput, m.captureErr
}

type mockDispatcher struct {
	published []Event
}

func (m *mockDispatcher) Publish(event Event) {
	m.published = append(m.published, event)
}

func (m *mockDispatcher) Subscribe(eventType string, handler EventHandler) {}

// Tests

func TestAgentService_Create(t *testing.T) {
	t.Run("new agent", func(t *testing.T) {
		// Path 1: Create new agent - no existing
		store := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "testproj", "/tmp")

		agent, err := svc.Create("claude", "task1", "echo hello")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if agent.Status != AgentStatusActive {
			t.Errorf("status = %v, want %v", agent.Status, AgentStatusActive)
		}
		if len(dispatcher.published) != 1 {
			t.Errorf("published %d events, want 1", len(dispatcher.published))
		}
	})

	t.Run("duplicate active agent", func(t *testing.T) {
		// Path 2: Agent exists and is active - error
		store := newTestStore()
		existing := &Agent{
			ID:     "craizy-testproj-claude-task1",
			Status: AgentStatusActive,
		}
		store.Add(existing)

		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "testproj", "/tmp")

		_, err := svc.Create("claude", "task1", "echo hello")

		if err == nil {
			t.Fatal("expected error for duplicate active agent")
		}
	})

	t.Run("replace terminated agent", func(t *testing.T) {
		// Path 3: Agent exists but terminated - replace
		store := newTestStore()
		existing := &Agent{
			ID:     "craizy-testproj-claude-task1",
			Status: AgentStatusTerminated,
		}
		store.Add(existing)

		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "testproj", "/tmp")

		agent, err := svc.Create("claude", "task1", "echo hello")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if agent.Status != AgentStatusActive {
			t.Errorf("status = %v, want %v", agent.Status, AgentStatusActive)
		}
	})
}

func TestAgentService_List(t *testing.T) {
	t.Run("filter by project and status", func(t *testing.T) {
		store := newTestStore()
		// Add agents with different projects/statuses
		store.Add(&Agent{ID: "a1", Project: "proj1", Status: AgentStatusActive})
		store.Add(&Agent{ID: "a2", Project: "proj1", Status: AgentStatusTerminated})
		store.Add(&Agent{ID: "a3", Project: "proj2", Status: AgentStatusActive})

		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj1", "/tmp")

		agents := svc.List()

		// Path: Filter active agents for current project only
		if len(agents) != 1 {
			t.Errorf("got %d agents, want 1", len(agents))
		}
		if len(agents) > 0 && agents[0].ID != "a1" {
			t.Errorf("got agent %s, want a1", agents[0].ID)
		}
	})
}

func TestAgentService_Reconcile(t *testing.T) {
	t.Run("mark orphaned store entries", func(t *testing.T) {
		// Path 1: Agent in store but session doesn't exist in tmux
		store := newTestStore()
		store.Add(&Agent{ID: "craizy-proj-claude-task1", Project: "proj", Status: AgentStatusActive})

		tmux := &mockTmuxClient{sessions: make(map[string]bool)} // No sessions
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj", "/tmp")

		err := svc.Reconcile()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		agent := store.Get("craizy-proj-claude-task1")
		if agent.Status != AgentStatusTerminated {
			t.Errorf("status = %v, want %v", agent.Status, AgentStatusTerminated)
		}
	})

	t.Run("skip terminated agents", func(t *testing.T) {
		// Path 2: Already terminated - skip
		store := newTestStore()
		store.Add(&Agent{ID: "craizy-proj-claude-task1", Project: "proj", Status: AgentStatusTerminated})

		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj", "/tmp")

		err := svc.Reconcile()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should remain terminated
		agent := store.Get("craizy-proj-claude-task1")
		if agent.Status != AgentStatusTerminated {
			t.Errorf("status = %v, want %v", agent.Status, AgentStatusTerminated)
		}
	})

	t.Run("kill orphaned tmux sessions", func(t *testing.T) {
		// Path 3: Session exists in tmux but not in store
		store := newTestStore()
		tmux := &mockTmuxClient{
			sessions: map[string]bool{
				"craizy-proj-claude-orphan": true,
			},
		}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj", "/tmp")

		err := svc.Reconcile()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tmux.SessionExists("craizy-proj-claude-orphan") {
			t.Error("orphaned session should have been killed")
		}
	})

	t.Run("handle tmux not running", func(t *testing.T) {
		// Path 4: ListSessions returns error - graceful handling
		store := newTestStore()
		tmux := &mockTmuxClient{
			sessions: make(map[string]bool),
			listErr:  exec.ErrNotFound,
		}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj", "/tmp")

		err := svc.Reconcile()

		// Should return nil, not error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}

func TestAgentService_Kill(t *testing.T) {
	t.Run("publishes event", func(t *testing.T) {
		store := newTestStore()
		tmux := &mockTmuxClient{sessions: make(map[string]bool)}
		dispatcher := &mockDispatcher{}
		svc := NewAgentService(tmux, store, dispatcher, "proj", "/tmp")

		_ = svc.Kill("some-session")

		if len(dispatcher.published) != 1 {
			t.Errorf("published %d events, want 1", len(dispatcher.published))
		}
		if _, ok := dispatcher.published[0].(AgentKilled); !ok {
			t.Errorf("wrong event type: %T", dispatcher.published[0])
		}
	})
}

// Helper to create test store
func newTestStore() *testStore {
	return &testStore{agents: make(map[string]*Agent)}
}

type testStore struct {
	agents map[string]*Agent
}

func (s *testStore) Add(agent *Agent) error {
	s.agents[agent.ID] = agent
	return nil
}

func (s *testStore) Remove(id string) error {
	delete(s.agents, id)
	return nil
}

func (s *testStore) List() []*Agent {
	var agents []*Agent
	for _, a := range s.agents {
		agents = append(agents, a)
	}
	return agents
}

func (s *testStore) Get(id string) *Agent {
	return s.agents[id]
}

func (s *testStore) Exists(id string) bool {
	_, exists := s.agents[id]
	return exists
}

func (s *testStore) UpdateStatus(id string, status AgentStatus) error {
	if a, exists := s.agents[id]; exists {
		a.Status = status
	}
	return nil
}
