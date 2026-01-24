package infra

import (
	"sync"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// MemoryAgentStore implements IAgentStore with an in-memory map.
// This is suitable for MVP; a persistent store can be added later.
type MemoryAgentStore struct {
	agents map[string]*domain.Agent
	mu     sync.RWMutex
}

// NewMemoryAgentStore creates a new in-memory agent store.
func NewMemoryAgentStore() *MemoryAgentStore {
	return &MemoryAgentStore{
		agents: make(map[string]*domain.Agent),
	}
}

// Add stores a new agent.
func (s *MemoryAgentStore) Add(agent *domain.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[agent.ID] = agent
	return nil
}

// Remove deletes an agent by ID.
func (s *MemoryAgentStore) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, id)
	return nil
}

// List returns all stored agents.
func (s *MemoryAgentStore) List() []*domain.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*domain.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents
}

// Get retrieves an agent by ID.
func (s *MemoryAgentStore) Get(id string) *domain.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agents[id]
}

// Exists checks if an agent with the given ID exists.
func (s *MemoryAgentStore) Exists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.agents[id]
	return exists
}

// UpdateStatus updates the status of an agent.
func (s *MemoryAgentStore) UpdateStatus(id string, status domain.AgentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if agent, exists := s.agents[id]; exists {
		agent.Status = status
	}
	return nil
}
