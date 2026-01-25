package domain

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// WorktreesDir is the directory under .craizy where worktrees are created.
	WorktreesDir = ".craizy/worktrees"
)

// AgentService orchestrates agent operations using the tmux client and store.
type AgentService struct {
	tmux       ITmuxClient
	store      IAgentStore
	dispatcher IEventDispatcher
	git        IGitClient
	project    string
	workDir    string
}

// NewAgentService creates a new AgentService with the given dependencies.
func NewAgentService(tmux ITmuxClient, store IAgentStore, dispatcher IEventDispatcher, git IGitClient, project, workDir string) *AgentService {
	return &AgentService{
		tmux:       tmux,
		store:      store,
		dispatcher: dispatcher,
		git:        git,
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

	// Build branch name from session ID
	branchName := sessionID

	// Check if branch already exists
	if s.git != nil && s.git.BranchExists(branchName) {
		return nil, fmt.Errorf("branch %q already exists", branchName)
	}

	// Get current branch as base
	var baseBranch string
	var worktreePath string
	if s.git != nil {
		var err error
		baseBranch, err = s.git.CurrentBranch(s.workDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}

		// Create worktree path
		worktreePath = filepath.Join(s.workDir, WorktreesDir, SanitizeName(name))

		// Create worktree with new branch
		if err := s.git.CreateWorktree(worktreePath, branchName, baseBranch); err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	// Set agent work directory to worktree if created, otherwise use main workDir
	agentWorkDir := s.workDir
	if worktreePath != "" {
		agentWorkDir = worktreePath
	}

	agent := &Agent{
		ID:         sessionID,
		Project:    s.project,
		AgentType:  agentType,
		Name:       name,
		Command:    command,
		WorkDir:    agentWorkDir,
		Status:     AgentStatusActive,
		CreatedAt:  time.Now(),
		Branch:     branchName,
		BaseBranch: baseBranch,
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

// CheckKill checks if an agent has uncommitted changes before killing.
// Returns true if there are uncommitted changes that need user confirmation.
func (s *AgentService) CheckKill(sessionID string) (hasUncommitted bool, err error) {
	if s.git == nil {
		return false, nil
	}

	agent := s.store.Get(sessionID)
	if agent == nil {
		return false, fmt.Errorf("agent %q not found", sessionID)
	}

	if agent.Branch == "" {
		return false, nil
	}

	return s.git.HasUncommittedChanges(agent.WorkDir), nil
}

// ForceKill terminates an agent, optionally discarding uncommitted changes.
func (s *AgentService) ForceKill(sessionID string, discardChanges bool) error {
	if s.git != nil && !discardChanges {
		agent := s.store.Get(sessionID)
		if agent != nil && agent.Branch != "" && s.git.HasUncommittedChanges(agent.WorkDir) {
			// Stash changes before killing
			_ = s.git.Stash(agent.WorkDir)
		}
	}

	return s.Kill(sessionID)
}

// MergeResult contains the result of a merge operation.
type MergeResult struct {
	Success     bool
	Stashed     bool
	ConflictErr error
}

// MergeAgent merges an agent's branch into the base branch.
// If there are uncommitted changes in the main workdir, they are stashed first.
func (s *AgentService) MergeAgent(sessionID string) (*MergeResult, error) {
	if s.git == nil {
		return nil, fmt.Errorf("git client not available")
	}

	agent := s.store.Get(sessionID)
	if agent == nil {
		return nil, fmt.Errorf("agent %q not found", sessionID)
	}

	if agent.Branch == "" {
		return nil, fmt.Errorf("agent has no branch to merge")
	}

	result := &MergeResult{Success: false}

	// Check for uncommitted changes in main workdir and stash if needed
	if s.git.HasUncommittedChanges(s.workDir) {
		if err := s.git.Stash(s.workDir); err != nil {
			return nil, fmt.Errorf("failed to stash changes: %w", err)
		}
		result.Stashed = true
	}

	// Merge the agent's branch
	if err := s.git.Merge(agent.Branch); err != nil {
		// Merge failed, likely a conflict
		result.ConflictErr = err
		// Pop stash if we stashed
		if result.Stashed {
			_ = s.git.StashPop(s.workDir)
		}
		return result, nil
	}

	result.Success = true

	// Pop stash if we stashed
	if result.Stashed {
		_ = s.git.StashPop(s.workDir)
	}

	return result, nil
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
