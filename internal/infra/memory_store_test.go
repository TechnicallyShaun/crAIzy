package infra

import (
	"sync"
	"testing"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

func TestMemoryAgentStore_CRUD(t *testing.T) {
	t.Run("add and get", func(t *testing.T) {
		store := NewMemoryAgentStore()
		agent := &domain.Agent{ID: "test-1", Status: domain.AgentStatusActive}

		err := store.Add(agent)

		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		got := store.Get("test-1")
		if got == nil {
			t.Fatal("Get returned nil")
		}
		if got.ID != "test-1" {
			t.Errorf("ID = %q, want %q", got.ID, "test-1")
		}
	})

	t.Run("get nonexistent", func(t *testing.T) {
		store := NewMemoryAgentStore()

		got := store.Get("nonexistent")

		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("exists", func(t *testing.T) {
		store := NewMemoryAgentStore()
		store.Add(&domain.Agent{ID: "test-1"})

		if !store.Exists("test-1") {
			t.Error("Exists should return true for existing agent")
		}
		if store.Exists("nonexistent") {
			t.Error("Exists should return false for nonexistent agent")
		}
	})

	t.Run("remove", func(t *testing.T) {
		store := NewMemoryAgentStore()
		store.Add(&domain.Agent{ID: "test-1"})

		err := store.Remove("test-1")

		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}
		if store.Exists("test-1") {
			t.Error("agent should have been removed")
		}
	})

	t.Run("list", func(t *testing.T) {
		store := NewMemoryAgentStore()
		store.Add(&domain.Agent{ID: "a1"})
		store.Add(&domain.Agent{ID: "a2"})

		agents := store.List()

		if len(agents) != 2 {
			t.Errorf("List returned %d agents, want 2", len(agents))
		}
	})

	t.Run("update status", func(t *testing.T) {
		store := NewMemoryAgentStore()
		store.Add(&domain.Agent{ID: "test-1", Status: domain.AgentStatusActive})

		err := store.UpdateStatus("test-1", domain.AgentStatusTerminated)

		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
		agent := store.Get("test-1")
		if agent.Status != domain.AgentStatusTerminated {
			t.Errorf("status = %v, want %v", agent.Status, domain.AgentStatusTerminated)
		}
	})

	t.Run("update status nonexistent", func(t *testing.T) {
		store := NewMemoryAgentStore()

		// Should not panic or error
		err := store.UpdateStatus("nonexistent", domain.AgentStatusTerminated)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestMemoryAgentStore_Concurrency(t *testing.T) {
	store := NewMemoryAgentStore()
	var wg sync.WaitGroup
	n := 100

	// Concurrent writes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			agent := &domain.Agent{ID: string(rune('a' + id%26))}
			store.Add(agent)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.List()
		}()
	}

	wg.Wait()

	// Test passes if no race condition panics
}
