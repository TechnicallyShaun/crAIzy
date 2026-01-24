package infra

import (
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// WireAdapters connects event handlers to the dispatcher for store and tmux operations.
func WireAdapters(dispatcher domain.IEventDispatcher, store domain.IAgentStore, tmux domain.ITmuxClient) {
	// Handle agent creation - create tmux session first, then store
	dispatcher.Subscribe("agent.created", func(e domain.Event) {
		event := e.(domain.AgentCreated)
		// Create tmux session first
		if err := tmux.CreateSession(event.Agent.ID, event.Agent.Command, event.Agent.WorkDir); err != nil {
			return // Don't store if tmux creation failed
		}
		// Then store the agent
		if err := store.Add(event.Agent); err != nil {
			// Clean up tmux session if store fails
			_ = tmux.KillSession(event.Agent.ID)
		}
	})

	// Handle agent killed - kill tmux and update status
	dispatcher.Subscribe("agent.killed", func(e domain.Event) {
		event := e.(domain.AgentKilled)
		_ = tmux.KillSession(event.AgentID)
		_ = store.UpdateStatus(event.AgentID, domain.AgentStatusTerminated)
	})
}
