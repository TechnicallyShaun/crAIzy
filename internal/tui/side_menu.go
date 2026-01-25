package tui

import (
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/tui/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentListItem implements list.Item for domain.Agent
type AgentListItem struct {
	agent *domain.Agent
}

func (i AgentListItem) Title() string {
	return i.agent.Name
}

func (i AgentListItem) Description() string {
	return i.agent.AgentType
}

func (i AgentListItem) FilterValue() string {
	return i.agent.Name
}

type SideMenuModel struct {
	width  int
	height int
	list   list.Model
	agents []*domain.Agent
}

func NewSideMenu() SideMenuModel {
	// Create delegate with minimal styling
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Agents"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return SideMenuModel{
		list:   l,
		agents: []*domain.Agent{},
	}
}

func (m SideMenuModel) Init() tea.Cmd {
	return nil
}

func (m SideMenuModel) Update(msg tea.Msg) (SideMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case AgentsUpdatedMsg:
		m.agents = msg.Agents
		items := make([]list.Item, len(m.agents))
		for i, agent := range m.agents {
			items[i] = AgentListItem{agent: agent}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		// Handle navigation within the list
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SideMenuModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	// Set list size to match panel
	m.list.SetWidth(w - 2)
	m.list.SetHeight(h - 2)
}

// SelectedAgent returns the currently selected agent, or nil if none selected.
func (m SideMenuModel) SelectedAgent() *domain.Agent {
	if len(m.agents) == 0 {
		return nil
	}
	if item, ok := m.list.SelectedItem().(AgentListItem); ok {
		return item.agent
	}
	return nil
}

// HasAgents returns true if there are agents in the list.
func (m SideMenuModel) HasAgents() bool {
	return len(m.agents) > 0
}

func (m SideMenuModel) View() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height)

	if len(m.agents) == 0 {
		emptyStyle := theme.SideMenuEmpty.Padding(1)
		return style.Render(emptyStyle.Render("No agents running\n\nPress 'n' to create one"))
	}

	return style.Render(m.list.View())
}
