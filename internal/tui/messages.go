package tui

import (
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
