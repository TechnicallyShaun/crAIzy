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
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("202")). // Orange/Red
		Width(m.width - 2).
		Height(m.height)

	// Build context-aware hints
	hints := "n - new agent"
	if m.agentSelected {
		hints += " • enter - port to agent • k - kill agent"
	}
	hints += " • q - quit"

	return style.Render(hints)
}
