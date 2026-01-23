package dashboard

import (
	"errors"
	"os/exec"
	"strings"
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
	sessions       map[string]*tmux.Session
	logs           map[string]string
	switchedTo     string
	createCalls    int
	createErr      error
	lastCreateCwd  string
	lastCreateName string
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		sessions: make(map[string]*tmux.Session),
		logs:     make(map[string]string),
	}
}

func (m *MockSessionManager) CreateSession(name, command, cwd string) (*tmux.Session, error) {
	m.createCalls++
	m.lastCreateCwd = cwd
	m.lastCreateName = name
	if m.createErr != nil {
		return nil, m.createErr
	}
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

func (m *MockSessionManager) AttachSession(sessionID string) error {
	// Mock implementation - just check if session exists
	if !m.SessionExists(sessionID) {
		return errors.New("session does not exist")
	}
	return nil
}

func (m *MockSessionManager) GetAttachCmd(sessionID string) *exec.Cmd {
	// Mock implementation - return a dummy command
	return exec.Command("true")
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
	if resultMsg.Content != testLogOutput {
		t.Errorf("Expected log output, got %s", resultMsg.Content)
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

	// Set TMUX env var to simulate being inside tmux
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")

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

func TestCreateSessionWithWorktree_Success(t *testing.T) {
	cfg := &config.Config{ProjectName: "proj"}
	mockTmux := NewMockSessionManager()
	agent := config.Agent{Name: "Claude", Command: "run"}

	msg := createSessionWithWorktreeCmd(mockTmux, nil, cfg, agent, "inst")()

	if _, ok := msg.(refreshListMsg); !ok {
		t.Fatalf("expected refreshListMsg, got %T", msg)
	}
	if mockTmux.createCalls != 1 {
		t.Fatalf("expected create session call once, got %d", mockTmux.createCalls)
	}
	if !mockTmux.SessionExists(sessionNameWithInstance("Claude", "inst")) {
		t.Fatal("session should exist after creation")
	}
}

func TestCreateSessionWithWorktree_AlreadyExists(t *testing.T) {
	cfg := &config.Config{ProjectName: "proj"}
	mockTmux := NewMockSessionManager()
	agent := config.Agent{Name: "Claude", Command: "run"}
	session := sessionNameWithInstance("Claude", "inst")
	mockTmux.sessions[session] = &tmux.Session{ID: session, Name: session}

	msg := createSessionWithWorktreeCmd(mockTmux, nil, cfg, agent, "inst")()

	res, ok := msg.(previewResultMsg)
	if !ok {
		t.Fatalf("expected previewResultMsg, got %T", msg)
	}
	if !strings.Contains(res.Content, "already exists") {
		t.Fatalf("expected already exists message, got %s", res.Content)
	}
	if mockTmux.createCalls != 0 {
		t.Fatalf("expected no create calls, got %d", mockTmux.createCalls)
	}
}

func TestCreateSessionWithWorktree_CreateError(t *testing.T) {
	cfg := &config.Config{ProjectName: "proj"}
	mockTmux := NewMockSessionManager()
	mockTmux.createErr = errors.New("boom")
	agent := config.Agent{Name: "Claude", Command: "run"}

	msg := createSessionWithWorktreeCmd(mockTmux, nil, cfg, agent, "inst")()

	res, ok := msg.(previewResultMsg)
	if !ok {
		t.Fatalf("expected previewResultMsg, got %T", msg)
	}
	if !strings.Contains(res.Content, "Error creating session") {
		t.Fatalf("expected create error message, got %s", res.Content)
	}
	if mockTmux.createCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", mockTmux.createCalls)
	}
}

func TestModal_SelectAgentPromptsForInstance(t *testing.T) {
	modal := NewModal([]config.Agent{{Name: "Claude"}, {Name: "Copilot"}})
	modal.Show()

	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command after selecting agent")
	}
	if _, ok := cmd().(promptInstanceMsg); !ok {
		t.Fatalf("expected promptInstanceMsg from modal selection")
	}
}

func TestModal_InstanceNameEnterValidates(t *testing.T) {
	modal := NewModal([]config.Agent{{Name: "Claude"}})
	modal.Show()
	modal.selected = modal.agents[0]
	modal.instanceName = "fix-typo"

	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command when confirming instance")
	}
	msg := cmd()
	v, ok := msg.(instanceValidatedMsg)
	if !ok {
		t.Fatalf("expected instanceValidatedMsg, got %T", msg)
	}
	if v.Name != "fix-typo" {
		t.Fatalf("expected name fix-typo, got %s", v.Name)
	}
	if v.Agent.Name != "Claude" {
		t.Fatalf("expected agent Claude, got %s", v.Agent.Name)
	}
}
