# Hierarchical Agent Organization

## Vision

Transform the flat agent list in the side menu into a hierarchical tree structure representing organizational relationships between agents.

## Structure

Agents report to other agents in a chain of command:

```
Director
├── Lead 1
│   ├── Worker 1
│   └── Worker 2
├── Lead 2
│   ├── Worker 3
│   └── Worker 4
└── Lead 3
    └── Worker 5
```

## Data Model

Add hierarchy awareness to agents:

```go
type Agent struct {
    ID        string
    Name      string
    Type      string
    ReportsTo *string  // nullable - Director reports to no one
    Order     int      // sibling ordering within same parent
}
```

Example data:
| Agent    | Reports To |
|----------|------------|
| Director | (none)     |
| Lead 1   | Director   |
| Lead 2   | Director   |
| Worker 1 | Lead 1     |
| Worker 2 | Lead 1     |
| Worker 3 | Lead 2     |

## UI Behavior

- Tree view in side menu with expand/collapse
- Visual connectors (├──, └──, │)
- Indentation indicates depth
- Keyboard navigation respects hierarchy

## Technical Options

1. **Lipgloss Tree** - Already in dependencies (v1.1.0), rendering only
2. **mariusor/bubbles-tree** - Community package, full interactivity
3. **Official Bubbles Tree** - Coming soon (PR #639)
4. **Custom List Delegate** - Extend current `bubbles/list` with hierarchy-aware rendering

## Open Questions

- Can an agent have multiple parents? (matrix org)
- Maximum depth limit?
- How does hierarchy affect agent spawning/messaging?
- Visual distinction for different roles (Director vs Lead vs Worker)?

## Related

- Current side menu: `internal/tui/side_menu.go`
- Uses `bubbles/list` component
