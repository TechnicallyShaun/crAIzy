package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MergeResultModel is a modal that shows the result of a merge operation.
type MergeResultModel struct {
	agentName     string
	agentID       string
	success       bool
	stashed       bool
	conflictErr   error
	conflictFiles []string
	baseBranch    string
	width         int
	height        int
	selectedIdx   int // 0 = Send to Terminal, 1 = Cancel
}

// NewMergeResultModal creates a new merge result modal.
func NewMergeResultModal(agentName, agentID string, success, stashed bool, conflictErr error, conflictFiles []string, baseBranch string, width, height int) MergeResultModel {
	return MergeResultModel{
		agentName:     agentName,
		agentID:       agentID,
		success:       success,
		stashed:       stashed,
		conflictErr:   conflictErr,
		conflictFiles: conflictFiles,
		baseBranch:    baseBranch,
		width:         width,
		height:        height,
		selectedIdx:   0,
	}
}

func (m MergeResultModel) Init() tea.Cmd {
	return nil
}

func (m MergeResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// For successful merges, just close on any key
		if m.success {
			switch msg.String() {
			case "enter", "esc", " ":
				return m, func() tea.Msg {
					return CloseModalMsg{}
				}
			}
			return m, nil
		}

		// For conflicts, handle option selection
		switch msg.String() {
		case "left", "h", "shift+tab":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "right", "l", "tab":
			if m.selectedIdx < 1 {
				m.selectedIdx++
			}
		case "enter", " ":
			choice := MergeConflictCancel
			if m.selectedIdx == 0 {
				choice = MergeConflictSendToTerminal
			}
			return m, func() tea.Msg {
				return MergeConflictResultMsg{
					AgentID:       m.agentID,
					BaseBranch:    m.baseBranch,
					ConflictFiles: m.conflictFiles,
					Choice:        choice,
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return MergeConflictResultMsg{
					AgentID:       m.agentID,
					BaseBranch:    m.baseBranch,
					ConflictFiles: m.conflictFiles,
					Choice:        MergeConflictCancel,
				}
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
		hint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render("Press Enter to close")
	} else {
		titleStyle = titleStyle.Foreground(lipgloss.Color("196")) // Red
		title = titleStyle.Render("Merge Failed")

		errMsg := "Unknown error"
		if m.conflictErr != nil {
			errMsg = "Merge conflict detected"
		}
		message = messageStyle.Render("Failed to merge branch from " + m.agentName + ":\n" + errMsg)

		// Show conflict files if available
		if len(m.conflictFiles) > 0 {
			fileList := strings.Join(m.conflictFiles, ", ")
			if len(fileList) > 60 {
				fileList = fileList[:57] + "..."
			}
			message += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("208")).
				Render(fmt.Sprintf("Conflicting files: %s", fileList))
		}

		if m.stashed {
			message += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Render("(Your stashed changes have been restored)")
		}

		// Build option buttons
		sendStyle := lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder())
		cancelStyle := lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder())

		if m.selectedIdx == 0 {
			sendStyle = sendStyle.
				BorderForeground(lipgloss.Color("42")).
				Foreground(lipgloss.Color("42"))
		} else {
			sendStyle = sendStyle.
				BorderForeground(lipgloss.Color("245")).
				Foreground(lipgloss.Color("245"))
		}

		if m.selectedIdx == 1 {
			cancelStyle = cancelStyle.
				BorderForeground(lipgloss.Color("196")).
				Foreground(lipgloss.Color("196"))
		} else {
			cancelStyle = cancelStyle.
				BorderForeground(lipgloss.Color("245")).
				Foreground(lipgloss.Color("245"))
		}

		sendBtn := sendStyle.Render("Send to Terminal")
		cancelBtn := cancelStyle.Render("Cancel")

		buttons := lipgloss.JoinHorizontal(lipgloss.Center, sendBtn, "  ", cancelBtn)

		hint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render("Use ←/→ to select, Enter to confirm")

		content := lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			message,
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
