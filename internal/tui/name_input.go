package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tui/theme"
)

type NameInputModel struct {
	textInput     textinput.Model
	selectedAgent config.Agent
	width         int
	height        int
}

func NewNameInput(agent config.Agent, width, height int) NameInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter a name for this session"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	return NameInputModel{
		textInput:     ti,
		selectedAgent: agent,
		width:         width,
		height:        height,
	}
}

func (m NameInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m NameInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, func() tea.Msg {
				return AgentCreatedMsg{
					Agent:      m.selectedAgent,
					CustomName: m.textInput.Value(),
				}
			}
		case tea.KeyEsc:
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m NameInputModel) View() string {
	title := theme.ModalTitle.
		Render("Name your " + m.selectedAgent.Name + " Agent")

	input := m.textInput.View()

	box := theme.ModalBorder.
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				title,
				"\n",
				input,
			),
		)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
