package infra

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init", tmpDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
	_ = cmd.Run()

	// Create initial commit so we have a valid HEAD
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create test file: %v", err)
	}
	cmd = exec.Command("git", "-C", tmpDir, "add", ".")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create initial commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestGitClient_IsRepo(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Test with valid repo
	if !client.IsRepo(repoDir) {
		t.Error("IsRepo should return true for a valid git repository")
	}

	// Test with non-repo directory
	tmpDir, _ := os.MkdirTemp("", "non-git-*")
	defer os.RemoveAll(tmpDir)
	if client.IsRepo(tmpDir) {
		t.Error("IsRepo should return false for a non-git directory")
	}
}

func TestGitClient_Init(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client := NewGitClient(tmpDir)

	// Test initializing a new repo
	newRepoPath := filepath.Join(tmpDir, "new-repo")
	if err := os.MkdirAll(newRepoPath, 0755); err != nil {
		t.Fatalf("failed to create new repo dir: %v", err)
	}

	if err := client.Init(newRepoPath); err != nil {
		t.Errorf("Init should not return error: %v", err)
	}

	if !client.IsRepo(newRepoPath) {
		t.Error("Init should create a valid git repository")
	}
}

func TestGitClient_CurrentBranch(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	branch, err := client.CurrentBranch(repoDir)
	if err != nil {
		t.Errorf("CurrentBranch should not return error: %v", err)
	}

	// Default branch is usually "main" or "master"
	if branch != "main" && branch != "master" {
		t.Errorf("CurrentBranch returned unexpected branch: %s", branch)
	}
}

func TestGitClient_BranchExists(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Get current branch name
	currentBranch, _ := client.CurrentBranch(repoDir)

	// Test existing branch
	if !client.BranchExists(currentBranch) {
		t.Error("BranchExists should return true for current branch")
	}

	// Test non-existing branch
	if client.BranchExists("non-existent-branch") {
		t.Error("BranchExists should return false for non-existent branch")
	}
}

func TestGitClient_CreateWorktree(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Get base branch
	baseBranch, _ := client.CurrentBranch(repoDir)

	// Create worktree
	worktreePath := filepath.Join(repoDir, "..", "test-worktree")
	err := client.CreateWorktree(worktreePath, "test-branch", baseBranch)
	if err != nil {
		t.Errorf("CreateWorktree should not return error: %v", err)
	}
	defer os.RemoveAll(worktreePath)

	// Verify worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("CreateWorktree should create the worktree directory")
	}

	// Verify branch was created
	if !client.BranchExists("test-branch") {
		t.Error("CreateWorktree should create the branch")
	}

	// Verify it's the correct branch
	worktreeBranch, err := client.CurrentBranch(worktreePath)
	if err != nil {
		t.Errorf("CurrentBranch on worktree should not error: %v", err)
	}
	if worktreeBranch != "test-branch" {
		t.Errorf("Worktree should be on test-branch, got: %s", worktreeBranch)
	}
}

func TestGitClient_RemoveWorktree(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)
	baseBranch, _ := client.CurrentBranch(repoDir)

	// Create worktree first
	worktreePath := filepath.Join(repoDir, "..", "worktree-to-remove")
	_ = client.CreateWorktree(worktreePath, "remove-branch", baseBranch)

	// Remove worktree
	err := client.RemoveWorktree(worktreePath)
	if err != nil {
		t.Errorf("RemoveWorktree should not return error: %v", err)
	}

	// Verify worktree is removed
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("RemoveWorktree should remove the worktree directory")
	}
}

func TestGitClient_DeleteBranch(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)
	baseBranch, _ := client.CurrentBranch(repoDir)

	// Create a branch
	cmd := exec.Command("git", "-C", repoDir, "branch", "branch-to-delete", baseBranch)
	_ = cmd.Run()

	if !client.BranchExists("branch-to-delete") {
		t.Fatal("Branch should exist before deletion")
	}

	// Delete the branch
	err := client.DeleteBranch("branch-to-delete")
	if err != nil {
		t.Errorf("DeleteBranch should not return error: %v", err)
	}

	if client.BranchExists("branch-to-delete") {
		t.Error("DeleteBranch should remove the branch")
	}
}

