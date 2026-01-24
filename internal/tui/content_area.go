package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ASCII art logo with "AI" emphasized
const asciiLogo = `               ___   ___
  ___ _ __    / _ \ |_ _|  _____  _
 / __| '__|  / /_\ \ | |  |_  / | | |
| (__| |    /  _  | | |   / /| |_| |
 \___|_|    \_/ |_/|___|  /___|\__, |
                               |___/`

const version = "v0.1.0"

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

func (m ContentAreaModel) Update(msg tea.Msg) (ContentAreaModel, tea.Cmd) {
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
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("86")). // Cyan
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
	taglineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")). // Light gray
		Align(lipgloss.Center).
		Width(innerWidth)

	// Style for logo (cyan to match border)
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Align(lipgloss.Center).
		Width(innerWidth)

	// Style for version
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")). // Medium gray
		Align(lipgloss.Center).
		Width(innerWidth)

	// Build content
	tagline := taglineStyle.Render("Using Artificial Intelligence for coding?\nYou must be")
	logo := logoStyle.Render(asciiLogo)
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

// renderPreview renders the tmux pane output.
func (m ContentAreaModel) renderPreview() string {
	// Just return the content, trimming to fit
	lines := strings.Split(m.previewContent, "\n")
	available := m.AvailableLines()

	if len(lines) > available {
		lines = lines[len(lines)-available:]
	}

	return strings.Join(lines, "\n")
}
