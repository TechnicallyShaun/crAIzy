package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Manager manages tmux sessions
type Manager struct {
	sessionPrefix string
	sessions      map[string]*Session
}

// Session represents a tmux session
type Session struct {
	ID      string
	Name    string
	Command string
	Active  bool
}

// NewManager creates a new tmux manager
func NewManager() *Manager {
	return &Manager{
		sessionPrefix: "craizy",
		sessions:      make(map[string]*Session),
	}
}

// CreateSession creates a new tmux session
func (m *Manager) CreateSession(name, command string) (*Session, error) {
	sessionName := fmt.Sprintf("%s-%s", m.sessionPrefix, name)

	// Check if session already exists
	if m.SessionExists(sessionName) {
		return nil, fmt.Errorf("session %s already exists", sessionName)
	}

	// Create new detached session
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, command)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	session := &Session{
		ID:      sessionName,
		Name:    name,
		Command: command,
		Active:  true,
	}

	m.sessions[sessionName] = session
	return session, nil
}

// CreateWindow creates a new tmux window and sends keys to it
func (m *Manager) CreateWindow(name, command string) (*Session, error) {
	windowName := fmt.Sprintf("%s-%s", m.sessionPrefix, name)

	// Check if we're in a tmux session
	// If not in tmux (TMUX env var not set), fall back to creating a new session
	// This ensures the command can still be executed in an isolated tmux session
	if os.Getenv("TMUX") == "" {
		return m.CreateSession(name, command)
	}

	// Create new window in current session
	cmd := exec.Command("tmux", "new-window", "-n", windowName)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Send the command keys to the new window
	if err := m.SendKeys(windowName, command); err != nil {
		return nil, fmt.Errorf("failed to send keys: %w", err)
	}

	session := &Session{
		ID:      windowName,
		Name:    name,
		Command: command,
		Active:  true,
	}

	m.sessions[windowName] = session
	return session, nil
}

// SendKeys sends keys to a tmux window or pane
func (m *Manager) SendKeys(target, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}
	return nil
}

// SessionExists checks if a tmux session exists
func (m *Manager) SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// ListSessions returns all managed sessions
func (m *Manager) ListSessions() []*Session {
	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// AttachSession attaches to a tmux session
func (m *Manager) AttachSession(sessionID string) error {
	if !m.SessionExists(sessionID) {
		return fmt.Errorf("session %s does not exist", sessionID)
	}

	cmd := exec.Command("tmux", "attach-session", "-t", sessionID)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// GetSessionContent retrieves the content of a tmux session
func (m *Manager) GetSessionContent(sessionID string) (string, error) {
	if !m.SessionExists(sessionID) {
		return "", fmt.Errorf("session %s does not exist", sessionID)
	}

	cmd := exec.Command("tmux", "capture-pane", "-t", sessionID, "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane: %w", err)
	}

	return string(output), nil
}

// KillSession terminates a tmux session
func (m *Manager) KillSession(sessionID string) error {
	if !m.SessionExists(sessionID) {
		return fmt.Errorf("session %s does not exist", sessionID)
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sessionID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}

	delete(m.sessions, sessionID)
	return nil
}

// IsTmuxAvailable checks if tmux is installed
func IsTmuxAvailable() bool {
	cmd := exec.Command("tmux", "-V")
	return cmd.Run() == nil
}

// GetTmuxVersion returns the tmux version
func GetTmuxVersion() (string, error) {
	cmd := exec.Command("tmux", "-V")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
