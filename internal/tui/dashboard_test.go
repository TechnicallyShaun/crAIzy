package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
)

func TestPreviewPollInterval(t *testing.T) {
	t.Run("interval is 2 seconds", func(t *testing.T) {
		expected := 2 * time.Second
		if PreviewPollInterval != expected {
			t.Errorf("PreviewPollInterval = %v, want %v", PreviewPollInterval, expected)
		}
	})
}

func TestModel_Update_PreviewTickMsg(t *testing.T) {
	t.Run("skips capture when ported in", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.isPortedIn = true
		m.width = 100
		m.height = 40

		tickMsg := PreviewTickMsg(time.Now())
		newModel, cmd := m.Update(tickMsg)

		// Should return a command (poll continuation) but model should still be ported in
		model := newModel.(Model)
		if !model.isPortedIn {
			t.Error("isPortedIn should remain true")
		}
		// Should have a command to continue polling
		if cmd == nil {
			t.Error("should return poll command even when ported in")
		}
	})

	t.Run("captures when not ported in", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.isPortedIn = false
		m.width = 100
		m.height = 40

		tickMsg := PreviewTickMsg(time.Now())
		_, cmd := m.Update(tickMsg)

		// Should return a batch command (capture + poll)
		if cmd == nil {
			t.Error("should return commands for capture and poll")
		}
	})
}

func TestModel_Update_PreviewUpdatedMsg(t *testing.T) {
	t.Run("updates content area preview", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40
		m.contentArea.SetSize(75, 35)

		msg := PreviewUpdatedMsg{
			SessionID: "test-session",
			Content:   "test output content",
		}
		newModel, _ := m.Update(msg)

		model := newModel.(Model)
		if model.contentArea.previewContent != "test output content" {
			t.Errorf("preview content = %q, want %q", model.contentArea.previewContent, "test output content")
		}
	})
}

func TestModel_Update_AgentsUpdatedMsg(t *testing.T) {
	t.Run("starts polling when agents exist", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40

		msg := AgentsUpdatedMsg{
			Agents: []*domain.Agent{
				{ID: "test-agent", Name: "test", Status: domain.AgentStatusActive},
			},
		}
		_, cmd := m.Update(msg)

		// Should return commands including poll and capture
		if cmd == nil {
			t.Error("should return commands when agents exist")
		}
	})

	t.Run("clears preview when no agents", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40
		m.contentArea.SetPreview("old content")

		msg := AgentsUpdatedMsg{Agents: []*domain.Agent{}}
		newModel, _ := m.Update(msg)

		model := newModel.(Model)
		if model.contentArea.previewContent != "" {
			t.Error("preview should be cleared when no agents")
		}
	})
}

func TestModel_Update_AgentDetachedMsg(t *testing.T) {
	t.Run("clears ported in flag", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.isPortedIn = true
		m.width = 100
		m.height = 40

		msg := domain.AgentDetachedMsg{}
		newModel, _ := m.Update(msg)

		model := newModel.(Model)
		if model.isPortedIn {
			t.Error("isPortedIn should be false after detach")
		}
	})

	t.Run("returns commands to resume polling", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.isPortedIn = true
		m.width = 100
		m.height = 40

		msg := domain.AgentDetachedMsg{}
		_, cmd := m.Update(msg)

		// Should return batch of commands (refresh, capture, poll)
		if cmd == nil {
			t.Error("should return commands to resume polling after detach")
		}
	})
}

func TestModel_Update_NavigationKeys(t *testing.T) {
	t.Run("up key is processed without panic", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40
		m.sideMenu.SetSize(25, 35)

		keyMsg := tea.KeyMsg{Type: tea.KeyUp}
		// Should not panic even with no agents
		newModel, _ := m.Update(keyMsg)

		// Model should be returned
		if newModel == nil {
			t.Error("Update should return a model")
		}
	})

	t.Run("down key is processed without panic", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40
		m.sideMenu.SetSize(25, 35)

		keyMsg := tea.KeyMsg{Type: tea.KeyDown}
		// Should not panic even with no agents
		newModel, _ := m.Update(keyMsg)

		// Model should be returned
		if newModel == nil {
			t.Error("Update should return a model")
		}
	})
}

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	t.Run("sets dimensions correctly", func(t *testing.T) {
		m := NewModel(nil, nil)

		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		newModel, _ := m.Update(msg)

		model := newModel.(Model)
		if model.width != 120 {
			t.Errorf("width = %d, want 120", model.width)
		}
		if model.height != 40 {
			t.Errorf("height = %d, want 40", model.height)
		}
	})

	t.Run("calculates correct content area size", func(t *testing.T) {
		m := NewModel(nil, nil)

		msg := tea.WindowSizeMsg{Width: 100, Height: 40}
		newModel, _ := m.Update(msg)

		model := newModel.(Model)
		// Content width should be 75% of total width
		expectedContentWidth := 100 - int(float64(100)*0.25)
		if model.contentArea.width != expectedContentWidth {
			t.Errorf("content width = %d, want %d", model.contentArea.width, expectedContentWidth)
		}
	})
}

func TestModel_Update_EnterKey(t *testing.T) {
	t.Run("sets ported in flag", func(t *testing.T) {
		// Create a mock service that doesn't actually attach
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40

		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, _ := m.Update(keyMsg)

		model := newModel.(Model)
		// Without an agent selected, should not change ported in state
		// (This tests the guard condition)
		if model.isPortedIn {
			t.Error("isPortedIn should remain false when no agent selected")
		}
	})
}

func TestModel_pollPreview(t *testing.T) {
	t.Run("returns tick command", func(t *testing.T) {
		m := NewModel(nil, nil)

		cmd := m.pollPreview()

		if cmd == nil {
			t.Error("pollPreview should return a command")
		}
	})
}

func TestModel_capturePreview(t *testing.T) {
	t.Run("returns nil when no agent selected", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40

		cmd := m.capturePreview()

		// No agent selected, should return nil
		if cmd != nil {
			t.Error("capturePreview should return nil when no agent selected")
		}
	})

	t.Run("returns nil when no service", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40

		cmd := m.capturePreview()

		if cmd != nil {
			t.Error("capturePreview should return nil when no service")
		}
	})
}

func TestModel_View(t *testing.T) {
	t.Run("returns loading when no dimensions", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 0
		m.height = 0

		view := m.View()

		if view != "Loading..." {
			t.Errorf("view = %q, want 'Loading...'", view)
		}
	})

	t.Run("renders full layout when dimensions set", func(t *testing.T) {
		m := NewModel(nil, nil)
		m.width = 100
		m.height = 40
		m.sideMenu.SetSize(25, 35)
		m.contentArea.SetSize(75, 35)
		m.quickCommands.SetSize(100, 3)

		view := m.View()

		// Should have some content (not just "Loading...")
		if view == "Loading..." || view == "" {
			t.Error("view should render full layout when dimensions set")
		}
	})
}
