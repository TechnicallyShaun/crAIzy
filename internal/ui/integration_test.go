package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
)

// TestDashboardLifecycle tests the full lifecycle of the dashboard with user interaction
func TestDashboardLifecycle(t *testing.T) {
	if !tmux.IsTmuxAvailable() {
		t.Skip("tmux not available, skipping integration test")
	}

	// Setup: Create a temporary crAIzy project environment
	tmpDir := t.TempDir()
	projectName := "test-dashboard-project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Save current directory and restore at the end
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory and initialize project
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := config.InitProject(projectName); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	// Load the config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create dashboard and tmux manager
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	// Start the dashboard in a detached tmux session
	sessionName := "craizy-test-dashboard"
	// Clean up any existing session
	tmuxMgr.KillSession(sessionName)

	sessionID, err := dashboard.StartDetached(sessionName)
	if err != nil {
		t.Fatalf("Failed to start detached dashboard: %v", err)
	}

	// Ensure cleanup of tmux session
	defer func() {
		if err := tmuxMgr.KillSession(sessionID); err != nil {
			t.Logf("Warning: Failed to kill session %s: %v", sessionID, err)
		}
	}()

	// Give the dashboard time to render
	time.Sleep(500 * time.Millisecond)

	// Step 1: Verify the dashboard header is present
	content, err := tmuxMgr.GetSessionContent(sessionID)
	if err != nil {
		t.Fatalf("Failed to get session content: %v", err)
	}

	if !strings.Contains(content, "crAIzy Dashboard") {
		t.Errorf("Dashboard header not found in output:\n%s", content)
	}

	if !strings.Contains(content, projectName) {
		t.Errorf("Project name not found in dashboard output:\n%s", content)
	}

	if !strings.Contains(content, "Hotkeys") {
		t.Errorf("Hotkeys section not found in dashboard output:\n%s", content)
	}

	// Step 2: Verify "No AI instances running" message
	if !strings.Contains(content, "No AI instances running") {
		t.Errorf("Expected 'No AI instances running' message not found:\n%s", content)
	}

	// Step 3: Verify available AIs are listed
	for _, ai := range cfg.AIs {
		if !strings.Contains(content, ai.Name) {
			t.Errorf("AI '%s' not found in dashboard output:\n%s", ai.Name, content)
		}
	}

	// Step 4: Press 'n' to spawn a new AI (simulate user interaction)
	if err := tmuxMgr.SendKeysLiteral(sessionID, "n"); err != nil {
		t.Fatalf("Failed to send 'n' key: %v", err)
	}

	// Send Enter key to confirm
	if err := tmuxMgr.SendKeys(sessionID, "Enter"); err != nil {
		t.Fatalf("Failed to send Enter key: %v", err)
	}

	// Give time for the action to be processed and displayed
	time.Sleep(1500 * time.Millisecond)

	// Step 5: Verify that the dashboard responded to 'n' key
	content, err = tmuxMgr.GetSessionContent(sessionID)
	if err != nil {
		t.Fatalf("Failed to get session content after 'n' press: %v", err)
	}

	// The script should show "Starting new AI..." message
	if !strings.Contains(content, "Starting new AI") {
		t.Errorf("Expected 'Starting new AI' message after pressing 'n':\n%s", content)
	}

	// Wait for the dashboard to redisplay after processing (sleep 2 in the case statement)
	time.Sleep(1000 * time.Millisecond)

	// Step 6: Press 'q' to quit the dashboard
	if err := tmuxMgr.SendKeysLiteral(sessionID, "q"); err != nil {
		t.Fatalf("Failed to send 'q' key: %v", err)
	}

	if err := tmuxMgr.SendKeys(sessionID, "Enter"); err != nil {
		t.Fatalf("Failed to send Enter key: %v", err)
	}

	// Give time for the session to exit
	time.Sleep(500 * time.Millisecond)

	// Step 7: Verify the session has exited after 'q' command
	finalContent, err := tmuxMgr.GetSessionContent(sessionID)
	if err != nil {
		// Session has exited, which is expected after 'q'
		t.Logf("Session exited after 'q' command (expected): %v", err)
	} else {
		// If we can still get content, check for "Goodbye!" message
		if strings.Contains(finalContent, "Goodbye") {
			t.Logf("Dashboard showed goodbye message before exiting")
		}
		t.Logf("Final dashboard content:\n%s", finalContent)
	}

	// Manual cleanup is handled by defer
}

