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

	tagline := "Using AI to write code?\nYou must be"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		tagline,
		logo,
		"\n(Press q to quit)",
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