func TestGitClient_HasUncommittedChanges(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Initially no uncommitted changes
	if client.HasUncommittedChanges(repoDir) {
		t.Error("HasUncommittedChanges should return false for clean repo")
	}

	// Create an uncommitted change
	testFile := filepath.Join(repoDir, "new-file.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !client.HasUncommittedChanges(repoDir) {
		t.Error("HasUncommittedChanges should return true after creating a file")
	}
}

func TestGitClient_DiscardChanges(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Create uncommitted changes
	testFile := filepath.Join(repoDir, "discard-me.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	// Modify existing file
	readmeFile := filepath.Join(repoDir, "README.md")
	_ = os.WriteFile(readmeFile, []byte("# Modified"), 0644)

	if !client.HasUncommittedChanges(repoDir) {
		t.Fatal("Should have uncommitted changes")
	}

	// Discard changes
	err := client.DiscardChanges(repoDir)
	if err != nil {
		t.Errorf("DiscardChanges should not return error: %v", err)
	}

	if client.HasUncommittedChanges(repoDir) {
		t.Error("DiscardChanges should remove all uncommitted changes")
	}

	// Verify new file was removed
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("DiscardChanges should remove untracked files")
	}
}

func TestGitClient_Stash(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Create uncommitted changes
	readmeFile := filepath.Join(repoDir, "README.md")
	_ = os.WriteFile(readmeFile, []byte("# Modified for stash"), 0644)

	// Stash changes
	err := client.Stash(repoDir)
	if err != nil {
		t.Errorf("Stash should not return error: %v", err)
	}

	if client.HasUncommittedChanges(repoDir) {
		t.Error("Stash should remove uncommitted changes from working directory")
	}
}

func TestGitClient_StashPop(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)

	// Create and stash changes
	readmeFile := filepath.Join(repoDir, "README.md")
	_ = os.WriteFile(readmeFile, []byte("# Modified for stash pop"), 0644)
	_ = client.Stash(repoDir)

	// Pop stash
	err := client.StashPop(repoDir)
	if err != nil {
		t.Errorf("StashPop should not return error: %v", err)
	}

	if !client.HasUncommittedChanges(repoDir) {
		t.Error("StashPop should restore uncommitted changes")
	}
}

func TestGitClient_Merge(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)
	baseBranch, _ := client.CurrentBranch(repoDir)

	// Create a feature branch with a commit
	cmd := exec.Command("git", "-C", repoDir, "checkout", "-b", "feature-branch")
	_ = cmd.Run()

	featureFile := filepath.Join(repoDir, "feature.txt")
	_ = os.WriteFile(featureFile, []byte("feature content"), 0644)

	cmd = exec.Command("git", "-C", repoDir, "add", ".")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Add feature")
	_ = cmd.Run()

	// Switch back to base branch
	cmd = exec.Command("git", "-C", repoDir, "checkout", baseBranch)
	_ = cmd.Run()

	// Merge feature branch
	err := client.Merge("feature-branch")
	if err != nil {
		t.Errorf("Merge should not return error: %v", err)
	}

	// Verify file from feature branch exists
	if _, err := os.Stat(featureFile); os.IsNotExist(err) {
		t.Error("Merge should bring in changes from feature branch")
	}
}

func TestGitClient_MergeAbort(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(repoDir)
	baseBranch, _ := client.CurrentBranch(repoDir)

	// Create conflicting changes
	readmeFile := filepath.Join(repoDir, "README.md")

	// Create feature branch with conflicting change
	cmd := exec.Command("git", "-C", repoDir, "checkout", "-b", "conflict-branch")
	_ = cmd.Run()
	_ = os.WriteFile(readmeFile, []byte("# Feature version"), 0644)
	cmd = exec.Command("git", "-C", repoDir, "add", ".")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Feature change")
	_ = cmd.Run()

	// Switch to base and make conflicting change
	cmd = exec.Command("git", "-C", repoDir, "checkout", baseBranch)
	_ = cmd.Run()
	_ = os.WriteFile(readmeFile, []byte("# Base version"), 0644)
	cmd = exec.Command("git", "-C", repoDir, "add", ".")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", repoDir, "commit", "-m", "Base change")
	_ = cmd.Run()

	// Attempt merge (should conflict)
	_ = client.Merge("conflict-branch")

	// Abort merge
	err := client.MergeAbort()
	if err != nil {
		t.Errorf("MergeAbort should not return error: %v", err)
	}
}
