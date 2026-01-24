Epic: MVP

# Theme System & Colour Scheme

Dependencies: None

## Description

A centralized theme system for consistent styling across the TUI and tmux components. Currently, colours are hardcoded inline across 6+ files using raw Lipgloss colour codes (e.g., `lipgloss.Color("86")`). This makes the UI inconsistent and difficult to maintain.

The solution is a **centralized theme package** - similar to CSS custom properties - where all colours, borders, and styles are defined once and imported by components. This also enables future theming (light/dark mode, user customization).

Additionally, the dashboard layout needs adjustment - previous terminal output bleeds through at the top, requiring a 2-line height increase.

## Colour Scheme Recommendation

For a terminal-based AI orchestration tool, the colour scheme should be:
- **Professional and calm** - users will stare at this for extended sessions
- **High contrast** - readable on various terminal backgrounds
- **Semantically meaningful** - colours communicate status at a glance
- **Limited palette** - 8-10 colours max to avoid visual noise

### Recommended: Nord-inspired palette

Nord is muted, professional, and works well in tmux. Adapted for 256-colour terminals:

| Role | Colour | ANSI 256 | Hex | Usage |
|------|--------|----------|-----|-------|
| **Background** | Polar Night | 235 | #2E3440 | Base (or terminal default) |
| **Foreground** | Snow Storm | 255 | #ECEFF4 | Primary text |
| **Muted** | Grey | 243 | #4C566A | Secondary text, borders |
| **Frost 1** | Cyan | 109 | #8FBCBB | Accent, highlights |
| **Frost 2** | Blue | 110 | #88C0D0 | Active elements, selection |
| **Frost 3** | Light Blue | 111 | #81A1C1 | Links, interactive |
| **Frost 4** | Deep Blue | 68 | #5E81AC | Borders, dividers |
| **Aurora Green** | Green | 108 | #A3BE8C | Success, running agents |
| **Aurora Yellow** | Yellow | 222 | #EBCB8B | Warning, pending |
| **Aurora Red** | Red | 174 | #BF616A | Error, stopped agents |
| **Aurora Purple** | Purple | 139 | #B48EAD | Special, modal accents |

### Alternative: Catppuccin Mocha

If a warmer feel is preferred:

| Role | ANSI 256 | Hex |
|------|----------|-----|
| Base | 236 | #1E1E2E |
| Text | 255 | #CDD6F4 |
| Subtext | 243 | #A6ADC8 |
| Blue | 111 | #89B4FA |
| Green | 114 | #A6E3A1 |
| Yellow | 221 | #F9E2AF |
| Red | 210 | #F38BA8 |
| Mauve | 183 | #CBA6F7 |

## Stories

### Centralize Theme Definitions

As a developer, I can import colours from a single theme package instead of hardcoding values.

#### Technical / Architecture

- Create `internal/tui/theme/theme.go`:
  ```go
  package theme

  import "github.com/charmbracelet/lipgloss"

  // Colour palette - Nord inspired
  var (
      // Base colours
      ColorBackground = lipgloss.Color("235")
      ColorForeground = lipgloss.Color("255")
      ColorMuted      = lipgloss.Color("243")

      // Accent colours (Frost)
      ColorPrimary    = lipgloss.Color("110")  // Main accent
      ColorSecondary  = lipgloss.Color("111")  // Secondary accent
      ColorBorder     = lipgloss.Color("68")   // Borders, dividers

      // Semantic colours (Aurora)
      ColorSuccess    = lipgloss.Color("108")  // Green - running, success
      ColorWarning    = lipgloss.Color("222")  // Yellow - pending, warning
      ColorError      = lipgloss.Color("174")  // Red - stopped, error
      ColorSpecial    = lipgloss.Color("139")  // Purple - modals, special
  )

  // Reusable styles
  var (
      // Text styles
      TextNormal = lipgloss.NewStyle().
          Foreground(ColorForeground)

      TextMuted = lipgloss.NewStyle().
          Foreground(ColorMuted)

      TextSuccess = lipgloss.NewStyle().
          Foreground(ColorSuccess)

      TextError = lipgloss.NewStyle().
          Foreground(ColorError)

      // Border styles
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
  var (
      // Side menu
      SideMenuTitle = lipgloss.NewStyle().
          Foreground(ColorPrimary).
          Bold(true)

      SideMenuEmpty = lipgloss.NewStyle().
          Foreground(ColorMuted).
          Italic(true)

      // Agent status indicators
      AgentRunning = lipgloss.NewStyle().
          Foreground(ColorSuccess)

      AgentStopped = lipgloss.NewStyle().
          Foreground(ColorError)

      AgentPending = lipgloss.NewStyle().
          Foreground(ColorWarning)

      // Content area
      ContentTitle = lipgloss.NewStyle().
          Foreground(ColorPrimary).
          Bold(true)

      ContentSubtitle = lipgloss.NewStyle().
          Foreground(ColorMuted)

      // Modal
      ModalTitle = lipgloss.NewStyle().
          Foreground(ColorSpecial).
          Bold(true)

      ModalBorder = BorderRounded.Copy().
          BorderForeground(ColorSpecial)

      // Quick commands bar
      QuickCommandKey = lipgloss.NewStyle().
          Foreground(ColorPrimary).
          Bold(true)

      QuickCommandDesc = lipgloss.NewStyle().
          Foreground(ColorMuted)
  )
  ```

