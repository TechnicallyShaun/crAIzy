package tui

import (
	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AgentItem struct {
	agent config.Agent
}

func (i AgentItem) Title() string       { return i.agent.Name }
func (i AgentItem) Description() string { return i.agent.Command }
func (i AgentItem) FilterValue() string { return i.agent.Name }

type AgentSelectorModel struct {
	list   list.Model
	width  int
	height int
}

func NewAgentSelector(agents []config.Agent, width, height int) AgentSelectorModel {
	items := make([]list.Item, len(agents))
	for i, a := range agents {
		items[i] = AgentItem{agent: a}
	}

	// Adjust dimensions for the list
	// Modal usually has some padding/border, let's give the list some room
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "Select an Agent"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false) // Simple selection for now
	l.KeyMap.Quit.SetEnabled(false) // Prevent 'q' from quitting - handled by dashboard only

	return AgentSelectorModel{
		list:   l,
		width:  width,
		height: height,
	}
}

func (m AgentSelectorModel) Init() tea.Cmd {
	return nil
}

func (m AgentSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(AgentItem); ok {
				return m, func() tea.Msg {
					return AgentSelectedMsg{Agent: i.agent}
				}
			}
		}
		if msg.String() == "esc" {
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AgentSelectorModel) View() string {
	return lipgloss.NewStyle().
		Margin(1, 2).
		Render(m.list.View())
}
