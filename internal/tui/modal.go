package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Modal is a generic wrapper for modal content.
// It handles centering and overlaying content on top of the application.
type Modal struct {
	content tea.Model
	isOpen  bool
	width   int
	height  int
}

func NewModal() Modal {
	return Modal{}
}

func (m *Modal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Modal) Open(content tea.Model) {
	m.content = content
	m.isOpen = true
}

func (m *Modal) Close() {
	m.content = nil
	m.isOpen = false
}

func (m *Modal) IsOpen() bool {
	return m.isOpen
}

// Init initializes the modal content if it exists
func (m *Modal) Init() tea.Cmd {
	if m.content != nil {
		return m.content.Init()
	}
	return nil
}

// Update handles messages for the modal content.
// Returns a command and whether the message was handled/consumed.
func (m *Modal) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !m.isOpen || m.content == nil {
		return nil, false
	}

	var cmd tea.Cmd
	m.content, cmd = m.content.Update(msg)
	return cmd, true
}

// View returns the rendered modal content centered in the screen.
// Note: This creates a full-screen view with the modal centered.
// To achieve a true "overlay" effect where the background is visible but dimmed,
// more complex string manipulation would be required, or simply accepting
// that the background is hidden/blanked out is the standard TUI approach.
func (m Modal) View() string {
	if !m.isOpen || m.content == nil {
		return ""
	}

	modalView := m.content.View()
	
	// Create a centered box for the modal
	return lipgloss.Place(
		m.width, 
		m.height, 
		lipgloss.Center, 
		lipgloss.Center, 
		modalView,
		lipgloss.WithWhitespaceChars(" "), // Clears the background
		// lipgloss.WithWhitespaceForeground(lipgloss.Color("240")), // Optional: dim color if we used a pattern
	)
}
