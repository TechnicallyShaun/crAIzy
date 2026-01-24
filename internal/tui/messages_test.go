package tui

import (
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

func TestPreviewTickMsg(t *testing.T) {
	t.Run("can be created from time.Time", func(t *testing.T) {
		now := time.Now()
		msg := PreviewTickMsg(now)

		// Should be convertible back to time.Time
		timestamp := time.Time(msg)
		if !timestamp.Equal(now) {
			t.Errorf("timestamp mismatch: got %v, want %v", timestamp, now)
		}
	})
}

func TestPreviewUpdatedMsg(t *testing.T) {
	t.Run("holds session and content", func(t *testing.T) {
		msg := PreviewUpdatedMsg{
			SessionID: "craizy-proj-claude-task",
			Content:   "Some terminal output\nMore output",
		}

		if msg.SessionID != "craizy-proj-claude-task" {
			t.Errorf("SessionID = %q, want %q", msg.SessionID, "craizy-proj-claude-task")
		}
		if msg.Content != "Some terminal output\nMore output" {
			t.Errorf("Content mismatch")
		}
	})

	t.Run("handles empty content", func(t *testing.T) {
		msg := PreviewUpdatedMsg{
			SessionID: "session",
			Content:   "",
		}

		if msg.Content != "" {
			t.Error("empty content should remain empty")
		}
	})
}

func TestAgentsUpdatedMsg(t *testing.T) {
	t.Run("holds agent list", func(t *testing.T) {
		agents := []*domain.Agent{
			{ID: "agent1", Name: "first"},
			{ID: "agent2", Name: "second"},
		}
		msg := AgentsUpdatedMsg{Agents: agents}

		if len(msg.Agents) != 2 {
			t.Errorf("agent count = %d, want 2", len(msg.Agents))
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		msg := AgentsUpdatedMsg{Agents: []*domain.Agent{}}

		if msg.Agents == nil {
			t.Error("Agents should be empty slice, not nil")
		}
		if len(msg.Agents) != 0 {
			t.Errorf("agent count = %d, want 0", len(msg.Agents))
		}
	})
}

func TestAgentSelectedMsg(t *testing.T) {
	t.Run("holds agent config", func(t *testing.T) {
		agent := config.Agent{
			Name:    "claude",
			Command: "claude-code",
		}
		msg := AgentSelectedMsg{Agent: agent}

		if msg.Agent.Name != "claude" {
			t.Errorf("Agent.Name = %q, want %q", msg.Agent.Name, "claude")
		}
	})
}

func TestAgentCreatedMsg(t *testing.T) {
	t.Run("holds agent and custom name", func(t *testing.T) {
		agent := config.Agent{
			Name:    "claude",
			Command: "claude-code",
		}
		msg := AgentCreatedMsg{
			Agent:      agent,
			CustomName: "my-task",
		}

		if msg.Agent.Name != "claude" {
			t.Errorf("Agent.Name = %q, want %q", msg.Agent.Name, "claude")
		}
		if msg.CustomName != "my-task" {
			t.Errorf("CustomName = %q, want %q", msg.CustomName, "my-task")
		}
	})
}

func TestCloseModalMsg(t *testing.T) {
	t.Run("is empty struct", func(t *testing.T) {
		msg := CloseModalMsg{}
		_ = msg // Should compile without issues
	})
}
