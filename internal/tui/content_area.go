package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ContentAreaModel struct {
	width  int
	height int
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

func (m ContentAreaModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("86")). // Cyan
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render("Content Area")
}
