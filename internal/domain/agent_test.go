package domain

import "testing"

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Path 1: Basic lowercase conversion
		{"uppercase", "MyProject", "myproject"},

		// Path 2: Period removal
		{"periods", "file.name.txt", "filenametxt"},

		// Path 3: Colon removal
		{"colons", "time:12:30", "time1230"},

		// Path 4: Space to hyphen
		{"spaces", "my project name", "my-project-name"},

		// Path 5: Non-alphanumeric removal (regex path)
		{"special chars", "project@#$%test", "projecttest"},

		// Path 6: Consecutive hyphen collapse (while loop)
		{"consecutive hyphens", "my---name", "my-name"},

		// Path 7: Leading/trailing hyphen trim
		{"trim hyphens", "-project-", "project"},

		// Path 8: Combined transformations
		{"combined", "My Project...V2:Final!", "my-projectv2final"},

		// Path 9: Empty after sanitization
		{"all special", "@#$%^&*()", ""},

		// Path 10: Already clean
		{"clean", "myproject123", "myproject123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildSessionID(t *testing.T) {
	tests := []struct {
		name      string
		project   string
		agentType string
		agentName string
		expected  string
	}{
		// Path 1: Clean inputs
		{"clean", "myproject", "claude", "task1", "craizy-myproject-claude-task1"},

		// Path 2: Inputs requiring sanitization
		{"dirty", "My Project", "Claude Code", "Task #1", "craizy-my-project-claude-code-task-1"},

		// Path 3: Empty name component
		{"empty name", "project", "agent", "", "craizy-project-agent-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSessionID(tt.project, tt.agentType, tt.agentName)
			if got != tt.expected {
				t.Errorf("BuildSessionID(%q, %q, %q) = %q, want %q",
					tt.project, tt.agentType, tt.agentName, got, tt.expected)
			}
		})
	}
}
