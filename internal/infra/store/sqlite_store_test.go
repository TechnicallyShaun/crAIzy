package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

func createTestStore(t *testing.T) (*SQLiteAgentStore, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "craizy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestSQLiteAgentStore_Add(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "craizy-test-claude-worker1",
		Project:   "test",
		AgentType: "claude",
		Name:      "worker1",
		Command:   "echo hello",
		WorkDir:   "/tmp",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}

	err := store.Add(agent)
	if err != nil {
		t.Fatalf("failed to add agent: %v", err)
	}

	// Verify it was stored
	retrieved := store.Get(agent.ID)
	if retrieved == nil {
		t.Fatal("expected to retrieve agent")
	}
	if retrieved.ID != agent.ID {
		t.Errorf("expected ID %q, got %q", agent.ID, retrieved.ID)
	}
	if retrieved.Project != agent.Project {
		t.Errorf("expected Project %q, got %q", agent.Project, retrieved.Project)
	}
	if retrieved.Status != domain.AgentStatusActive {
		t.Errorf("expected status 'active', got %q", retrieved.Status)
	}
}

func TestSQLiteAgentStore_AddDuplicate(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "craizy-test-claude-worker1",
		Project:   "test",
		AgentType: "claude",
		Name:      "worker1",
		Command:   "echo hello",
		WorkDir:   "/tmp",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}

	err := store.Add(agent)
	if err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	// Adding duplicate should fail (primary key constraint)
	err = store.Add(agent)
	if err == nil {
		t.Error("expected error when adding duplicate agent")
	}
}

func TestSQLiteAgentStore_Remove(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "craizy-test-claude-worker1",
		Project:   "test",
		AgentType: "claude",
		Name:      "worker1",
		Command:   "echo hello",
		WorkDir:   "/tmp",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}

	_ = store.Add(agent)

	err := store.Remove(agent.ID)
	if err != nil {
		t.Fatalf("failed to remove agent: %v", err)
	}

	// Verify it was removed
	if store.Exists(agent.ID) {
		t.Error("agent should not exist after removal")
	}
}

func TestSQLiteAgentStore_List(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	// Add multiple agents
	agents := []*domain.Agent{
		{ID: "agent-1", Project: "proj1", AgentType: "claude", Name: "a1", Command: "cmd1", WorkDir: "/", Status: domain.AgentStatusActive, CreatedAt: time.Now()},
		{ID: "agent-2", Project: "proj2", AgentType: "aider", Name: "a2", Command: "cmd2", WorkDir: "/", Status: domain.AgentStatusActive, CreatedAt: time.Now()},
		{ID: "agent-3", Project: "proj1", AgentType: "claude", Name: "a3", Command: "cmd3", WorkDir: "/", Status: domain.AgentStatusTerminated, CreatedAt: time.Now()},
	}

	for _, a := range agents {
		_ = store.Add(a)
	}

	list := store.List()
	if len(list) != 3 {
		t.Errorf("expected 3 agents, got %d", len(list))
	}
}

func TestSQLiteAgentStore_Get(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "craizy-test-claude-worker1",
		Project:   "test",
		AgentType: "claude",
		Name:      "worker1",
		Command:   "echo hello",
		WorkDir:   "/tmp",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}
	_ = store.Add(agent)

	retrieved := store.Get(agent.ID)
	if retrieved == nil {
		t.Fatal("expected to get agent")
	}
	if retrieved.AgentType != "claude" {
		t.Errorf("expected AgentType 'claude', got %q", retrieved.AgentType)
	}
}

func TestSQLiteAgentStore_GetNonExistent(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	retrieved := store.Get("non-existent-id")
	if retrieved != nil {
		t.Error("expected nil for non-existent agent")
	}
}

func TestSQLiteAgentStore_Exists(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	if store.Exists("test-agent") {
		t.Error("agent should not exist initially")
	}

	_ = store.Add(&domain.Agent{
		ID:        "test-agent",
		Project:   "test",
		AgentType: "claude",
		Name:      "test",
		Command:   "cmd",
		WorkDir:   "/",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	})

	if !store.Exists("test-agent") {
		t.Error("agent should exist after adding")
	}
}

func TestSQLiteAgentStore_UpdateStatus(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "test-agent",
		Project:   "test",
		AgentType: "claude",
		Name:      "test",
		Command:   "cmd",
		WorkDir:   "/",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}
	_ = store.Add(agent)

	err := store.UpdateStatus(agent.ID, domain.AgentStatusTerminated)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	retrieved := store.Get(agent.ID)
	if retrieved.Status != domain.AgentStatusTerminated {
		t.Errorf("expected status 'terminated', got %q", retrieved.Status)
	}
	if retrieved.TerminatedAt == nil {
		t.Error("expected TerminatedAt to be set when status is terminated")
	}
}

func TestSQLiteAgentStore_UpdateStatusToActive(t *testing.T) {
	store, cleanup := createTestStore(t)
	defer cleanup()

	agent := &domain.Agent{
		ID:        "test-agent",
		Project:   "test",
		AgentType: "claude",
		Name:      "test",
		Command:   "cmd",
		WorkDir:   "/",
		Status:    domain.AgentStatusTerminated,
		CreatedAt: time.Now(),
	}
	_ = store.Add(agent)

	// Update back to active - terminated_at should be cleared
	err := store.UpdateStatus(agent.ID, domain.AgentStatusActive)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	retrieved := store.Get(agent.ID)
	if retrieved.Status != domain.AgentStatusActive {
		t.Errorf("expected status 'active', got %q", retrieved.Status)
	}
}

func TestSQLiteAgentStore_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "craizy-persistence-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create store and add agent
	store1, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	agent := &domain.Agent{
		ID:        "persistent-agent",
		Project:   "test",
		AgentType: "claude",
		Name:      "persist",
		Command:   "echo persist",
		WorkDir:   "/tmp",
		Status:    domain.AgentStatusActive,
		CreatedAt: time.Now(),
	}
	_ = store1.Add(agent)
	store1.Close()

	// Reopen store and verify agent persisted
	store2, err := NewSQLiteAgentStore(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen store: %v", err)
	}
	defer store2.Close()

	retrieved := store2.Get("persistent-agent")
	if retrieved == nil {
		t.Fatal("agent should persist across store reopens")
	}
	if retrieved.Name != "persist" {
		t.Errorf("expected Name 'persist', got %q", retrieved.Name)
	}
}
