package tui

import (
	"strings"
	"testing"
)

func TestContentAreaModel_AvailableLines(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		expected int
	}{
		{"standard height", 24, 22},
		{"small height", 10, 8},
		{"minimum height", 3, 1},
		{"zero height", 0, 1},
		{"negative would be clamped", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewContentArea()
			m.SetSize(80, tt.height)

			got := m.AvailableLines()

			if got != tt.expected {
				t.Errorf("AvailableLines() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestContentAreaModel_SetPreview(t *testing.T) {
	t.Run("sets preview content", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 24)

		content := "line1\nline2\nline3"
		m.SetPreview(content)

		// Preview should be stored
		if m.previewContent != content {
			t.Errorf("previewContent = %q, want %q", m.previewContent, content)
		}
	})

	t.Run("clears preview with empty string", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 24)

		m.SetPreview("some content")
		m.SetPreview("")

		if m.previewContent != "" {
			t.Errorf("previewContent should be empty after clear")
		}
	})
}

func TestContentAreaModel_View(t *testing.T) {
	t.Run("renders empty state when no preview", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 24)
		m.SetPreview("")

		view := m.View()

		// Should contain welcome message elements
		if !strings.Contains(view, "crAIzy") && !strings.Contains(view, "v0.1.0") {
			t.Error("empty state should show branded content")
		}
	})

	t.Run("renders preview when content set", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 24)
		m.SetPreview("test output line")

		view := m.View()

		if !strings.Contains(view, "test output line") {
			t.Error("view should contain preview content")
		}
	})
}

func TestContentAreaModel_availableWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected int
	}{
		{"standard width", 80, 78},
		{"small width", 20, 18},
		{"minimum width", 3, 1},
		{"zero width", 0, 1},
		{"width of 1", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewContentArea()
			m.SetSize(tt.width, 24)

			got := m.availableWidth()

			if got != tt.expected {
				t.Errorf("availableWidth() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		maxWidth int
		expected string
	}{
		{"short line unchanged", "hello", 10, "hello"},
		{"exact length unchanged", "hello", 5, "hello"},
		{"long line truncated", "hello world", 5, "hello"},
		{"zero width", "hello", 0, ""},
		{"negative width", "hello", -1, ""},
		{"unicode truncation", "héllo wörld", 5, "héllo"},
		{"emoji truncation", "👋🌍🎉", 2, "👋🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateLine(tt.line, tt.maxWidth)

			if got != tt.expected {
				t.Errorf("truncateLine(%q, %d) = %q, want %q", tt.line, tt.maxWidth, got, tt.expected)
			}
		})
	}
}

func TestContentAreaModel_renderPreview(t *testing.T) {
	t.Run("truncates to available lines", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 10) // 8 available lines

		// Create 20 lines of content
		lines := make([]string, 20)
		for i := 0; i < 20; i++ {
			lines[i] = "line"
		}
		m.SetPreview(strings.Join(lines, "\n"))

		rendered := m.renderPreview()
		renderedLines := strings.Split(rendered, "\n")

		// Should only have 8 lines (the last 8)
		if len(renderedLines) != 8 {
			t.Errorf("rendered %d lines, want 8", len(renderedLines))
		}
	})

	t.Run("shows all lines when fewer than available", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 24) // 22 available lines

		m.SetPreview("line1\nline2\nline3")

		rendered := m.renderPreview()
		renderedLines := strings.Split(rendered, "\n")

		if len(renderedLines) != 3 {
			t.Errorf("rendered %d lines, want 3", len(renderedLines))
		}
	})

	t.Run("truncates long lines to available width", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(20, 10) // 18 available width

		longLine := "this is a very long line that should be truncated"
		m.SetPreview(longLine)

		rendered := m.renderPreview()

		if len([]rune(rendered)) > 18 {
			t.Errorf("rendered line has %d chars, want max 18", len([]rune(rendered)))
		}
	})

	t.Run("responds to size changes", func(t *testing.T) {
		m := NewContentArea()
		longLine := "this is a very long line that should be truncated differently at different sizes"
		m.SetPreview(longLine)

		// Start with wide width
		m.SetSize(50, 10) // 48 available width
		rendered1 := m.renderPreview()

		// Resize to narrow width
		m.SetSize(20, 10) // 18 available width
		rendered2 := m.renderPreview()

		// Content should be shorter after resize
		if len([]rune(rendered2)) >= len([]rune(rendered1)) {
			t.Errorf("narrower width should produce shorter content: got %d >= %d",
				len([]rune(rendered2)), len([]rune(rendered1)))
		}

		// Verify it respects the new width
		if len([]rune(rendered2)) > 18 {
			t.Errorf("after resize, rendered line has %d chars, want max 18", len([]rune(rendered2)))
		}
	})
}

func TestContentAreaModel_renderEmptyState(t *testing.T) {
	t.Run("contains tagline", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 30)

		emptyState := m.renderEmptyState()

		if !strings.Contains(emptyState, "Using Artificial Intelligence") {
			t.Error("empty state should contain tagline")
		}
		if !strings.Contains(emptyState, "You must be") {
			t.Error("empty state should contain 'You must be'")
		}
	})

	t.Run("contains version", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(80, 30)

		emptyState := m.renderEmptyState()

		if !strings.Contains(emptyState, version) {
			t.Errorf("empty state should contain version %s", version)
		}
	})

	t.Run("handles very small size gracefully", func(t *testing.T) {
		m := NewContentArea()
		m.SetSize(5, 5) // Too small for content

		// Should not panic
		emptyState := m.renderEmptyState()

		// May be empty for very small sizes
		_ = emptyState
	})
}

func TestGenerateLogo(t *testing.T) {
	t.Run("generates non-empty logo", func(t *testing.T) {
		logo := generateLogo()

		if logo == "" {
			t.Error("logo should not be empty")
		}
	})

	t.Run("logo contains multiple lines", func(t *testing.T) {
		logo := generateLogo()
		lines := strings.Split(logo, "\n")

		// ASCII art should have multiple lines
		if len(lines) < 3 {
			t.Errorf("logo has %d lines, expected at least 3", len(lines))
		}
	})

	t.Run("logo has consistent structure", func(t *testing.T) {
		logo := generateLogo()
		lines := strings.Split(logo, "\n")

		// Logo should have some non-empty lines with actual content
		hasContent := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				hasContent = true
				break
			}
		}
		if !hasContent {
			t.Error("logo should have non-empty content lines")
		}
	})
}
