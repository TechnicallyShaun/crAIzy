package tui

import (
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

const version = "v0.1.0"

// generateLogo creates the ASCII art logo using go-figure.
// Returns the logo with normalized whitespace for consistent alignment.
func generateLogo() string {
	fig := figure.NewFigure("crAIzy", "slant", true)
	raw := fig.String()

	// Trim trailing whitespace from each line and find the longest line
	lines := strings.Split(raw, "\n")
	maxLen := 0
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
		if len(lines[i]) > maxLen {
			maxLen = len(lines[i])
		}
	}

	// Find minimum leading whitespace across non-empty lines
	minLeading := maxLen
	for _, line := range lines {
		if line == "" {
			continue
		}
		leading := len(line) - len(strings.TrimLeft(line, " "))
		if leading < minLeading {
			minLeading = leading
		}
	}

	// Remove common leading whitespace from all lines
	for i, line := range lines {
		if len(line) >= minLeading {
			lines[i] = line[minLeading:]
		}
	}

	return strings.Join(lines, "\n")
}

type ContentAreaModel struct {
	width          int
	height         int
	previewContent string
}

func NewContentArea() ContentAreaModel {
	return ContentAreaModel{}
}

func (m ContentAreaModel) Init() tea.Cmd {
	return nil
}

func (m ContentAreaModel) Update(_ tea.Msg) (ContentAreaModel, tea.Cmd) {
	return m, nil
}

func (m *ContentAreaModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// SetPreview updates the preview content to display.
func (m *ContentAreaModel) SetPreview(content string) {
	m.previewContent = content
}

// AvailableLines returns the number of lines available for preview content.
// Accounts for border (2 lines).
func (m ContentAreaModel) AvailableLines() int {
	available := m.height - 2
	if available < 1 {
		return 1
	}
	return available
}

func (m ContentAreaModel) View() string {
	borderStyle := theme.BorderNormal.
		Width(m.width - 2).
		Height(m.height - 2)

	if m.previewContent == "" {
		return borderStyle.Render(m.renderEmptyState())
	}

	return borderStyle.Render(m.renderPreview())
}

// renderEmptyState renders the branded welcome screen.
func (m ContentAreaModel) renderEmptyState() string {
	// Available space inside border
	innerWidth := m.width - 4
	innerHeight := m.height - 4

	if innerWidth < 10 || innerHeight < 10 {
		return ""
	}

	// Style for tagline
	taglineStyle := theme.ContentTagline.
		Align(lipgloss.Center).
		Width(innerWidth)

	// Style for logo - no centering, we'll pad manually
	logoStyle := theme.ContentLogo

	// Style for version
	versionStyle := theme.ContentVersion.
		Align(lipgloss.Center).
		Width(innerWidth)

	// Build content
	tagline := taglineStyle.Render("Using Artificial Intelligence for coding?\nYou must be")
	asciiLogo := generateLogo()

	// Center the logo block manually by adding left padding
	logoLines := strings.Split(asciiLogo, "\n")
	logoWidth := 0
	for _, line := range logoLines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoPadding := (innerWidth - logoWidth) / 2
	if logoPadding < 0 {
		logoPadding = 0
	}
	paddedLogo := make([]string, len(logoLines))
	for i, line := range logoLines {
		paddedLogo[i] = strings.Repeat(" ", logoPadding) + line
	}
	logo := logoStyle.Render(strings.Join(paddedLogo, "\n"))

	ver := versionStyle.Render(version)

	// Calculate vertical spacing
	contentLines := strings.Count(tagline, "\n") + 1 +
		strings.Count(asciiLogo, "\n") + 1 +
		1 // version line

	topPadding := (innerHeight - contentLines) / 3
	if topPadding < 0 {
		topPadding = 0
	}

	// Build with vertical centering
	var builder strings.Builder
	for i := 0; i < topPadding; i++ {
		builder.WriteString("\n")
	}
	builder.WriteString(tagline)
	builder.WriteString("\n")
	builder.WriteString(logo)
	builder.WriteString("\n\n")
	builder.WriteString(ver)

	return builder.String()
}

// availableWidth returns the number of characters available per line.
// Accounts for border (2 chars).
func (m ContentAreaModel) availableWidth() int {
	available := m.width - 2
	if available < 1 {
		return 1
	}
	return available
}

// truncateLine truncates a line to fit within maxWidth.
// Uses rune-aware truncation to handle multi-byte characters.
func truncateLine(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(line)
	if len(runes) <= maxWidth {
		return line
	}
	return string(runes[:maxWidth])
}

// renderPreview renders the tmux pane output.
func (m ContentAreaModel) renderPreview() string {
	lines := strings.Split(m.previewContent, "\n")
	availableLines := m.AvailableLines()
	availableWidth := m.availableWidth()

	// Take the last N lines that fit
	if len(lines) > availableLines {
		lines = lines[len(lines)-availableLines:]
	}

	// Truncate each line to fit width
	for i, line := range lines {
		lines[i] = truncateLine(line, availableWidth)
	}

	return strings.Join(lines, "\n")
}
