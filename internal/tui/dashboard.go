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
	width         int
	height        int
	sideMenu      SideMenuModel
	contentArea   ContentAreaModel
	quickCommands QuickCommandsModel
}

func NewModel() Model {
	return Model{
		sideMenu:      NewSideMenu(),
		contentArea:   NewContentArea(),
		quickCommands: NewQuickCommands(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.sideMenu.Init(),
		m.contentArea.Init(),
		m.quickCommands.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate dimensions
		bottomHeight := 5 // 3 lines text + 2 border
		mainHeight := m.height - bottomHeight
		if mainHeight < 0 {
			mainHeight = 0
		}

		sideWidth := int(float64(m.width) * 0.25)
		contentWidth := m.width - sideWidth

		m.sideMenu.SetSize(sideWidth, mainHeight)
		m.contentArea.SetSize(contentWidth, mainHeight)
		// Quick commands height is internal to the component style (3), 
		// but we pass it anyway or the component expects just width?
		// In my quick_commands.go I implemented SetSize to set m.height and View uses m.height.
		// Original style was Height(3). So I should pass 3.
		m.quickCommands.SetSize(m.width, 3)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// In a real app we would pass msg to children here
	// m.sideMenu, cmd = m.sideMenu.Update(msg)
	// cmds = append(cmds, cmd)
	// ...

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render sections
	sideView := m.sideMenu.View()
	contentView := m.contentArea.View()
	quickCommandsView := m.quickCommands.View()

	// Join layout
	// Top section: Side Menu + Content
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, sideView, contentView)

	// Full layout: Top Section + Quick Commands
	return lipgloss.JoinVertical(lipgloss.Left, topSection, quickCommandsView)
}
