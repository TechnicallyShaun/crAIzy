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

	// CapturePaneOutput captures the last N lines from a tmux pane.
	CapturePaneOutput(sessionID string, lines int) (string, error)
}

// IGitClient defines the interface for git operations.
type IGitClient interface {
	// IsRepo checks if the given path is inside a git repository.
	IsRepo(path string) bool

	// Init initializes a new git repository at the given path.
	Init(path string) error

	// CurrentBranch returns the current branch name for the repo at path.
	CurrentBranch(path string) (string, error)

	// BranchExists checks if a branch exists in the repository.
	BranchExists(branch string) bool

	// CreateWorktree creates a new worktree at path with the given branch.
	// If the branch doesn't exist, it creates it from baseBranch.
	CreateWorktree(path, branch, baseBranch string) error

	// RemoveWorktree removes the worktree at the given path.
	RemoveWorktree(path string) error

	// DeleteBranch deletes a branch from the repository.
	DeleteBranch(branch string) error

	// HasUncommittedChanges checks if the worktree at path has uncommitted changes.
	HasUncommittedChanges(path string) bool

	// DiscardChanges discards all uncommitted changes in the worktree at path.
	DiscardChanges(path string) error

	// Stash stashes changes in the worktree at path.
	Stash(path string) error

	// StashPop pops the stash in the worktree at path.
	StashPop(path string) error

	// Merge merges the given branch into the current branch.
	Merge(branch string) error

	// MergeAbort aborts an in-progress merge.
	MergeAbort() error
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
