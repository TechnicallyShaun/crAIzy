package infra

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/logging"
)

// GitClient implements domain.IGitClient using git commands.
type GitClient struct {
	// repoRoot is the root directory of the git repository.
	repoRoot string
}

// NewGitClient creates a new GitClient for the given repository root.
func NewGitClient(repoRoot string) *GitClient {
	return &GitClient{repoRoot: repoRoot}
}

// IsRepo checks if the given path is inside a git repository.
func (g *GitClient) IsRepo(path string) bool {
	logging.Entry("path", path)
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	result := cmd.Run() == nil
	logging.Debug("IsRepo result=%v", result)
	return result
}

// Init initializes a new git repository at the given path.
func (g *GitClient) Init(path string) error {
	logging.Entry("path", path)
	cmd := exec.Command("git", "init", path)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "path", path)
		return err
	}
	logging.Info("git repository initialized, path=%s", path)
	return nil
}

// CurrentBranch returns the current branch name for the repo at path.
func (g *GitClient) CurrentBranch(path string) (string, error) {
	logging.Entry("path", path)
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		logging.Error(err, "path", path)
		return "", err
	}
	branch := strings.TrimSpace(string(output))
	logging.Debug("current branch=%s", branch)
	return branch, nil
}

// BranchExists checks if a branch exists in the repository.
func (g *GitClient) BranchExists(branch string) bool {
	logging.Entry("branch", branch)
	cmd := exec.Command("git", "-C", g.repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	exists := cmd.Run() == nil
	logging.Debug("branch exists=%v", exists)
	return exists
}

// CreateWorktree creates a new worktree at path with the given branch.
// If the branch doesn't exist, it creates it from baseBranch.
func (g *GitClient) CreateWorktree(path, branch, baseBranch string) error {
	logging.Entry("path", path, "branch", branch, "baseBranch", baseBranch)
	// Make path absolute if it isn't already
	absPath, err := filepath.Abs(path)
	if err != nil {
		logging.Error(err, "path", path)
		return err
	}

	// Check if branch already exists
	if g.BranchExists(branch) {
		// Use existing branch
		cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "add", absPath, branch)
		if err := cmd.Run(); err != nil {
			logging.Error(err, "absPath", absPath, "branch", branch)
			return err
		}
		logging.Info("worktree created with existing branch, path=%s, branch=%s", absPath, branch)
		return nil
	}

	// Create new branch from baseBranch
	cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "add", "-b", branch, absPath, baseBranch)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "absPath", absPath, "branch", branch, "baseBranch", baseBranch)
		return err
	}
	logging.Info("worktree created with new branch, path=%s, branch=%s", absPath, branch)
	return nil
}

// RemoveWorktree removes the worktree at the given path.
func (g *GitClient) RemoveWorktree(path string) error {
	logging.Entry("path", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		logging.Error(err, "path", path)
		return err
	}

	cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "remove", "--force", absPath)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "absPath", absPath)
		return err
	}
	logging.Info("worktree removed, path=%s", absPath)
	return nil
}

// DeleteBranch deletes a branch from the repository.
func (g *GitClient) DeleteBranch(branch string) error {
	logging.Entry("branch", branch)
	cmd := exec.Command("git", "-C", g.repoRoot, "branch", "-D", branch)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "branch", branch)
		return err
	}
	logging.Info("branch deleted, branch=%s", branch)
	return nil
}

// HasUncommittedChanges checks if the worktree at path has uncommitted changes.
func (g *GitClient) HasUncommittedChanges(path string) bool {
	logging.Entry("path", path)
	// Check for staged or unstaged changes
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		logging.Error(err, "path", path)
		return false
	}
	hasChanges := len(strings.TrimSpace(string(output))) > 0
	logging.Debug("hasUncommittedChanges=%v", hasChanges)
	return hasChanges
}

// DiscardChanges discards all uncommitted changes in the worktree at path.
func (g *GitClient) DiscardChanges(path string) error {
	logging.Entry("path", path)
	// Reset staged changes
	cmd := exec.Command("git", "-C", path, "reset", "--hard", "HEAD")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "path", path, "action", "reset")
		return err
	}

	// Clean untracked files
	cmd = exec.Command("git", "-C", path, "clean", "-fd")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "path", path, "action", "clean")
		return err
	}
	logging.Info("changes discarded, path=%s", path)
	return nil
}

// Stash stashes changes in the worktree at path.
func (g *GitClient) Stash(path string) error {
	logging.Entry("path", path)
	cmd := exec.Command("git", "-C", path, "stash", "push", "-u", "-m", "craizy-auto-stash")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "path", path)
		return err
	}
	logging.Info("changes stashed, path=%s", path)
	return nil
}

// StashPop pops the stash in the worktree at path.
func (g *GitClient) StashPop(path string) error {
	logging.Entry("path", path)
	cmd := exec.Command("git", "-C", path, "stash", "pop")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "path", path)
		return err
	}
	logging.Info("stash popped, path=%s", path)
	return nil
}

// Merge merges the given branch into the current branch.
func (g *GitClient) Merge(branch string) error {
	logging.Entry("branch", branch)
	cmd := exec.Command("git", "-C", g.repoRoot, "merge", branch, "--no-edit")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "branch", branch)
		return err
	}
	logging.Info("branch merged, branch=%s", branch)
	return nil
}

// MergeAbort aborts an in-progress merge.
func (g *GitClient) MergeAbort() error {
	logging.Entry()
	cmd := exec.Command("git", "-C", g.repoRoot, "merge", "--abort")
	if err := cmd.Run(); err != nil {
		logging.Error(err)
		return err
	}
	logging.Info("merge aborted")
	return nil
}
