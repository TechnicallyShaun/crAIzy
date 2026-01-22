package dashboard

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
	"github.com/TechnicallyShaun/crAIzy/internal/worktree"
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
	CreateSession(name, command, cwd string) (*tmux.Session, error)
	ListSessions() []*tmux.Session
	SessionExists(name string) bool
	KillSession(name string) error
	// SwitchClient switches the current client to the target session
	SwitchClient(target string) error
	// CapturePane returns the last N lines of the target session's output
	CapturePane(target string, lines int) (string, error)
}

// tickMsg is sent to the update loop to trigger a preview refresh
type tickMsg time.Time

// AgentItem adapts config.Agent to the bubbles/list.Item interface
type AgentItem struct {
	Name      string
	Command   string
	Active    bool
	SessionID string
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
	worktrees      *worktree.Manager
	previewContent string
	width          int
	height         int
	quitting       bool
	modal          Modal
	activeSessions map[string]string // sessionID -> agent name
}

// NewModel initializes the dashboard model
func NewModel(cfg *config.Config, tm SessionManager) Model {
	items := make([]list.Item, 0)
	activeSessions := make(map[string]string)

	if tm != nil {
		for _, s := range tm.ListSessions() {
			activeSessions[s.Name] = s.Name
		}
	}

	for _, agent := range cfg.Agents {
		full := sessionName(agent.Name)
		if _, ok := activeSessions[full]; ok {
			items = append(items, AgentItem{
				Name:      agent.Name,
				Command:   agent.Command,
				Active:    true,
				SessionID: full,
			})
		}
	}

	// Setup the list component
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Active Agents"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)

	return Model{
		list:           l,
		cfg:            cfg,
		tmux:           tm,
		worktrees:      worktree.NewManager(""),
		modal:          NewModal(cfg.Agents),
		activeSessions: activeSessions,
	}
}

// Init starts the tick loop for live previews
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Modal has priority
	if m.modal.IsActive() {
		switch msg.(type) {
		case instanceValidatedMsg, promptInstanceMsg:
			// allow main switch to process
		default:
			var modalCmd tea.Cmd
			m.modal, modalCmd = m.modal.Update(msg)
			if modalCmd != nil {
				cmds = append(cmds, modalCmd)
			}
			return m, tea.Batch(cmds...)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case promptInstanceMsg:
		// nothing; modal handles input

	case instanceValidatedMsg:
		return m.handleInstanceValidated(msg)

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tickMsg:
		return m.handleTick(cmds)

	case refreshListMsg:
		m = refreshList(m)

	case previewResultMsg:
		m.previewContent = string(msg)
	}

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd, func() tea.Msg { return refreshListMsg{} })

	return m, tea.Batch(cmds...)
}

// handleKeyMsg processes keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "n":
		m.modal.Show()
		return m, nil
	case "k":
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			agent := selectedItem.(AgentItem)
			return m, killSessionCmd(m.tmux, agent)
		}
	case "enter":
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			agent := selectedItem.(AgentItem)
			return m, attachToSessionCmd(m.tmux, agent)
		}
	}
	return m, nil
}

// handleInstanceValidated processes instance validation
func (m Model) handleInstanceValidated(msg instanceValidatedMsg) (tea.Model, tea.Cmd) {
	m.modal.Hide()
	return m, tea.Batch(
		createSessionWithWorktreeCmd(m.tmux, m.worktrees, m.cfg, msg.Agent, msg.Name),
		func() tea.Msg { return refreshListMsg{} },
	)
}

// handleWindowResize processes window resize events
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	listWidth := msg.Width / 4
	if listWidth < 24 {
		listWidth = 24
	}
	previewWidth := msg.Width - listWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}

	m.list.SetSize(listWidth, msg.Height-3)
	previewStyle = previewStyle.Width(previewWidth).Height(msg.Height - 3)
	return m, nil
}

