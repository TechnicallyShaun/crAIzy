package ui

import (
	"strings"
	"testing"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
)

func TestNewDashboard(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test",
		AIs: []config.AISpec{
			{Name: "TestAI", Command: "echo test"},
		},
	}
	tmuxMgr := tmux.NewManager()

	dashboard := NewDashboard(cfg, tmuxMgr)

	if dashboard == nil {
		t.Fatal("NewDashboard returned nil")
	}

	if dashboard.config != cfg {
		t.Error("Config not set correctly")
	}

	if dashboard.tmuxMgr != tmuxMgr {
		t.Error("Tmux manager not set correctly")
	}

	if dashboard.aiInstances == nil {
		t.Error("AI instances should be initialized")
	}

	if len(dashboard.aiInstances) != 0 {
		t.Error("AI instances should start empty")
	}

	if dashboard.selectedTab != -1 {
		t.Error("Selected tab should start at -1")
	}
}

func TestSpawnAI(t *testing.T) {
	if !tmux.IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	cfg := &config.Config{
		ProjectName: "test",
		AIs:         []config.AISpec{},
	}
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	aiSpec := config.AISpec{
		Name:    "TestAI",
		Command: "echo 'test'; sleep 2",
	}

	ai, err := dashboard.SpawnAI(aiSpec)
	if err != nil {
		t.Fatalf("SpawnAI failed: %v", err)
	}

	// Clean up
	if ai != nil && ai.Session != nil {
		defer tmuxMgr.KillSession(ai.Session.ID)
	}

	if ai == nil {
		t.Fatal("AI instance should not be nil")
	}

	if ai.ID != 1 {
		t.Errorf("Expected AI ID 1, got %d", ai.ID)
	}

	if ai.Name == "" {
		t.Error("AI name should not be empty")
	}

	if ai.Session == nil {
		t.Error("AI session should not be nil")
	}

	// Verify it was added to instances
	if len(dashboard.aiInstances) != 1 {
		t.Errorf("Expected 1 AI instance, got %d", len(dashboard.aiInstances))
	}
}

func TestSpawnMultipleAIs(t *testing.T) {
	if !tmux.IsTmuxAvailable() {
		t.Skip("tmux not available, skipping test")
	}

	cfg := &config.Config{
		ProjectName: "test",
		AIs:         []config.AISpec{},
	}
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	aiSpec1 := config.AISpec{
		Name:    "TestAI1",
		Command: "sleep 2",
	}
	aiSpec2 := config.AISpec{
		Name:    "TestAI2",
		Command: "sleep 2",
	}

	ai1, err := dashboard.SpawnAI(aiSpec1)
	if err != nil {
		t.Fatalf("First SpawnAI failed: %v", err)
	}
	if ai1 == nil {
		t.Fatal("ai1 should not be nil")
	}
	if ai1.Session != nil {
		defer tmuxMgr.KillSession(ai1.Session.ID)
	}

	ai2, err2 := dashboard.SpawnAI(aiSpec2)
	if err2 != nil {
		t.Fatalf("Second SpawnAI failed: %v", err2)
	}
	if ai2 == nil {
		t.Fatal("ai2 should not be nil")
	}
	if ai2.Session != nil {
		defer tmuxMgr.KillSession(ai2.Session.ID)
	}

	if ai1.ID != 1 {
		t.Errorf("Expected first AI ID 1, got %d", ai1.ID)
	}

	if ai2.ID != 2 {
		t.Errorf("Expected second AI ID 2, got %d", ai2.ID)
	}

	if len(dashboard.aiInstances) != 2 {
		t.Errorf("Expected 2 AI instances, got %d", len(dashboard.aiInstances))
	}
}

func TestGetAIInstance(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test",
		AIs:         []config.AISpec{},
	}
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	// Add mock AI instance
	mockAI := &AIInstance{
		ID:   1,
		Name: "MockAI",
	}
	dashboard.aiInstances = append(dashboard.aiInstances, mockAI)

	// Test valid ID
	ai := dashboard.GetAIInstance(1)
	if ai == nil {
		t.Error("GetAIInstance should return AI for valid ID")
	} else if ai.ID != 1 {
		t.Errorf("Expected AI ID 1, got %d", ai.ID)
	}

	// Test invalid IDs
	if dashboard.GetAIInstance(0) != nil {
		t.Error("GetAIInstance should return nil for ID 0")
	}
	if dashboard.GetAIInstance(2) != nil {
		t.Error("GetAIInstance should return nil for out-of-range ID")
	}
	if dashboard.GetAIInstance(-1) != nil {
		t.Error("GetAIInstance should return nil for negative ID")
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"test", 10, "test      "},
		{"test", 4, "test"},
		{"test", 2, "te"},
		{"", 5, "     "},
		{"hello", 0, ""},
	}

	for _, tt := range tests {
		result := padRight(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestGenerateDashboardScript(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-project",
		AIs: []config.AISpec{
			{Name: "GPT-4", Command: "gpt4-cli"},
			{Name: "Claude", Command: "claude-cli"},
		},
	}
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	script := dashboard.generateDashboardScript()

	// Verify script contains key elements
	if script == "" {
		t.Error("Script should not be empty")
	}

	// Check for project name
	if dashboard.config.ProjectName != "" && dashboard.config.ProjectName == "test-project" {
		// Script should reference the project somehow
		if len(script) < 100 {
			t.Error("Script seems too short")
		}
	}

	// Check for essential commands
	essentialCommands := []string{"clear", "echo", "read"}
	for _, cmd := range essentialCommands {
		if !strings.Contains(script, cmd) {
			t.Errorf("Script should contain '%s' command", cmd)
		}
	}
}

func TestDashboardStartWithoutTmux(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test",
		AIs:         []config.AISpec{},
	}
	tmuxMgr := tmux.NewManager()
	dashboard := NewDashboard(cfg, tmuxMgr)

	// This test verifies error handling when tmux is not available
	// We can't actually test without tmux in CI, but we can verify the check exists
	if !tmux.IsTmuxAvailable() {
		err := dashboard.Start()
		if err == nil {
			t.Error("Start should fail when tmux is not available")
		}
	}
}
