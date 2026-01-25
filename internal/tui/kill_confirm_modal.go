package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KillConfirmModel is a modal that confirms killing an agent with uncommitted changes.
type KillConfirmModel struct {
	sessionID string
	agentName string
	width     int
	height    int
	selected  int // 0 = Keep, 1 = Discard, 2 = Cancel
}

// NewKillConfirmModal creates a new kill confirmation modal.
func NewKillConfirmModal(sessionID, agentName string, width, height int) KillConfirmModel {
	return KillConfirmModel{
		sessionID: sessionID,
		agentName: agentName,
		width:     width,
		height:    height,
		selected:  2, // Default to Cancel for safety
	}
}

func (m KillConfirmModel) Init() tea.Cmd {
	return nil
}

func (m KillConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < 2 {
				m.selected++
			}
		case "enter":
			var choice KillConfirmChoice
			switch m.selected {
			case 0:
				choice = KillConfirmKeep
			case 1:
				choice = KillConfirmDiscard
			case 2:
				choice = KillConfirmCancel
			}
			return m, func() tea.Msg {
				return KillConfirmResultMsg{
					SessionID: m.sessionID,
					Choice:    choice,
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		}
	}
	return m, nil
}

func (m KillConfirmModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208"))

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())

	selectedStyle := buttonStyle.
		BorderForeground(lipgloss.Color("205")).
		Bold(true)

	unselectedStyle := buttonStyle.
		BorderForeground(lipgloss.Color("240"))

	title := titleStyle.Render("Kill Agent: " + m.agentName)
	warning := warningStyle.Render("This agent has uncommitted changes!")

	// Render buttons
	keepStyle := unselectedStyle
	discardStyle := unselectedStyle
	cancelStyle := unselectedStyle

	switch m.selected {
	case 0:
		keepStyle = selectedStyle
	case 1:
		discardStyle = selectedStyle
	case 2:
		cancelStyle = selectedStyle
	}

	keepBtn := keepStyle.Render("Keep (Stash)")
	discardBtn := discardStyle.Render("Discard")
	cancelBtn := cancelStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, keepBtn, " ", discardBtn, " ", cancelBtn)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("Use arrow keys to select, Enter to confirm")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		warning,
		"",
		buttons,
		"",
		hint,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 3).
		BorderForeground(lipgloss.Color("63")).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