// handleTick processes tick events for updating previews
func (m Model) handleTick(cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	if selectedItem := m.list.SelectedItem(); selectedItem != nil {
		agent := selectedItem.(AgentItem)
		cmds = append(cmds, fetchPreviewCmd(m.tmux, agent.SessionID))
	}
	cmds = append(cmds, tickCmd(), func() tea.Msg { return refreshListMsg{} })
	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	left := docStyle.Render(m.list.View())

	title := "PREVIEW: (None)"
	if i := m.list.SelectedItem(); i != nil {
		title = fmt.Sprintf("PREVIEW: %s", i.(AgentItem).Name)
	}

	rightContent := fmt.Sprintf("%s\n\n%s", title, m.previewContent)
	rightView := previewStyle.Render(rightContent)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)
	helpText := helpStyle.Render("[n] New Agent  [k] Kill Session  [↑/↓] Navigate  [Enter] Attach  [q] Quit")

	view := lipgloss.JoinHorizontal(lipgloss.Top, left, rightView)
	view = lipgloss.JoinVertical(lipgloss.Left, view, helpText)

	if m.modal.IsActive() {
		modalView := m.modal.View(m.width, m.height)
		view = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("0")),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

	return view
}

// --- Commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(previewInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// previewResultMsg carries the output of CapturePane
type previewResultMsg string
type refreshListMsg struct{}

func fetchPreviewCmd(tm SessionManager, session string) tea.Cmd {
	return func() tea.Msg {
		if session == "" || !tm.SessionExists(session) {
			return previewResultMsg("Agent is offline or idle.")
		}
		content, err := tm.CapturePane(session, 20)
		if err != nil {
			return previewResultMsg(fmt.Sprintf("Error fetching preview: %v", err))
		}
		return previewResultMsg(content)
	}
}

func attachToSessionCmd(tm SessionManager, agent AgentItem) tea.Cmd {
	return func() tea.Msg {
		session := agent.SessionID
		if session == "" {
			session = sessionName(agent.Name)
		}
		if !tm.SessionExists(session) {
			_, err := tm.CreateSession(session, agent.Command, "")
			if err != nil {
				return nil
			}
		}
		_ = tm.SwitchClient(session)
		return tea.Quit
	}
}

func createSessionWithWorktreeCmd(tm SessionManager, wt *worktree.Manager, cfg *config.Config, agent config.Agent, instance string) tea.Cmd {
	return func() tea.Msg {
		session := sessionNameWithInstance(agent.Name, instance)

		if tm.SessionExists(session) {
			return previewResultMsg(fmt.Sprintf("Session %s already exists", session))
		}

		cwd := ""
		if wt != nil && cfg != nil {
			path, err := wt.CreateWorktree(cfg.ProjectName, session)
			if err != nil {
				return previewResultMsg(fmt.Sprintf("Worktree error: %v", err))
			}
			cwd = path
		}

		if _, err := tm.CreateSession(session, agent.Command, cwd); err != nil {
			return previewResultMsg(fmt.Sprintf("Error creating session: %v", err))
		}

		return refreshListMsg{}
	}
}

func killSessionCmd(tm SessionManager, agent AgentItem) tea.Cmd {
	return func() tea.Msg {
		session := agent.SessionID
		if session == "" {
			session = sessionName(agent.Name)
		}
		_ = tm.KillSession(session)
		return refreshListMsg{}
	}
}

// refreshList recomputes active agent list from tmux sessions
func refreshList(m Model) Model {
	activeSessions := make(map[string]bool)
	if m.tmux != nil {
		for _, s := range m.tmux.ListSessions() {
			activeSessions[s.Name] = true
		}
	}

	items := make([]list.Item, 0)
	for _, agent := range m.cfg.Agents {
		full := sessionName(agent.Name)
		for sessionID := range activeSessions {
			if sessionID == full || sessionIDHasAgent(sessionID, agent.Name) {
				items = append(items, AgentItem{
					Name:      agent.Name,
					Command:   agent.Command,
					Active:    true,
					SessionID: sessionID,
				})
				break
			}
		}
	}

	m.list.SetItems(items)
	return m
}

func sessionIDHasAgent(sessionID, agentName string) bool {
	return sessionID == sessionName(agentName) ||
		len(sessionID) >= len(sessionName(agentName)) && sessionID[:len(sessionName(agentName))] == sessionName(agentName)
}
