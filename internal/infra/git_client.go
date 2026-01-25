package infra

import (
	"os/exec"
	"path/filepath"
	"strings"
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
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// Init initializes a new git repository at the given path.
func (g *GitClient) Init(path string) error {
	cmd := exec.Command("git", "init", path)
	return cmd.Run()
}

// CurrentBranch returns the current branch name for the repo at path.
func (g *GitClient) CurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// BranchExists checks if a branch exists in the repository.
func (g *GitClient) BranchExists(branch string) bool {
	cmd := exec.Command("git", "-C", g.repoRoot, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return cmd.Run() == nil
}

// CreateWorktree creates a new worktree at path with the given branch.
// If the branch doesn't exist, it creates it from baseBranch.
func (g *GitClient) CreateWorktree(path, branch, baseBranch string) error {
	// Make path absolute if it isn't already
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if branch already exists
	if g.BranchExists(branch) {
		// Use existing branch
		cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "add", absPath, branch)
		return cmd.Run()
	}

	// Create new branch from baseBranch
	cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "add", "-b", branch, absPath, baseBranch)
	return cmd.Run()
}

// RemoveWorktree removes the worktree at the given path.
func (g *GitClient) RemoveWorktree(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "-C", g.repoRoot, "worktree", "remove", "--force", absPath)
	return cmd.Run()
}

// DeleteBranch deletes a branch from the repository.
func (g *GitClient) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "-C", g.repoRoot, "branch", "-D", branch)
	return cmd.Run()
}

// HasUncommittedChanges checks if the worktree at path has uncommitted changes.
func (g *GitClient) HasUncommittedChanges(path string) bool {
	// Check for staged or unstaged changes
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// DiscardChanges discards all uncommitted changes in the worktree at path.
func (g *GitClient) DiscardChanges(path string) error {
	// Reset staged changes
	cmd := exec.Command("git", "-C", path, "reset", "--hard", "HEAD")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Clean untracked files
	cmd = exec.Command("git", "-C", path, "clean", "-fd")
	return cmd.Run()
}

// Stash stashes changes in the worktree at path.
func (g *GitClient) Stash(path string) error {
	cmd := exec.Command("git", "-C", path, "stash", "push", "-u", "-m", "craizy-auto-stash")
	return cmd.Run()
}

// StashPop pops the stash in the worktree at path.
func (g *GitClient) StashPop(path string) error {
	cmd := exec.Command("git", "-C", path, "stash", "pop")
	return cmd.Run()
}

// Merge merges the given branch into the current branch.
func (g *GitClient) Merge(branch string) error {
	cmd := exec.Command("git", "-C", g.repoRoot, "merge", branch, "--no-edit")
	return cmd.Run()
}

// MergeAbort aborts an in-progress merge.
func (g *GitClient) MergeAbort() error {
	cmd := exec.Command("git", "-C", g.repoRoot, "merge", "--abort")
	return cmd.Run()
}
