package tui

import (
	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
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
	modal         Modal
	agentService  *domain.AgentService
}

func NewModel(agentService *domain.AgentService) Model {
	return Model{
		sideMenu:      NewSideMenu(),
		contentArea:   NewContentArea(),
		quickCommands: NewQuickCommands(),
		modal:         NewModal(),
		agentService:  agentService,
	}
}

func (m Model) Init() tea.Cmd {
	// Send initial agents update to populate the list
	return tea.Batch(
		m.sideMenu.Init(),
		m.contentArea.Init(),
		m.quickCommands.Init(),
		m.modal.Init(),
		m.refreshAgents(),
	)
}

// refreshAgents returns a command that sends an AgentsUpdatedMsg with current agents.
func (m Model) refreshAgents() tea.Cmd {
	return func() tea.Msg {
		if m.agentService == nil {
			return AgentsUpdatedMsg{Agents: []*domain.Agent{}}
		}
		return AgentsUpdatedMsg{Agents: m.agentService.List()}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case CloseModalMsg:
		_ = msg // Suppress unused variable error
		m.modal.Close()
		return m, nil

	case AgentSelectedMsg:
		// Transition to name input step
		nameInput := NewNameInput(msg.Agent, m.width, m.height)
		m.modal.Open(nameInput)
		return m, nil

	case AgentCreatedMsg:
		m.modal.Close()
		// Create the agent using the service
		if m.agentService != nil {
			_, err := m.agentService.Create(msg.Agent.Name, msg.CustomName, msg.Agent.Command)
			if err != nil {
				// TODO: Show error to user
				return m, nil
			}
		}
		return m, m.refreshAgents()

	case AgentsUpdatedMsg:
		// Update the side menu with new agents
		var cmd tea.Cmd
		m.sideMenu, cmd = m.sideMenu.Update(msg)
		cmds = append(cmds, cmd)
		// Update quick commands based on selection state
		m.quickCommands.SetAgentSelected(m.sideMenu.HasAgents())
		return m, tea.Batch(cmds...)

	case domain.AgentDetachedMsg:
		// Returned from tmux session, refresh the agent list
		return m, m.refreshAgents()
	}

	if m.modal.IsOpen() {
		if cmd, handled := m.modal.Update(msg); handled {
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.modal.SetSize(m.width, m.height)

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
		m.quickCommands.SetSize(m.width, 3)

	case tea.KeyMsg:
		// Don't process keys if modal is open
		if m.modal.IsOpen() {
			break
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "n":
			agents, err := config.LoadAgents("AGENTS.yml")
			if err == nil {
				selector := NewAgentSelector(agents, m.width/2, m.height/2)
				m.modal.Open(selector)
			}

		case "enter":
			// Attach to selected agent
			if agent := m.sideMenu.SelectedAgent(); agent != nil && m.agentService != nil {
				return m, m.agentService.Attach(agent.ID)
			}

		case "k":
			// Kill selected agent
			if agent := m.sideMenu.SelectedAgent(); agent != nil && m.agentService != nil {
				_ = m.agentService.Kill(agent.ID)
				return m, m.refreshAgents()
			}
		}

		// Forward arrow key navigation to side menu
		if msg.String() == "up" || msg.String() == "down" {
			var cmd tea.Cmd
			m.sideMenu, cmd = m.sideMenu.Update(msg)
			cmds = append(cmds, cmd)
			// Update quick commands after navigation
			m.quickCommands.SetAgentSelected(m.sideMenu.SelectedAgent() != nil)
		}
	}

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
	baseView := lipgloss.JoinVertical(lipgloss.Left, topSection, quickCommandsView)

	if m.modal.IsOpen() {
		return m.modal.View()
	}
	return baseView
}
