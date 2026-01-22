package dashboard

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
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

func (m *MockSessionManager) CreateSession(name, command, cwd string) (*tmux.Session, error) {
	m.createCalls++
	id := name
	s := &tmux.Session{ID: id, Name: id, Command: command, Active: true}
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

func (m *MockSessionManager) KillSession(name string) error {
	delete(m.sessions, name)
	return nil
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
	mockTmux.CreateSession(sessionName("Claude"), "claude", "")

	model := NewModel(cfg, mockTmux)

	// Verify list population
	items := model.list.Items()
	if len(items) != 1 {
		t.Fatalf("Expected 1 active item, got %d", len(items))
	}
	claudeItem := items[0].(AgentItem)
	if !claudeItem.Active {
		t.Error("Expected Claude to be marked Active")
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
	session := sessionName("Claude")
	mockTmux.CreateSession(session, "claude", "")
	mockTmux.logs[session] = testLogOutput

	model := NewModel(cfg, mockTmux)
	model.list.Select(0)

	tick := tickMsg(time.Now())
	_, cmd := model.Update(tick)
	if cmd == nil {
		t.Fatal("Tick should trigger a command batch")
	}

	fetchCmd := fetchPreviewCmd(mockTmux, session)
	msg := fetchCmd()

	resultMsg, ok := msg.(previewResultMsg)
	if !ok {
		t.Fatalf("Expected previewResultMsg, got %T", msg)
	}
	if string(resultMsg) != testLogOutput {
		t.Errorf("Expected log output, got %s", resultMsg)
	}

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
	model.list.SetItems([]list.Item{AgentItem{Name: "Claude", Command: "claude", SessionID: sessionName("Claude")}})
	model.list.Select(0)
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Fatal("Expected command after Enter")
	}

	cmd()

	expected := sessionName("Claude")
	if mockTmux.switchedTo != expected {
		t.Errorf("Expected switch to %s, got %s", expected, mockTmux.switchedTo)
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
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}
}
