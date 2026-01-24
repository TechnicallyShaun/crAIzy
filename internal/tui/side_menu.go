package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SideMenuModel struct {
	width  int
	height int
}

func NewSideMenu() SideMenuModel {
	return SideMenuModel{}
}

func (m SideMenuModel) Init() tea.Cmd {
	return nil
}

func (m SideMenuModel) Update(msg tea.Msg) (SideMenuModel, tea.Cmd) {
	return m, nil
}

func (m *SideMenuModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SideMenuModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")). // Purple/Blue
		Width(m.width - 2).                     // Account for border
		Height(m.height - 2)

	return style.Render("Side Menu")
}
