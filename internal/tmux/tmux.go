package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager manages tmux sessions by delegating to the tmux CLI.
type Manager struct{}

// Session represents a tmux session
type Session struct {
	ID      string
	Name    string
	Command string
	Active  bool
}

// NewManager creates a new tmux manager
func NewManager() *Manager {
	return &Manager{}
}

// CreateSession creates a new tmux session with the provided name and command.
// If cwd is non-empty, the session starts in that directory.
func (m *Manager) CreateSession(name, command, cwd string) (*Session, error) {
	if m.SessionExists(name) {
		return nil, fmt.Errorf("session %s already exists", name)
	}

	args := []string{"new-session", "-d", "-s", name}
	if command != "" {
		args = append(args, "bash", "-lc", command)
	}
	cmd := exec.Command("tmux", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	session := &Session{
		ID:      name,
		Name:    name,
		Command: command,
		Active:  true,
	}
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

// ListSessions returns all tmux sessions currently running
func (m *Manager) ListSessions() []*Session {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return []*Session{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	sessions := make([]*Session, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		sessions = append(sessions, &Session{
			ID:     line,
			Name:   line,
			Active: true,
		})
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

// SendKeysNoEnter sends keystrokes to a tmux session without appending Enter
func (m *Manager) SendKeysNoEnter(sessionID, keys string) error {
	if !m.SessionExists(sessionID) {
		return fmt.Errorf("session %s does not exist", sessionID)
	}

	cmd := exec.Command("tmux", "send-keys", "-t", sessionID, keys)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}

	return nil
}

// SendKeysLiteral sends literal keys to a tmux session (without Enter)
func (m *Manager) SendKeysLiteral(sessionID, keys string) error {
	if !m.SessionExists(sessionID) {
		return fmt.Errorf("session %s does not exist", sessionID)
	}

	cmd := exec.Command("tmux", "send-keys", "-t", sessionID, "-l", keys)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}

	return nil
}

// SwitchClient switches the current tmux client to the target session
func (m *Manager) SwitchClient(target string) error {
	if !m.SessionExists(target) {
		return fmt.Errorf("session %s does not exist", target)
	}

	cmd := exec.Command("tmux", "switch-client", "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch client: %w", err)
	}

	return nil
}

// CapturePane returns the last N lines of the target session's output
func (m *Manager) CapturePane(target string, lines int) (string, error) {
	if !m.SessionExists(target) {
		return "", fmt.Errorf("session %s does not exist", target)
	}

	// #nosec G204 -- target is validated via SessionExists, lines is an integer formatted safely
	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-S", fmt.Sprintf("-%d", lines))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane: %w", err)
	}

	return string(output), nil
}
