package theme

import "github.com/charmbracelet/lipgloss"

// Nord-inspired colour palette for 256-colour terminals.
// See: https://www.nordtheme.com/
var (
	// Base colours (Polar Night / Snow Storm)
	ColorBackground = lipgloss.Color("235") // #2E3440
	ColorForeground = lipgloss.Color("255") // #ECEFF4
	ColorMuted      = lipgloss.Color("243") // #4C566A

	// Accent colours (Frost)
	ColorPrimary   = lipgloss.Color("110") // #88C0D0 - Main accent
	ColorSecondary = lipgloss.Color("111") // #81A1C1 - Secondary accent
	ColorBorder    = lipgloss.Color("68")  // #5E81AC - Borders, dividers

	// Semantic colours (Aurora)
	ColorSuccess = lipgloss.Color("108") // #A3BE8C - Green: running, success
	ColorWarning = lipgloss.Color("222") // #EBCB8B - Yellow: pending, warning
	ColorError   = lipgloss.Color("174") // #BF616A - Red: stopped, error
	ColorSpecial = lipgloss.Color("139") // #B48EAD - Purple: modals, special
)

// Reusable text styles
var (
	TextNormal = lipgloss.NewStyle().
			Foreground(ColorForeground)

	TextMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	TextSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	TextWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	TextError = lipgloss.NewStyle().
			Foreground(ColorError)
)

// Reusable border styles
var (
	BorderNormal = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	BorderFocused = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorPrimary)

	BorderRounded = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSpecial)
)

// Component-specific styles

// Side menu styles
var (
	SideMenuTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SideMenuEmpty = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)

// Agent status indicator styles
var (
	AgentRunning = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	AgentStopped = lipgloss.NewStyle().
			Foreground(ColorError)

	AgentPending = lipgloss.NewStyle().
			Foreground(ColorWarning)
)

// Content area styles
var (
	ContentTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	ContentSubtitle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ContentLogo = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	ContentVersion = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ContentTagline = lipgloss.NewStyle().
			Foreground(ColorForeground)
)

// Modal styles
var (
	ModalTitle = lipgloss.NewStyle().
			Foreground(ColorSpecial).
			Bold(true)

	ModalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSpecial)
)

// Quick commands bar styles
var (
	QuickCommandKey = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	QuickCommandDesc = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// TmuxStatusBar contains color values for tmux status bar configuration.
// Uses hex values for broader tmux compatibility.
var TmuxStatusBar = struct {
	Background      string
	Foreground      string
	BrandColor      string
	AccentColor     string
	MutedColor      string
	SeparatorColor  string
}{
	Background:     "#3B4252", // Nord1 - slightly lighter than base
	Foreground:     "#ECEFF4", // Nord6 - Snow Storm
	BrandColor:     "#88C0D0", // Nord8 - Frost (primary)
	AccentColor:    "#81A1C1", // Nord9 - Frost (secondary)
	MutedColor:     "#4C566A", // Nord3 - muted gray
	SeparatorColor: "#4C566A", // Nord3
}
