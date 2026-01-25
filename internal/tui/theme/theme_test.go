package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorsAreDefined(t *testing.T) {
	// Verify all base colors are defined and non-empty
	colors := []struct {
		name  string
		color lipgloss.Color
	}{
		{"ColorBackground", ColorBackground},
		{"ColorForeground", ColorForeground},
		{"ColorMuted", ColorMuted},
		{"ColorPrimary", ColorPrimary},
		{"ColorSecondary", ColorSecondary},
		{"ColorBorder", ColorBorder},
		{"ColorSuccess", ColorSuccess},
		{"ColorWarning", ColorWarning},
		{"ColorError", ColorError},
		{"ColorSpecial", ColorSpecial},
	}

	for _, c := range colors {
		if string(c.color) == "" {
			t.Errorf("%s should not be empty", c.name)
		}
	}
}

func TestStylesAreInitialized(t *testing.T) {
	// Test that text styles can render without panic
	testCases := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TextNormal", TextNormal},
		{"TextMuted", TextMuted},
		{"TextSuccess", TextSuccess},
		{"TextWarning", TextWarning},
		{"TextError", TextError},
		{"BorderNormal", BorderNormal},
		{"BorderFocused", BorderFocused},
		{"BorderRounded", BorderRounded},
		{"SideMenuTitle", SideMenuTitle},
		{"SideMenuEmpty", SideMenuEmpty},
		{"AgentRunning", AgentRunning},
		{"AgentStopped", AgentStopped},
		{"AgentPending", AgentPending},
		{"ContentTitle", ContentTitle},
		{"ContentSubtitle", ContentSubtitle},
		{"ContentLogo", ContentLogo},
		{"ContentVersion", ContentVersion},
		{"ContentTagline", ContentTagline},
		{"ModalTitle", ModalTitle},
		{"ModalBorder", ModalBorder},
		{"QuickCommandKey", QuickCommandKey},
		{"QuickCommandDesc", QuickCommandDesc},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should render without panic
			result := tc.style.Render("test")
			if result == "" {
				t.Errorf("%s.Render() returned empty string", tc.name)
			}
		})
	}
}

func TestTmuxStatusBarColors(t *testing.T) {
	// Verify tmux colors are defined
	if TmuxStatusBar.Background == "" {
		t.Error("TmuxStatusBar.Background should not be empty")
	}
	if TmuxStatusBar.Foreground == "" {
		t.Error("TmuxStatusBar.Foreground should not be empty")
	}
	if TmuxStatusBar.BrandColor == "" {
		t.Error("TmuxStatusBar.BrandColor should not be empty")
	}
	if TmuxStatusBar.AccentColor == "" {
		t.Error("TmuxStatusBar.AccentColor should not be empty")
	}
	if TmuxStatusBar.MutedColor == "" {
		t.Error("TmuxStatusBar.MutedColor should not be empty")
	}
	if TmuxStatusBar.SeparatorColor == "" {
		t.Error("TmuxStatusBar.SeparatorColor should not be empty")
	}
}

func TestNordPaletteValues(t *testing.T) {
	// Verify we're using the correct Nord ANSI 256 values
	expectedColors := map[string]string{
		"ColorBackground": "235",
		"ColorForeground": "255",
		"ColorMuted":      "243",
		"ColorPrimary":    "110",
		"ColorSecondary":  "111",
		"ColorBorder":     "68",
		"ColorSuccess":    "108",
		"ColorWarning":    "222",
		"ColorError":      "174",
		"ColorSpecial":    "139",
	}

	actualColors := map[string]lipgloss.Color{
		"ColorBackground": ColorBackground,
		"ColorForeground": ColorForeground,
		"ColorMuted":      ColorMuted,
		"ColorPrimary":    ColorPrimary,
		"ColorSecondary":  ColorSecondary,
		"ColorBorder":     ColorBorder,
		"ColorSuccess":    ColorSuccess,
		"ColorWarning":    ColorWarning,
		"ColorError":      ColorError,
		"ColorSpecial":    ColorSpecial,
	}

	for name, expected := range expectedColors {
		actual := string(actualColors[name])
		if actual != expected {
			t.Errorf("%s: expected %q, got %q", name, expected, actual)
		}
	}
}

func TestBorderStylesHaveBorders(t *testing.T) {
	// Verify border styles actually have borders configured
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"BorderNormal", BorderNormal},
		{"BorderFocused", BorderFocused},
		{"BorderRounded", BorderRounded},
		{"ModalBorder", ModalBorder},
	}

	for _, s := range styles {
		// Render a simple string and check it has more characters than the input
		// (indicating borders were added)
		input := "X"
		result := s.style.Render(input)
		if len(result) <= len(input) {
			t.Errorf("%s should add border characters", s.name)
		}
	}
}
