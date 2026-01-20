package tmux

import (
	"os/exec"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.sessionPrefix != "craizy" {
		t.Errorf("Expected session prefix 'craizy', got '%s'", mgr.sessionPrefix)
	}

	if mgr.sessions == nil {
		t.Error("Sessions map should not be nil")
	}
}

func TestIsTmuxAvailable(t *testing.T) {
	// This test checks if tmux is available on the system
	// It's expected to pass in CI environments with tmux installed
	available := IsTmuxAvailable()
	
	// Try to verify with direct command
	cmd := exec.Command("which", "tmux")
	err := cmd.Run()
	
	if err == nil && !available {
		t.Error("tmux is in PATH but IsTmuxAvailable returned false")
	}
}

func TestGetTmuxVersion(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	version, err := GetTmuxVersion()
	if err != nil {
		t.Fatalf("GetTmuxVersion failed: %v", err)
	}

	if version == "" {
		t.Error("Version should not be empty")
	}

	// Version should start with "tmux"
	if len(version) < 4 {
		t.Errorf("Version string too short: %s", version)
	}
}

func TestCreateSession(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	sessionName := "test-session"
	command := "sleep 5"

	// Clean up any existing session
	exec.Command("tmux", "kill-session", "-t", mgr.sessionPrefix+"-"+sessionName).Run()
	defer exec.Command("tmux", "kill-session", "-t", mgr.sessionPrefix+"-"+sessionName).Run()

	session, err := mgr.CreateSession(sessionName, command)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session == nil {
		t.Fatal("Session should not be nil")
	}

	if session.Name != sessionName {
		t.Errorf("Expected session name %s, got %s", sessionName, session.Name)
	}

	if session.Command != command {
		t.Errorf("Expected command %s, got %s", command, session.Command)
	}

	if !session.Active {
		t.Error("Session should be active")
	}

	// Give session a moment to start
	time.Sleep(200 * time.Millisecond)

	// Verify session exists in tmux
	fullName := mgr.sessionPrefix + "-" + sessionName
	if !mgr.SessionExists(fullName) {
		t.Errorf("Session %s should exist in tmux", fullName)
	}
}

func TestCreateDuplicateSession(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	sessionName := "test-duplicate"
	command := "sleep 10"

	// Clean up
	fullName := mgr.sessionPrefix + "-" + sessionName
	exec.Command("tmux", "kill-session", "-t", fullName).Run()
	defer exec.Command("tmux", "kill-session", "-t", fullName).Run()

	// Create first session
	_, err := mgr.CreateSession(sessionName, command)
	if err != nil {
		t.Fatalf("First CreateSession failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Try to create duplicate
	_, err = mgr.CreateSession(sessionName, command)
	if err == nil {
		t.Error("CreateSession should fail for duplicate session")
	}
}

func TestSessionExists(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	sessionName := "test-exists"
	fullName := mgr.sessionPrefix + "-" + sessionName

	// Clean up
	exec.Command("tmux", "kill-session", "-t", fullName).Run()

	// Should not exist
	if mgr.SessionExists(fullName) {
		t.Error("Session should not exist before creation")
	}

	// Create session
	mgr.CreateSession(sessionName, "sleep 5")
	defer exec.Command("tmux", "kill-session", "-t", fullName).Run()

	time.Sleep(100 * time.Millisecond)

	// Should exist
	if !mgr.SessionExists(fullName) {
		t.Error("Session should exist after creation")
	}
}

func TestListSessions(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()

	// Initially empty
	sessions := mgr.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}

	// Create a session
	sessionName := "test-list"
	fullName := mgr.sessionPrefix + "-" + sessionName
	defer exec.Command("tmux", "kill-session", "-t", fullName).Run()

	mgr.CreateSession(sessionName, "sleep 5")
	time.Sleep(100 * time.Millisecond)

	// Should have one session
	sessions = mgr.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	if sessions[0].Name != sessionName {
		t.Errorf("Expected session name %s, got %s", sessionName, sessions[0].Name)
	}
}

func TestGetSessionContent(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	sessionName := "test-content"
	fullName := mgr.sessionPrefix + "-" + sessionName
	testOutput := "Hello crAIzy"

	// Clean up
	exec.Command("tmux", "kill-session", "-t", fullName).Run()
	defer exec.Command("tmux", "kill-session", "-t", fullName).Run()

	// Create session with echo command
	mgr.CreateSession(sessionName, "echo '"+testOutput+"'; sleep 2")
	time.Sleep(500 * time.Millisecond)

	// Get content
	content, err := mgr.GetSessionContent(fullName)
	if err != nil {
		t.Fatalf("GetSessionContent failed: %v", err)
	}

	if content == "" {
		t.Error("Content should not be empty")
	}
}

func TestGetSessionContentNonExistent(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	_, err := mgr.GetSessionContent("non-existent-session")
	if err == nil {
		t.Error("GetSessionContent should fail for non-existent session")
	}
}

func TestKillSession(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	sessionName := "test-kill"
	fullName := mgr.sessionPrefix + "-" + sessionName

	// Clean up
	exec.Command("tmux", "kill-session", "-t", fullName).Run()

	// Create session
	mgr.CreateSession(sessionName, "sleep 10")
	time.Sleep(100 * time.Millisecond)

	// Verify it exists
	if !mgr.SessionExists(fullName) {
		t.Fatal("Session should exist before kill")
	}

	// Kill it
	err := mgr.KillSession(fullName)
	if err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Verify it's gone
	if mgr.SessionExists(fullName) {
		t.Error("Session should not exist after kill")
	}
}

func TestKillNonExistentSession(t *testing.T) {
	if !IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	mgr := NewManager()
	err := mgr.KillSession("non-existent-session")
	if err == nil {
		t.Error("KillSession should fail for non-existent session")
	}
}
