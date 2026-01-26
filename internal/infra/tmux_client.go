package infra

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/logging"
	"github.com/TechnicallyShaun/crAIzy/internal/tui/theme"
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
	logging.Entry("id", id, "command", command, "workDir", workDir)
	args := []string{"new-session", "-d", "-s", id, "-c", workDir}
	if command != "" {
		args = append(args, command)
	}
	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "id", id)
		return err
	}

	// Configure custom status bar for this session
	t.configureStatusBar(id)
	logging.Info("tmux session created, id=%s", id)
	return nil
}

// configureStatusBar sets up tmux session options including mouse support
// and a custom status bar. Uses Nord-inspired colors from the theme package.
func (t *TmuxClient) configureStatusBar(sessionID string) {
	ts := theme.TmuxStatusBar

	// Session configuration using theme colors
	setOptions := [][]string{
		// Enable mouse support for scrollback, pane selection, etc.
		{"-t", sessionID, "mouse", "on"},
		// Status bar colors
		{"-t", sessionID, "status-style", fmt.Sprintf("bg=%s,fg=%s", ts.Background, ts.Foreground)},
		// Left side: crAIzy branding + session info
		{"-t", sessionID, "status-left", fmt.Sprintf("#[fg=%s,bold] crAIzy #[fg=%s]│ #[fg=%s]#{session_name} ", ts.BrandColor, ts.SeparatorColor, ts.AccentColor)},
		{"-t", sessionID, "status-left-length", "50"},
		// Right side: detach hint + time
		{"-t", sessionID, "status-right", fmt.Sprintf("#[fg=%s]Detach: Ctrl+B, D #[fg=%s]│ #[fg=%s]%%H:%%M ", ts.MutedColor, ts.SeparatorColor, ts.AccentColor)},
		{"-t", sessionID, "status-right-length", "40"},
		// Center the window list
		{"-t", sessionID, "status-justify", "center"},
		// Window styling
		{"-t", sessionID, "window-status-format", fmt.Sprintf("#[fg=%s] #W ", ts.MutedColor)},
		{"-t", sessionID, "window-status-current-format", fmt.Sprintf("#[fg=%s,bold] #W ", ts.AccentColor)},
	}

	for _, opt := range setOptions {
		args := append([]string{"set-option"}, opt...)
		_ = exec.Command("tmux", args...).Run()
	}
}

// KillSession terminates a tmux session.
// Command: tmux kill-session -t {id}
func (t *TmuxClient) KillSession(id string) error {
	logging.Entry("id", id)
	cmd := exec.Command("tmux", "kill-session", "-t", id)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "id", id)
		return err
	}
	logging.Info("tmux session killed, id=%s", id)
	return nil
}

// ListSessions returns all tmux session names.
// Command: tmux list-sessions -F "#{session_name}"
func (t *TmuxClient) ListSessions() ([]string, error) {
	logging.Entry()
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		logging.Error(err)
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
	logging.Debug("listed %d tmux sessions", len(sessions))
	return sessions, nil
}

// AttachCmd returns an exec.Cmd that can be used to attach to a session.
// This command can be passed to tea.ExecProcess for proper terminal handling.
func (t *TmuxClient) AttachCmd(id string) *exec.Cmd {
	logging.Entry("id", id)
	return exec.Command("tmux", "attach", "-t", id)
}

// SessionExists checks if a tmux session exists.
// Command: tmux has-session -t {id}
func (t *TmuxClient) SessionExists(id string) bool {
	logging.Entry("id", id)
	cmd := exec.Command("tmux", "has-session", "-t", id)
	exists := cmd.Run() == nil
	logging.Debug("session exists=%v, id=%s", exists, id)
	return exists
}

// CapturePaneOutput captures the last N lines from a tmux pane.
// Command: tmux capture-pane -t {id} -p -S -{lines}
// Uses -S with negative number to start from N lines back in history.
func (t *TmuxClient) CapturePaneOutput(sessionID string, lines int) (string, error) {
	logging.Entry("sessionID", sessionID, "lines", lines)
	startLine := "-" + strconv.Itoa(lines)
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionID, "-p", "-S", startLine)
	output, err := cmd.Output()
	if err != nil {
		logging.Error(err, "sessionID", sessionID)
	}
	return string(output), err
}

// SendKeys sends text/commands to a tmux session.
// Uses two-step approach: sends text literally with -l flag, then sends C-m separately.
// This ensures text with special characters (like newlines) is sent exactly as-is,
// and Enter is sent as a distinct action to submit the input.
func (t *TmuxClient) SendKeys(sessionID, text string) error {
	logging.Entry("sessionID", sessionID, "textLen", len(text))

	// Step 1: Send text literally (no key interpretation)
	cmdText := exec.Command("tmux", "send-keys", "-l", "-t", sessionID, text)
	if err := cmdText.Run(); err != nil {
		logging.Error(err, "sessionID", sessionID, "step", "send text")
		return err
	}

	// Step 2: Send Enter separately to submit
	cmdEnter := exec.Command("tmux", "send-keys", "-t", sessionID, "C-m")
	if err := cmdEnter.Run(); err != nil {
		logging.Error(err, "sessionID", sessionID, "step", "send enter")
		return err
	}

	logging.Info("keys sent to tmux session, id=%s", sessionID)
	return nil
}
