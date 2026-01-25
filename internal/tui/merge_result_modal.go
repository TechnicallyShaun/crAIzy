package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MergeResultModel is a modal that shows the result of a merge operation.
type MergeResultModel struct {
	agentName   string
	success     bool
	stashed     bool
	conflictErr error
	width       int
	height      int
}

// NewMergeResultModal creates a new merge result modal.
func NewMergeResultModal(agentName string, success, stashed bool, conflictErr error, width, height int) MergeResultModel {
	return MergeResultModel{
		agentName:   agentName,
		success:     success,
		stashed:     stashed,
		conflictErr: conflictErr,
		width:       width,
		height:      height,
	}
}

func (m MergeResultModel) Init() tea.Cmd {
	return nil
}

func (m MergeResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", " ":
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		}
	}
	return m, nil
}

func (m MergeResultModel) View() string {
	var title, message, hint string

	titleStyle := lipgloss.NewStyle().Bold(true)
	messageStyle := lipgloss.NewStyle()

	if m.success {
		titleStyle = titleStyle.Foreground(lipgloss.Color("42")) // Green
		title = titleStyle.Render("Merge Successful")
		message = messageStyle.Render("Branch from " + m.agentName + " has been merged.")
		if m.stashed {
			message += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Render("(Your stashed changes have been restored)")
		}
	} else {
		titleStyle = titleStyle.Foreground(lipgloss.Color("196")) // Red
		title = titleStyle.Render("Merge Failed")

		errMsg := "Unknown error"
		if m.conflictErr != nil {
			errMsg = "Merge conflict detected"
		}
		message = messageStyle.Render("Failed to merge branch from " + m.agentName + ":\n" + errMsg)

		if m.stashed {
			message += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Render("(Your stashed changes have been restored)")
		}

		message += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Render("Resolve conflicts manually with:\n  git merge --abort  (to cancel)\n  git add . && git merge --continue  (after resolving)")
	}

	hint = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render("Press Enter to close")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		message,
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
