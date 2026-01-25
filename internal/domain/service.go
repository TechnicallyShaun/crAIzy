package domain

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/logging"
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
	logging.Entry("agentType", agentType, "name", name, "command", command)
	sessionID := BuildSessionID(s.project, agentType, name)

	// Check if an active session already exists
	existing := s.store.Get(sessionID)
	if existing != nil && existing.Status == AgentStatusActive {
		err := fmt.Errorf("agent session %q already exists", sessionID)
		logging.Error(err, "sessionID", sessionID)
		return nil, err
	}

	// Remove any terminated agent with same ID before creating new one
	if existing != nil {
		_ = s.store.Remove(sessionID)
	}

	// Build branch name from session ID
	branchName := sessionID

	// Check if branch already exists
	if s.git != nil && s.git.BranchExists(branchName) {
		err := fmt.Errorf("branch %q already exists", branchName)
		logging.Error(err, "branch", branchName)
		return nil, err
	}

	// Get current branch as base
	var baseBranch string
	var worktreePath string
	if s.git != nil {
		var err error
		baseBranch, err = s.git.CurrentBranch(s.workDir)
		if err != nil {
			err = fmt.Errorf("failed to get current branch: %w", err)
			logging.Error(err, "workDir", s.workDir)
			return nil, err
		}

		// Create worktree path
		worktreePath = filepath.Join(s.workDir, WorktreesDir, SanitizeName(name))

		// Create worktree with new branch
		if err := s.git.CreateWorktree(worktreePath, branchName, baseBranch); err != nil {
			err = fmt.Errorf("failed to create worktree: %w", err)
			logging.Error(err, "worktreePath", worktreePath, "branch", branchName)
			return nil, err
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

	logging.Info("agent created successfully, sessionID=%s", sessionID)
	return agent, nil
}

// Kill terminates an agent session.
func (s *AgentService) Kill(sessionID string) error {
	logging.Entry("sessionID", sessionID)
	// Publish event - adapters will kill tmux session and update status
	s.dispatcher.Publish(AgentKilled{
		AgentID:   sessionID,
		Timestamp: time.Now(),
	})

	logging.Info("agent kill event published, sessionID=%s", sessionID)
	return nil
}

// CheckKill checks if an agent has uncommitted changes before killing.
// Returns true if there are uncommitted changes that need user confirmation.
func (s *AgentService) CheckKill(sessionID string) (hasUncommitted bool, err error) {
	logging.Entry("sessionID", sessionID)
	if s.git == nil {
		return false, nil
	}

	agent := s.store.Get(sessionID)
	if agent == nil {
		err := fmt.Errorf("agent %q not found", sessionID)
		logging.Error(err, "sessionID", sessionID)
		return false, err
	}

	if agent.Branch == "" {
		return false, nil
	}

	hasUncommitted = s.git.HasUncommittedChanges(agent.WorkDir)
	logging.Info("checked for uncommitted changes, sessionID=%s, hasUncommitted=%v", sessionID, hasUncommitted)
	return hasUncommitted, nil
}

// ForceKill terminates an agent, optionally discarding uncommitted changes.
func (s *AgentService) ForceKill(sessionID string, discardChanges bool) error {
	logging.Entry("sessionID", sessionID, "discardChanges", discardChanges)
	if s.git != nil && !discardChanges {
		agent := s.store.Get(sessionID)
		if agent != nil && agent.Branch != "" && s.git.HasUncommittedChanges(agent.WorkDir) {
			// Stash changes before killing
			logging.Info("stashing changes before kill, sessionID=%s", sessionID)
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
	logging.Entry("sessionID", sessionID)
	if s.git == nil {
		err := fmt.Errorf("git client not available")
		logging.Error(err)
		return nil, err
	}

	agent := s.store.Get(sessionID)
	if agent == nil {
		err := fmt.Errorf("agent %q not found", sessionID)
		logging.Error(err, "sessionID", sessionID)
		return nil, err
	}

	if agent.Branch == "" {
		err := fmt.Errorf("agent has no branch to merge")
		logging.Error(err, "sessionID", sessionID)
		return nil, err
	}

	result := &MergeResult{Success: false}

	// Check for uncommitted changes in main workdir and stash if needed
	if s.git.HasUncommittedChanges(s.workDir) {
		logging.Info("stashing uncommitted changes before merge")
		if err := s.git.Stash(s.workDir); err != nil {
			err = fmt.Errorf("failed to stash changes: %w", err)
			logging.Error(err)
			return nil, err
		}
		result.Stashed = true
	}

	// Merge the agent's branch
	if err := s.git.Merge(agent.Branch); err != nil {
		// Merge failed, likely a conflict
		logging.Error(err, "branch", agent.Branch, "conflict", true)
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

	logging.Info("merge completed successfully, sessionID=%s, branch=%s", sessionID, agent.Branch)
	return result, nil
}

// List returns active agents for the current project.
func (s *AgentService) List() []*Agent {
	logging.Entry("project", s.project)
	all := s.store.List()
	var active []*Agent
	for _, agent := range all {
		if agent.Project == s.project && agent.Status == AgentStatusActive {
			active = append(active, agent)
		}
	}
	logging.Debug("listed agents, count=%d", len(active))
	return active
}

// Attach returns a tea.Cmd that attaches to the given session.
// This will suspend the TUI and take over the terminal.
func (s *AgentService) Attach(sessionID string) tea.Cmd {
	logging.Entry("sessionID", sessionID)
	cmd := s.tmux.AttachCmd(sessionID)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			logging.Error(err, "sessionID", sessionID)
		}
		return AgentDetachedMsg{SessionID: sessionID, Err: err}
	})
}

// Exists checks if an agent exists in the store.
func (s *AgentService) Exists(sessionID string) bool {
	logging.Entry("sessionID", sessionID)
	return s.store.Exists(sessionID)
}

// CaptureOutput captures the last N lines from an agent's tmux pane.
func (s *AgentService) CaptureOutput(sessionID string, lines int) (string, error) {
	logging.Entry("sessionID", sessionID, "lines", lines)
	output, err := s.tmux.CapturePaneOutput(sessionID, lines)
	if err != nil {
		logging.Error(err, "sessionID", sessionID)
	}
	return output, err
}

// Reconcile synchronizes the store with actual tmux sessions.
// It marks agents as terminated if their tmux session no longer exists,
// and kills orphaned tmux sessions that aren't in the store.
func (s *AgentService) Reconcile() error {
	logging.Entry("project", s.project)
	// Get all stored agents
	agents := s.store.List()

	// Check for orphaned store entries (session doesn't exist in tmux)
	for _, agent := range agents {
		if agent.Status == AgentStatusTerminated {
			continue
		}
		if !s.tmux.SessionExists(agent.ID) {
			// Mark as terminated rather than removing
			logging.Info("marking orphaned agent as terminated, agentID=%s", agent.ID)
			_ = s.store.UpdateStatus(agent.ID, AgentStatusTerminated)
		}
	}

	// Get all tmux sessions
	sessions, err := s.tmux.ListSessions()
	if err != nil {
		// tmux might not be running, which is fine
		logging.Debug("tmux list sessions failed (may not be running): %v", err)
		return nil
	}

	// Check for orphaned tmux sessions (matches our prefix but not in store)
	prefix := "craizy-" + SanitizeName(s.project) + "-"
	for _, session := range sessions {
		if strings.HasPrefix(session, prefix) {
			if !s.store.Exists(session) {
				logging.Info("killing orphaned tmux session, session=%s", session)
				_ = s.tmux.KillSession(session)
			}
		}
	}

	logging.Info("reconcile completed")
	return nil
}

// AgentDetachedMsg is sent when returning from an attached tmux session.
type AgentDetachedMsg struct {
	SessionID string
	Err       error
}
