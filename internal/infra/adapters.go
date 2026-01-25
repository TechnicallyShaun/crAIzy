package infra

import (
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

// WireAdapters connects event handlers to the dispatcher for store and tmux operations.
func WireAdapters(dispatcher domain.IEventDispatcher, store domain.IAgentStore, tmux domain.ITmuxClient, git domain.IGitClient) {
	// Handle agent creation - create tmux session first, then store
	dispatcher.Subscribe("agent.created", func(e domain.Event) {
		event := e.(domain.AgentCreated)
		// Create tmux session first
		if err := tmux.CreateSession(event.Agent.ID, event.Agent.Command, event.Agent.WorkDir); err != nil {
			// Clean up worktree if tmux creation failed
			if git != nil && event.Agent.Branch != "" {
				_ = git.RemoveWorktree(event.Agent.WorkDir)
				_ = git.DeleteBranch(event.Agent.Branch)
			}
			return // Don't store if tmux creation failed
		}
		// Then store the agent
		if err := store.Add(event.Agent); err != nil {
			// Clean up tmux session if store fails
			_ = tmux.KillSession(event.Agent.ID)
			// Clean up worktree and branch
			if git != nil && event.Agent.Branch != "" {
				_ = git.RemoveWorktree(event.Agent.WorkDir)
				_ = git.DeleteBranch(event.Agent.Branch)
			}
		}
	})

	// Handle agent killed - kill tmux, clean up git, and update status
	dispatcher.Subscribe("agent.killed", func(e domain.Event) {
		event := e.(domain.AgentKilled)
		_ = tmux.KillSession(event.AgentID)

		// Get agent info before updating status so we can clean up git
		agent := store.Get(event.AgentID)
		if agent != nil && git != nil && agent.Branch != "" {
			// Remove worktree and delete branch
			_ = git.RemoveWorktree(agent.WorkDir)
			_ = git.DeleteBranch(agent.Branch)
		}

		_ = store.UpdateStatus(event.AgentID, domain.AgentStatusTerminated)
	})
}
