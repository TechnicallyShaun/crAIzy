package infra

import (
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
)

// WireAdapters connects event handlers to the dispatcher for store and tmux operations.
func WireAdapters(dispatcher domain.IEventDispatcher, store domain.IAgentStore, tmux domain.ITmuxClient, git domain.IGitClient) {
	logging.Entry()

	// Handle agent creation - create tmux session first, then store
	dispatcher.Subscribe("agent.created", func(e domain.Event) {
		event := e.(domain.AgentCreated)
		logging.Info("handling agent.created event, agentID=%s", event.Agent.ID)

		// Create tmux session first
		if err := tmux.CreateSession(event.Agent.ID, event.Agent.Command, event.Agent.WorkDir); err != nil {
			logging.Error(err, "agentID", event.Agent.ID, "action", "tmux.CreateSession")
			// Clean up worktree if tmux creation failed
			if git != nil && event.Agent.Branch != "" {
				logging.Info("cleaning up worktree after tmux creation failure")
				_ = git.RemoveWorktree(event.Agent.WorkDir)
				_ = git.DeleteBranch(event.Agent.Branch)
			}
			return // Don't store if tmux creation failed
		}

		// Then store the agent
		if err := store.Add(event.Agent); err != nil {
			logging.Error(err, "agentID", event.Agent.ID, "action", "store.Add")
			// Clean up tmux session if store fails
			_ = tmux.KillSession(event.Agent.ID)
			// Clean up worktree and branch
			if git != nil && event.Agent.Branch != "" {
				_ = git.RemoveWorktree(event.Agent.WorkDir)
				_ = git.DeleteBranch(event.Agent.Branch)
			}
		}
		logging.Info("agent.created event handled successfully, agentID=%s", event.Agent.ID)
	})

	// Handle agent killed - kill tmux, clean up git, and update status
	dispatcher.Subscribe("agent.killed", func(e domain.Event) {
		event := e.(domain.AgentKilled)
		logging.Info("handling agent.killed event, agentID=%s", event.AgentID)

		if err := tmux.KillSession(event.AgentID); err != nil {
			logging.Error(err, "agentID", event.AgentID, "action", "tmux.KillSession")
		}

		// Get agent info before updating status so we can clean up git
		agent := store.Get(event.AgentID)
		if agent != nil && git != nil && agent.Branch != "" {
			// Remove worktree and delete branch
			logging.Info("cleaning up git worktree and branch, branch=%s", agent.Branch)
			if err := git.RemoveWorktree(agent.WorkDir); err != nil {
				logging.Error(err, "workDir", agent.WorkDir, "action", "git.RemoveWorktree")
			}
			if err := git.DeleteBranch(agent.Branch); err != nil {
				logging.Error(err, "branch", agent.Branch, "action", "git.DeleteBranch")
			}
		}

		if err := store.UpdateStatus(event.AgentID, domain.AgentStatusTerminated); err != nil {
			logging.Error(err, "agentID", event.AgentID, "action", "store.UpdateStatus")
		}
		logging.Info("agent.killed event handled successfully, agentID=%s", event.AgentID)
	})
}
