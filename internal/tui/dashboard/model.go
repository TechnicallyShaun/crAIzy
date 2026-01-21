package dashboard

import (
	"fmt"
	"time"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	previewInterval = 2 * time.Second
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	// Preview pane styling
	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
)

// SessionManager defines the interface for interacting with tmux.
// This allows us to mock the tmux layer for testing and ensures the dashboard
// relies on behavior, not concrete implementations.
type SessionManager interface {
	CreateSession(name, command string) (*tmux.Session, error)
	ListSessions() []*tmux.Session
	SessionExists(name string) bool
	// SwitchClient switches the current client to the target session
	SwitchClient(target string) error
	// CapturePane returns the last N lines of the target session's output
	CapturePane(target string, lines int) (string, error)
}

// tickMsg is sent to the update loop to trigger a preview refresh
type tickMsg time.Time

// AgentItem adapts config.Agent to the bubbles/list.Item interface
type AgentItem struct {
	Name    string
	Command string
	Active  bool
}

func (i AgentItem) Title() string { return i.Name }
func (i AgentItem) Description() string {
	if i.Active {
		return "● Active"
	}
	return "○ Idle"
}
func (i AgentItem) FilterValue() string { return i.Name }

// Model represents the state of the dashboard
type Model struct {
	list           list.Model
	cfg            *config.Config
	tmux           SessionManager
	previewContent string
	width          int
	height         int
	quitting       bool
}

// NewModel initializes the dashboard model
func NewModel(cfg *config.Config, tm SessionManager) Model {
	items := make([]list.Item, len(cfg.Agents))
	activeSessions := make(map[string]bool)

	// Pre-fetch active sessions to set initial state
	if tm != nil {
		for _, s := range tm.ListSessions() {
			activeSessions[s.Name] = true
		}
	}

	for i, agent := range cfg.Agents {
		items[i] = AgentItem{
			Name:    agent.Name,
			Command: agent.Command,
			Active:  activeSessions[agent.Name],
		}
	}

	// Setup the list component
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "crAIzy Agents"
	l.SetShowHelp(false) // We render our own help in the footer if needed

	return Model{
		list: l,
		cfg:  cfg,
		tmux: tm,
	}
}

// Init starts the tick loop for live previews
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Attach to the selected agent's session
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				agent := selectedItem.(AgentItem)
				return m, attachToSessionCmd(m.tmux, agent)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Layout calculation: List takes 1/3, Preview takes 2/3
		listWidth := msg.Width / 3
		m.list.SetSize(listWidth, msg.Height-2)

		// Update styles with new dimensions
		previewStyle = previewStyle.
			Width(msg.Width - listWidth - 4). // Account for borders/margins
			Height(msg.Height - 2)

	case tickMsg:
		// Refresh the preview content for the selected agent
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			agent := selectedItem.(AgentItem)
			// Trigger a command to fetch logs
			cmds = append(cmds, fetchPreviewCmd(m.tmux, agent.Name))
		}
		// Schedule the next tick
		cmds = append(cmds, tickCmd())

	case previewResultMsg:
		// Update the view with the fetched content
		m.previewContent = string(msg)
	}

	// Update the list component
	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	leftView := docStyle.Render(m.list.View())

	// Right side: Preview Pane
	title := "PREVIEW: (None)"
	if i := m.list.SelectedItem(); i != nil {
		title = fmt.Sprintf("PREVIEW: %s", i.(AgentItem).Name)
	}

	rightContent := fmt.Sprintf("%s\n\n%s", title, m.previewContent)
	rightView := previewStyle.Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}

// --- Commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(previewInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// previewResultMsg carries the output of CapturePane
type previewResultMsg string

func fetchPreviewCmd(tm SessionManager, agentName string) tea.Cmd {
	return func() tea.Msg {
		sessionName := fmt.Sprintf("craizy-%s", agentName)
		if !tm.SessionExists(sessionName) {
			return previewResultMsg("Agent is offline or idle.")
		}

		content, err := tm.CapturePane(sessionName, 20)
		if err != nil {
			return previewResultMsg(fmt.Sprintf("Error fetching preview: %v", err))
		}
		return previewResultMsg(content)
	}
}

func attachToSessionCmd(tm SessionManager, agent AgentItem) tea.Cmd {
	return func() tea.Msg {
		sessionName := fmt.Sprintf("craizy-%s", agent.Name)

		// Ensure session exists
		if !tm.SessionExists(sessionName) {
			_, err := tm.CreateSession(agent.Name, agent.Command)
			if err != nil {
				// In a real app we might send an error msg, here we just fail silently or log
				return nil
			}
		}

		// Switch to it
		_ = tm.SwitchClient(sessionName)
		return tea.Quit // We quit the dashboard to let tmux handle the switch
	}
}