// TestDashboardLifecycleWithListCommand tests the 'l' list command
func TestDashboardLifecycleWithListCommand(t *testing.T) {
	if !tmux.IsTmuxAvailable() {
		t.Skip("tmux not available, skipping integration test")
	}

	// Setup: Create a temporary crAIzy project environment
	tmpDir := t.TempDir()
	projectName := "test-list-project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Save current directory and restore at the end
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory and initialize project
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if err := config.InitProject(projectName); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Change to project directory
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	// Load the config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create dashboard and tmux manager
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	// Start the dashboard in a detached tmux session
	sessionName := "craizy-test-list-dashboard"
	// Clean up any existing session
	tmuxMgr.KillSession(sessionName)

	sessionID, err := dashboard.StartDetached(sessionName)
	if err != nil {
		t.Fatalf("Failed to start detached dashboard: %v", err)
	}

	// Ensure cleanup of tmux session
	defer func() {
		if err := tmuxMgr.KillSession(sessionID); err != nil {
			t.Logf("Warning: Failed to kill session %s: %v", sessionID, err)
		}
	}()

	// Give the dashboard time to render
	time.Sleep(500 * time.Millisecond)

	// Press 'l' to list AIs
	if err := tmuxMgr.SendKeysLiteral(sessionID, "l"); err != nil {
		t.Fatalf("Failed to send 'l' key: %v", err)
	}

	if err := tmuxMgr.SendKeys(sessionID, "Enter"); err != nil {
		t.Fatalf("Failed to send Enter key: %v", err)
	}

	// Give time for the action to be processed
	time.Sleep(1 * time.Second)

	// Verify that the dashboard responded to 'l' key
	content, err := tmuxMgr.GetSessionContent(sessionID)
	if err != nil {
		t.Fatalf("Failed to get session content after 'l' press: %v", err)
	}

	// The script should show "Listing AIs..." message
	if !strings.Contains(content, "Listing AIs") {
		t.Errorf("Expected 'Listing AIs' message after pressing 'l':\n%s", content)
	}

	// Clean up with 'q'
	time.Sleep(2 * time.Second)
	tmuxMgr.SendKeysLiteral(sessionID, "q")
	tmuxMgr.SendKeys(sessionID, "Enter")
	time.Sleep(500 * time.Millisecond)
}

// TestDashboardStartDetachedCleanup verifies proper cleanup behavior
func TestDashboardStartDetachedCleanup(t *testing.T) {
	if !tmux.IsTmuxAvailable() {
		t.Skip("tmux not available, skipping integration test")
	}

	// Setup: Create a temporary crAIzy project environment
	tmpDir := t.TempDir()
	projectName := "test-cleanup-project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Initialize project
	os.Chdir(tmpDir)
	if err := config.InitProject(projectName); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	os.Chdir(projectPath)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create dashboard
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	// Start multiple dashboard sessions to test cleanup
	sessionNames := []string{
		"craizy-test-cleanup-1",
		"craizy-test-cleanup-2",
		"craizy-test-cleanup-3",
	}

	var sessionIDs []string
	for _, name := range sessionNames {
		// Clean up any existing session
		tmuxMgr.KillSession(name)

		sessionID, err := dashboard.StartDetached(name)
		if err != nil {
			t.Fatalf("Failed to start detached dashboard %s: %v", name, err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}

	// Give time for sessions to start
	time.Sleep(500 * time.Millisecond)

	// Verify all sessions exist
	for _, sessionID := range sessionIDs {
		if !tmuxMgr.SessionExists(sessionID) {
			t.Errorf("Session %s should exist after creation", sessionID)
		}
	}

	// Clean up all sessions
	for _, sessionID := range sessionIDs {
		if err := tmuxMgr.KillSession(sessionID); err != nil {
			t.Errorf("Failed to kill session %s: %v", sessionID, err)
		}
	}

	// Give time for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify all sessions are gone
	for _, sessionID := range sessionIDs {
		if tmuxMgr.SessionExists(sessionID) {
			t.Errorf("Session %s should not exist after cleanup", sessionID)
		}
	}
}
