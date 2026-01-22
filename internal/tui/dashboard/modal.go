package dashboard

import (
	"fmt"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Modal represents a modal dialog for agent selection and instance naming
type Modal struct {
	agents       []config.Agent
	cursor       int
	active       bool
	selected     config.Agent
	instanceName string
}

// NewModal creates a new modal dialog
func NewModal(agents []config.Agent) Modal {
	return Modal{
		agents: agents,
		cursor: 0,
		active: false,
	}
}

// Show displays the modal
func (m *Modal) Show() {
	m.active = true
	m.cursor = 0
	m.selected = config.Agent{}
	m.instanceName = ""
}

// Hide hides the modal
func (m *Modal) Hide() {
	m.active = false
	m.selected = config.Agent{}
	m.instanceName = ""
}

// IsActive returns whether the modal is currently active
func (m Modal) IsActive() bool {
	return m.active
}

// Update handles modal key events
func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.selected.Name == "" && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.selected.Name == "" && m.cursor < len(m.agents)-1 {
				m.cursor++
			}
		case "esc":
			m.Hide()
			return m, promptInstanceNameCmd() // signal cancel to close modal
		case "enter":
			if m.selected.Name == "" {
				m.selected = m.agents[m.cursor]
				return m, promptInstanceNameCmd()
			}
		default:
			if m.selected.Name != "" {
				return m.handleTextInput(msg)
			}
		}
	}

	return m, nil
}

// handleTextInput processes text input for instance naming
func (m Modal) handleTextInput(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.Type {
	case tea.KeyBackspace, tea.KeyDelete:
		if m.instanceName != "" {
			m.instanceName = m.instanceName[:len(m.instanceName)-1]
		}
	case tea.KeySpace:
		m.instanceName += " "
	case tea.KeyEnter:
		name := strings.TrimSpace(m.instanceName)
		if name != "" {
			return m, instanceValidatedCmd(name, m.selected)
		}
	case tea.KeyEsc:
		m.Hide()
		return m, promptInstanceNameCmd()
	default:
		if len(msg.Runes) > 0 {
			m.instanceName += string(msg.Runes)
		}
	}
	return m, nil
}

// View renders the modal
func (m Modal) View(width, height int) string {
	if !m.active {
		return ""
	}

	if m.selected.Name != "" {
		return m.renderInstancePrompt(width, height)
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2).
		Width(50).
		Background(lipgloss.Color("235"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Align(lipgloss.Center)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	title := titleStyle.Render("Select an Agent to Spawn")
	content := title + "\n\n"

	for i, agent := range m.agents {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%d. %s", cursor, i+1, agent.Name)
		if m.cursor == i {
			line = selectedStyle.Render(line)
		}
		content += line + "\n"
	}

	content += "\n"
	content += lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[↑/↓] Navigate  [Enter] Select  [Esc] Cancel")

	modal := modalStyle.Render(content)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

func (m Modal) renderInstancePrompt(width, height int) string {
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2).
		Width(60).
		Background(lipgloss.Color("235"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Align(lipgloss.Center)

	title := titleStyle.Render("Name your agent instance")
	content := title + "\n\n"
	content += "Agent: " + m.selected.Name + "\n\n"
	content += "Enter a unique name (e.g., fix-typo):\n"
	content += m.instanceName + "\n\n"
	content += lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("[type to edit]  [Enter] Confirm  [Esc] Cancel")

	modal := modalStyle.Render(content)
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

type promptInstanceMsg struct{}
type instanceValidatedMsg struct {
	Agent config.Agent
	Name  string
}

func promptInstanceNameCmd() tea.Cmd {
	return func() tea.Msg { return promptInstanceMsg{} }
}

func instanceValidatedCmd(name string, agent config.Agent) tea.Cmd {
	return func() tea.Msg {
		return instanceValidatedMsg{Agent: agent, Name: name}
	}
}
