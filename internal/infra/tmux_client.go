package infra

import (
	"os/exec"
	"strconv"
	"strings"
)

// TmuxClient implements ITmuxClient using real tmux commands.
type TmuxClient struct{}

// NewTmuxClient creates a new TmuxClient.
func NewTmuxClient() *TmuxClient {
	return &TmuxClient{}
}

// CreateSession creates a new detached tmux session with a custom status bar.
// Command: tmux new-session -d -s {id} -c {workDir} {command}
func (t *TmuxClient) CreateSession(id, command, workDir string) error {
	args := []string{"new-session", "-d", "-s", id, "-c", workDir}
	if command != "" {
		args = append(args, command)
	}
	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Configure custom status bar for this session
	t.configureStatusBar(id)
	return nil
}

// configureStatusBar sets up a custom status bar for the tmux session.
func (t *TmuxClient) configureStatusBar(sessionID string) {
	// Status bar styling
	setOptions := [][]string{
		// Status bar colors
		{"-t", sessionID, "status-style", "bg=#1a1a2e,fg=#8b8b9e"},
		// Left side: crAIzy branding + session info
		{"-t", sessionID, "status-left", "#[fg=#7c3aed,bold] crAIzy #[fg=#8b8b9e]│ #[fg=#a78bfa]#{session_name} "},
		{"-t", sessionID, "status-left-length", "50"},
		// Right side: detach hint + time
		{"-t", sessionID, "status-right", "#[fg=#6b7280]Detach: Ctrl+B, D #[fg=#8b8b9e]│ #[fg=#a78bfa]%H:%M "},
		{"-t", sessionID, "status-right-length", "40"},
		// Center the window list
		{"-t", sessionID, "status-justify", "centre"},
		// Window styling
		{"-t", sessionID, "window-status-format", "#[fg=#6b7280] #W "},
		{"-t", sessionID, "window-status-current-format", "#[fg=#a78bfa,bold] #W "},
	}

	for _, opt := range setOptions {
		args := append([]string{"set-option"}, opt...)
		exec.Command("tmux", args...).Run()
	}
}

// KillSession terminates a tmux session.
// Command: tmux kill-session -t {id}
func (t *TmuxClient) KillSession(id string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", id)
	return cmd.Run()
}

// ListSessions returns all tmux session names.
// Command: tmux list-sessions -F "#{session_name}"
func (t *TmuxClient) ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty lines
	var sessions []string
	for _, line := range lines {
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// AttachCmd returns an exec.Cmd that can be used to attach to a session.
// This command can be passed to tea.ExecProcess for proper terminal handling.
func (t *TmuxClient) AttachCmd(id string) *exec.Cmd {
	return exec.Command("tmux", "attach", "-t", id)
}

// SessionExists checks if a tmux session exists.
// Command: tmux has-session -t {id}
func (t *TmuxClient) SessionExists(id string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", id)
	return cmd.Run() == nil
}

// CapturePaneOutput captures the last N lines from a tmux pane.
// Command: tmux capture-pane -t {id} -p -S -{lines}
// Uses -S with negative number to start from N lines back in history.
func (t *TmuxClient) CapturePaneOutput(sessionID string, lines int) (string, error) {
	startLine := "-" + strconv.Itoa(lines)
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionID, "-p", "-S", startLine)
	output, err := cmd.Output()
	return string(output), err
}
