package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuickCommandsModel struct {
	width         int
	height        int
	agentSelected bool
}

func NewQuickCommands() QuickCommandsModel {
	return QuickCommandsModel{}
}

func (m QuickCommandsModel) Init() tea.Cmd {
	return nil
}

func (m QuickCommandsModel) Update(msg tea.Msg) (QuickCommandsModel, tea.Cmd) {
	return m, nil
}

func (m *QuickCommandsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// SetAgentSelected updates whether an agent is currently selected.
func (m *QuickCommandsModel) SetAgentSelected(selected bool) {
	m.agentSelected = selected
}

func (m QuickCommandsModel) View() string {
	// Build context-aware hints
	hints := "n - new agent"
	if m.agentSelected {
		hints += " • enter - port to agent • m - merge agent • k - kill agent"
	}
	hints += " • q - quit"

	// Style: no border, grey text, centered horizontally, aligned to bottom
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")). // Grey text
		Width(m.width).
		Align(lipgloss.Center)

	// Align content to bottom of the allocated height
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom)

	return containerStyle.Render(textStyle.Render(hints))
}
