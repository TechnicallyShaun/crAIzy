package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	BaseDir string
}

func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, "craizy-worktrees")
	}
	return &Manager{BaseDir: baseDir}
}

// CreateWorktree creates a new git worktree at base/{project}/{session}
func (m *Manager) CreateWorktree(project, session string) (string, error) {
	repoRoot, err := gitRoot()
	if err != nil {
		return "", fmt.Errorf("git root not found: %w", err)
	}

	target := filepath.Join(m.BaseDir, project, session)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", fmt.Errorf("failed to create worktree base: %w", err)
	}

	var out bytes.Buffer
	cmd := exec.Command("git", "worktree", "add", "--force", target)
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git worktree add failed: %w: %s", err, strings.TrimSpace(out.String()))
	}

	return target, nil
}

func gitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(output)), nil
}
