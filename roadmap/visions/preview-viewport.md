# Preview Pane Viewport

Notes on using bubbles Viewport component for the preview pane to enable scrolling and better overflow handling.

## Current Implementation

Preview pane in `internal/tui/content_area.go` uses simple line truncation:
- Lines exceeding pane width are cut off at the edge
- Only the last N lines that fit are shown
- No scrolling capability

This works but loses information when output is too wide or when users want to see history.

## Bubbles Viewport

The `github.com/charmbracelet/bubbles/viewport` package provides:
- Scrollable content area with keyboard navigation
- Automatic content wrapping (optional)
- Built-in scroll position tracking
- Mouse wheel support (when enabled)

### Basic Usage

```go
import "github.com/charmbracelet/bubbles/viewport"

type ContentAreaModel struct {
    viewport viewport.Model
}

func NewContentArea(width, height int) ContentAreaModel {
    vp := viewport.New(width, height)
    vp.SetContent("")
    return ContentAreaModel{viewport: vp}
}

func (m *ContentAreaModel) SetPreview(content string) {
    m.viewport.SetContent(content)
    // Auto-scroll to bottom for live preview
    m.viewport.GotoBottom()
}

func (m ContentAreaModel) Update(msg tea.Msg) (ContentAreaModel, tea.Cmd) {
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m ContentAreaModel) View() string {
    return m.viewport.View()
}
```

### Key Bindings

Default viewport navigation:
- `j/k` or `up/down` - Scroll line by line
- `d/u` - Scroll half page
- `g/G` - Go to top/bottom
- Mouse wheel - Scroll (if mouse enabled)

These would need coordination with dashboard to avoid conflicts with agent list navigation.

## Implementation Considerations

### Pros
- Proper overflow handling without content loss
- Users can scroll back to see earlier output
- Well-tested, maintained component
- Mouse support possible

### Cons
- Adds complexity to key handling (viewport vs agent list navigation)
- Need to manage scroll position (auto-follow for live, manual when user scrolls)
- May need visual scroll indicators
- Slight performance overhead

### Focus Management

Key challenge: when should arrow keys control the viewport vs the agent list?

Options:
1. **Mode-based**: Press `Tab` to switch focus between list and preview
2. **Modifier-based**: Use `Shift+arrows` for viewport, plain arrows for list
3. **Context-based**: Only enable viewport scrolling when preview is "focused" (e.g., after clicking or pressing a key)

### Auto-Follow Behavior

For live preview, typically want:
- Auto-scroll to bottom on new content
- Stop auto-scroll when user manually scrolls up
- Resume auto-scroll when user goes to bottom or presses `G`

```go
func (m *ContentAreaModel) SetPreview(content string) {
    wasAtBottom := m.viewport.AtBottom()
    m.viewport.SetContent(content)
    if wasAtBottom || m.autoFollow {
        m.viewport.GotoBottom()
    }
}
```

### Visual Indicators

Could add:
- Scroll position indicator (e.g., percentage or scrollbar)
- "More content above/below" indicators
- Visual hint when not at bottom (live content arriving)

## Related Work

- Preview pane feature: `roadmap/features/complete/preview-pane.md`
- Content area implementation: `internal/tui/content_area.go`
- Bubbles viewport: https://github.com/charmbracelet/bubbles/tree/master/viewport

## Decision

Currently using simple truncation (Option A) for simplicity. Viewport (Option C) is a future enhancement if users need to scroll through preview history.
