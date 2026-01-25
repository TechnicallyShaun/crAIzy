package tui

import (
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// CloseModalMsg signals to close the current modal.
type CloseModalMsg struct{}

// AgentSelectedMsg is sent when a user selects an agent type from the selector.
type AgentSelectedMsg struct {
	Agent config.Agent
}

// AgentCreatedMsg is sent when a user confirms agent creation with a custom name.
type AgentCreatedMsg struct {
	Agent      config.Agent
	CustomName string
}

// AgentsUpdatedMsg signals that the agent list has changed and UI should refresh.
type AgentsUpdatedMsg struct {
	Agents []*domain.Agent
}

// PreviewTickMsg signals that it's time to poll for preview updates.
type PreviewTickMsg time.Time

// PreviewUpdatedMsg carries updated preview content from a tmux pane.
type PreviewUpdatedMsg struct {
	SessionID string
	Content   string
}

// KillConfirmChoice represents the user's choice in the kill confirmation modal.
type KillConfirmChoice int

const (
	KillConfirmCancel KillConfirmChoice = iota
	KillConfirmKeep
	KillConfirmDiscard
)

// KillConfirmResultMsg is sent when the user makes a choice in the kill confirmation modal.
type KillConfirmResultMsg struct {
	SessionID string
	Choice    KillConfirmChoice
}

// MergeResultMsg is sent when a merge operation completes.
type MergeResultMsg struct {
	AgentName     string
	Success       bool
	Stashed       bool
	ConflictErr   error
	ConflictFiles []string
	BaseBranch    string
	AgentID       string
}

// MergeConflictChoice represents the user's choice in the merge conflict modal.
type MergeConflictChoice int

const (
	MergeConflictCancel MergeConflictChoice = iota
	MergeConflictSendToTerminal
)

// MergeConflictResultMsg is sent when the user makes a choice in the merge conflict modal.
type MergeConflictResultMsg struct {
	AgentID       string
	BaseBranch    string
	ConflictFiles []string
	Choice        MergeConflictChoice
}
