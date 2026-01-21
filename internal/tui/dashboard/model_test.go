package dashboard

import (
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

const testLogOutput = "Output Log Line 1"

// MockSessionManager implements SessionManager for testing
type MockSessionManager struct {
	sessions    map[string]*tmux.Session
	logs        map[string]string
	switchedTo  string
	createCalls int
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		sessions: make(map[string]*tmux.Session),
		logs:     make(map[string]string),
	}
}

func (m *MockSessionManager) CreateSession(name, command string) (*tmux.Session, error) {
	m.createCalls++
	id := "craizy-" + name
	s := &tmux.Session{ID: id, Name: name, Command: command, Active: true}
	m.sessions[id] = s
	return s, nil
}

func (m *MockSessionManager) ListSessions() []*tmux.Session {
	var list []*tmux.Session
	for _, s := range m.sessions {
		list = append(list, s)
	}
	return list
}

func (m *MockSessionManager) SessionExists(name string) bool {
	_, exists := m.sessions[name]
	return exists
}

func (m *MockSessionManager) SwitchClient(target string) error {
	m.switchedTo = target
	return nil
}

func (m *MockSessionManager) CapturePane(target string, lines int) (string, error) {
	if val, ok := m.logs[target]; ok {
		return val, nil
	}
	return "", nil
}

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-project",
		Agents: []config.Agent{
			{Name: "Claude", Command: "claude"},
			{Name: "GPT4", Command: "gpt4"},
		},
	}
	mockTmux := NewMockSessionManager()

	// Setup existing session
	mockTmux.CreateSession("Claude", "claude")

	model := NewModel(cfg, mockTmux)

	// Verify list population
	items := model.list.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Check if active status was detected
	claudeItem := items[0].(AgentItem)
	if !claudeItem.Active {
		t.Error("Expected Claude to be marked Active")
	}

	gptItem := items[1].(AgentItem)
	if gptItem.Active {
		t.Error("Expected GPT4 to be marked Idle")
	}
}

func TestUpdate_Resize(t *testing.T) {
	cfg := &config.Config{Agents: []config.Agent{{Name: "Test", Command: "cmd"}}}
	model := NewModel(cfg, NewMockSessionManager())

	width, height := 120, 40
	msg := tea.WindowSizeMsg{Width: width, Height: height}

	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	if m.width != width {
		t.Errorf("Width not updated. Got %d, want %d", m.width, width)
	}

	// Check if view renders without error after resize
	view := m.View()
	if view == "" {
		t.Error("View returned empty string after resize")
	}
}

func TestUpdate_PreviewLoop(t *testing.T) {
	cfg := &config.Config{
		Agents: []config.Agent{{Name: "Claude", Command: "claude"}},
	}
	mockTmux := NewMockSessionManager()
	mockTmux.CreateSession("Claude", "claude")
	mockTmux.logs["craizy-Claude"] = testLogOutput

	model := NewModel(cfg, mockTmux)
	model.list.Select(0) // Select Claude

	// 1. Simulate Tick -> Should return Fetch Command
	tick := tickMsg(time.Now())
	_, cmd := model.Update(tick)

	if cmd == nil {
		t.Fatal("Tick should trigger a command batch")
	}

	// 2. Execute the batch manually (Bubble Tea internals usually do this)
	// We assume the batch contains fetchPreviewCmd. We invoke it directly for testing.
	fetchCmd := fetchPreviewCmd(mockTmux, "Claude")
	msg := fetchCmd()

	// 3. Verify the result message
	resultMsg, ok := msg.(previewResultMsg)
	if !ok {
		t.Fatalf("Expected previewResultMsg, got %T", msg)
	}
	if string(resultMsg) != testLogOutput {
		t.Errorf("Expected log output, got %s", resultMsg)
	}

	// 4. Send result back to update model state
	updatedModel, _ := model.Update(resultMsg)
	m := updatedModel.(Model)

	if m.previewContent != testLogOutput {
		t.Errorf("Preview content not updated in model. Got: %s", m.previewContent)
	}
}

func TestUpdate_Attach(t *testing.T) {
	cfg := &config.Config{
		Agents: []config.Agent{{Name: "Claude", Command: "claude"}},
	}
	mockTmux := NewMockSessionManager()
	model := NewModel(cfg, mockTmux)

	// Simulate pressing Enter
	model.list.Select(0)
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Fatal("Expected command after Enter")
	}

	// Execute command
	cmd()

	// Verify session was created and switched to
	if mockTmux.createCalls != 1 {
		t.Error("Expected CreateSession to be called")
	}

	expectedTarget := "craizy-Claude"
	if mockTmux.switchedTo != expectedTarget {
		t.Errorf("Expected switch to %s, got %s", expectedTarget, mockTmux.switchedTo)
	}
}

func TestUpdate_Quit(t *testing.T) {
	model := NewModel(&config.Config{}, nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if !m.quitting {
		t.Error("Model should be quitting")
	}

	// tea.Quit is a special internal command, but we can check if it returns something
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}
}