- Update components to import theme:
  ```go
  // Before (content_area.go)
  style := lipgloss.NewStyle().
      Border(lipgloss.NormalBorder()).
      BorderForeground(lipgloss.Color("86"))

  // After
  import "craizy/internal/tui/theme"

  style := theme.BorderNormal.Copy().
      Width(m.width).
      Height(m.height)
  ```

### Update Existing Components

As a developer, I refactor existing TUI components to use the centralized theme.

#### Technical / Architecture

Files to update:
- `internal/tui/content_area.go` - Replace `"86"`, `"250"`, `"245"`
- `internal/tui/side_menu.go` - Replace `"240"`
- `internal/tui/quick_commands.go` - Replace `"245"`
- `internal/tui/name_input.go` - Replace `"205"`, `"63"`
- `internal/tui/modal.go` - Apply modal styles
- `internal/tui/agent_selector.go` - Apply selection styles

Example migration for `side_menu.go`:
```go
// Before
emptyStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("240")).
    Italic(true)

// After
emptyStyle := theme.SideMenuEmpty
```

### Fix Dashboard Height

As a user, I should not see previous terminal output bleeding through at the top of the dashboard.

#### Technical / Architecture

In `internal/tui/dashboard.go`, the `bottomHeight` constant needs adjustment:

```go
// Current (line ~150)
bottomHeight := 5

// Fix: Add 2 lines to account for terminal scroll buffer bleed
bottomHeight := 7
```

Alternatively, the issue may be in how the main area height is calculated. A more robust fix:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height

    // Clear screen on resize to prevent bleed-through
    // (handled by Bubble Tea, but ensure full repaint)

    bottomHeight := 5
    mainHeight := m.height - bottomHeight - 2  // Extra padding for safety

    if mainHeight < 1 {
        mainHeight = 1
    }
```

Or use Lipgloss's `Place` to ensure content fills the viewport:
```go
func (m DashboardModel) View() string {
    // ... build view ...

    // Ensure view fills entire terminal
    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Left, lipgloss.Top,
        baseView,
    )
}
```

### Tmux Status Bar Colours

As a user, the tmux status bar styling should match the TUI colour scheme.

#### Technical / Architecture

Update the tmux configuration in agent spawn logic to use consistent colours:

```bash
# Nord-inspired tmux status bar
tmux set-option -t "$session" status-style "bg=#3B4252,fg=#ECEFF4"
tmux set-option -t "$session" status-left-style "bg=#5E81AC,fg=#ECEFF4"
tmux set-option -t "$session" status-right-style "bg=#4C566A,fg=#D8DEE9"
```

Or using 256-colour codes for broader compatibility:
```bash
tmux set-option -t "$session" status-style "bg=colour237,fg=colour255"
tmux set-option -t "$session" status-left-style "bg=colour68,fg=colour255"
```

## Package Structure

```
internal/
└── tui/
    ├── theme/
    │   └── theme.go          # Centralized colours and styles
    ├── dashboard.go          # Uses theme imports
    ├── side_menu.go          # Uses theme imports
    ├── content_area.go       # Uses theme imports
    ├── quick_commands.go     # Uses theme imports
    ├── modal.go              # Uses theme imports
    ├── name_input.go         # Uses theme imports
    └── agent_selector.go     # Uses theme imports
```

## Open Questions

1. **Which palette?** Nord (professional, muted) or Catppuccin (warmer, popular)?
2. **User theming?** Should users be able to customize colours via config file? (Suggest: out of scope for MVP)
3. **Terminal background assumption?** Should we assume dark terminal, or detect/adapt?

## Out of Scope

- Light mode / theme switching
- User-configurable colour schemes via config file
- True colour (24-bit) support - sticking with 256-colour for compatibility
- Syntax highlighting for code in preview pane (separate feature)
