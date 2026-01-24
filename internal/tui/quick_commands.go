package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuickCommandsModel struct {
	width  int
	height int
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

func (m QuickCommandsModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("202")). // Orange/Red
		Width(m.width - 2).
		Height(m.height) // Already accounted for in dashboard calculation? No, style needs explicit height.
	// In dashboard.go it was: Height(3) (fixed)
	// But let's allow dynamic resize if needed, though usually fixed. 
	// We'll trust the passed height.

	return style.Render("q - quit")
}
