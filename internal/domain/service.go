package domain

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AgentService orchestrates agent operations using the tmux client and store.
type AgentService struct {
	tmux       ITmuxClient
	store      IAgentStore
	dispatcher IEventDispatcher
	project    string
	workDir    string
}

// NewAgentService creates a new AgentService with the given dependencies.
func NewAgentService(tmux ITmuxClient, store IAgentStore, dispatcher IEventDispatcher, project, workDir string) *AgentService {
	return &AgentService{
		tmux:       tmux,
		store:      store,
		dispatcher: dispatcher,
		project:    project,
		workDir:    workDir,
	}
}

// Create spawns a new agent session and stores it.
func (s *AgentService) Create(agentType, name, command string) (*Agent, error) {
	sessionID := BuildSessionID(s.project, agentType, name)

	// Check if an active session already exists
	existing := s.store.Get(sessionID)
	if existing != nil && existing.Status == AgentStatusActive {
		return nil, fmt.Errorf("agent session %q already exists", sessionID)
	}

	// Remove any terminated agent with same ID before creating new one
	if existing != nil {
		_ = s.store.Remove(sessionID)
	}

	agent := &Agent{
		ID:        sessionID,
		Project:   s.project,
		AgentType: agentType,
		Name:      name,
		Command:   command,
		WorkDir:   s.workDir,
		Status:    AgentStatusActive,
		CreatedAt: time.Now(),
	}

	// Publish event - adapters will create tmux session and store agent
	s.dispatcher.Publish(AgentCreated{
		Agent:     agent,
		Timestamp: time.Now(),
	})

	return agent, nil
}

// Kill terminates an agent session.
func (s *AgentService) Kill(sessionID string) error {
	// Publish event - adapters will kill tmux session and update status
	s.dispatcher.Publish(AgentKilled{
		AgentID:   sessionID,
		Timestamp: time.Now(),
	})

	return nil
}

// List returns active agents for the current project.
func (s *AgentService) List() []*Agent {
	all := s.store.List()
	var active []*Agent
	for _, agent := range all {
		if agent.Project == s.project && agent.Status == AgentStatusActive {
			active = append(active, agent)
		}
	}
	return active
}

// Attach returns a tea.Cmd that attaches to the given session.
// This will suspend the TUI and take over the terminal.
func (s *AgentService) Attach(sessionID string) tea.Cmd {
	cmd := s.tmux.AttachCmd(sessionID)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return AgentDetachedMsg{SessionID: sessionID, Err: err}
	})
}

// Exists checks if an agent exists in the store.
func (s *AgentService) Exists(sessionID string) bool {
	return s.store.Exists(sessionID)
}

// CaptureOutput captures the last N lines from an agent's tmux pane.
func (s *AgentService) CaptureOutput(sessionID string, lines int) (string, error) {
	return s.tmux.CapturePaneOutput(sessionID, lines)
}

// Reconcile synchronizes the store with actual tmux sessions.
// It marks agents as terminated if their tmux session no longer exists,
// and kills orphaned tmux sessions that aren't in the store.
func (s *AgentService) Reconcile() error {
	// Get all stored agents
	agents := s.store.List()

	// Check for orphaned store entries (session doesn't exist in tmux)
	for _, agent := range agents {
		if agent.Status == AgentStatusTerminated {
			continue
		}
		if !s.tmux.SessionExists(agent.ID) {
			// Mark as terminated rather than removing
			_ = s.store.UpdateStatus(agent.ID, AgentStatusTerminated)
		}
	}

	// Get all tmux sessions
	sessions, err := s.tmux.ListSessions()
	if err != nil {
		// tmux might not be running, which is fine
		return nil
	}

	// Check for orphaned tmux sessions (matches our prefix but not in store)
	prefix := "craizy-" + SanitizeName(s.project) + "-"
	for _, session := range sessions {
		if strings.HasPrefix(session, prefix) {
			if !s.store.Exists(session) {
				_ = s.tmux.KillSession(session)
			}
		}
	}

	return nil
}

// AgentDetachedMsg is sent when returning from an attached tmux session.
type AgentDetachedMsg struct {
	SessionID string
	Err       error
}
