package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ASCII art for "crAIzy"
const logo = `
               _    ___            
   ___ _ __   / \  |_ _|_____   _  
  / __| '__| / _ \  | ||_  / | | | 
 | (__| |   / ___ \ | | / /| |_| | 
  \___|_|  /_/   \_\___\___|\__, | 
                            |___/  
`

type Model struct {
	width  int
	height int
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate dimensions
	// Bottom section: 3 lines of text + 2 border lines = 5 lines total
	bottomHeight := 5
	mainHeight := m.height - bottomHeight
	if mainHeight < 0 {
		mainHeight = 0
	}

	// Side menu: 25% of width
	sideWidth := int(float64(m.width) * 0.25)
	// Content takes remaining width
	contentWidth := m.width - sideWidth

	// Define styles with different border colors
	sideStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")). // Purple/Blue
		Width(sideWidth - 2). // Account for border
		Height(mainHeight - 2)

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("86")). // Cyan
		Width(contentWidth - 2).
		Height(mainHeight - 2)

	quickCommandsStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("202")). // Orange/Red
		Width(m.width - 2).
		Height(3) // 3 text lines high

	// Render sections
	sideView := sideStyle.Render("Side Menu")
	contentView := contentStyle.Render("Content Area")
	quickCommandsView := quickCommandsStyle.Render("q - quit")

	// Join layout
	// Top section: Side Menu + Content
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, sideView, contentView)

	// Full layout: Top Section + Quick Commands
	return lipgloss.JoinVertical(lipgloss.Left, topSection, quickCommandsView)
}
