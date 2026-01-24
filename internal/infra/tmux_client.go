package infra

import (
	"os/exec"
	"strings"
)

// TmuxClient implements ITmuxClient using real tmux commands.
type TmuxClient struct{}

// NewTmuxClient creates a new TmuxClient.
func NewTmuxClient() *TmuxClient {
	return &TmuxClient{}
}

// CreateSession creates a new detached tmux session.
// Command: tmux new-session -d -s {id} -c {workDir} {command}
func (t *TmuxClient) CreateSession(id, command, workDir string) error {
	args := []string{"new-session", "-d", "-s", id, "-c", workDir}
	if command != "" {
		args = append(args, command)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
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
