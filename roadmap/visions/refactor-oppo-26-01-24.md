# Refactoring Opportunities - Code Quality Review

**Date:** 2026-01-24
**Status:** Identified
**Scope:** Full codebase analysis for duplication, SOLID violations, and refactoring opportunities

---

## P0 - Critical Issues

### 1. God Object: Dashboard Model (Single Responsibility Violation)

**Location:** `internal/tui/dashboard.go:15-34, 78-211`

**Problem:** The main `Model` struct violates Single Responsibility Principle by managing multiple concerns:
- 4 child UI models (sideMenu, contentArea, quickCommands, modal)
- Agent service orchestration
- Preview polling logic
- Event handling
- Size management
- Key binding logic

The `Update` method is 133 lines and handles 30+ different message types.

**Impact:** Hard to test, difficult to maintain, poor separation of concerns. Any change to one feature affects the entire model.

**Recommendation:** Split into:
- `LayoutManager` - handle sizing and dimensions
- `KeyBindingHandler` - handle key events
- `PreviewPoller` - manage preview polling
- Keep `Model` as thin orchestrator

---

### 2. Concrete Dependency Coupling (Dependency Inversion Violation)

**Location:** `internal/tui/dashboard.go:22, 50-73`

**Problem:** The dashboard directly accesses multiple methods on `AgentService`:
```go
m.agentService.List()
m.agentService.CaptureOutput()
m.agentService.Create()
m.agentService.Attach()
m.agentService.Kill()
```

The TUI layer depends on concrete implementations instead of abstractions.

**Impact:** Cannot unit test UI logic independently; tight coupling makes future refactoring difficult.

**Recommendation:** Create a `UIService` interface:
```go
type UIService interface {
    ListAgents() []*domain.Agent
    CaptureAgentOutput(sessionID string, lines int) (string, error)
    CreateAgent(agentType, name, command string) (*domain.Agent, error)
    AttachToAgent(sessionID string) tea.Cmd
    KillAgent(sessionID string) error
}
```

---

### 3. Silent Error Swallowing

**Locations:**
- `internal/tui/dashboard.go:73, 110-113`
- `internal/domain/service.go:43, 122, 138`
- `internal/infra/adapters.go:19, 26-27`

**Problem:** Errors are silently ignored throughout the codebase (12+ instances of `_ =` error suppression):
```go
_, err := m.agentService.Create(msg.Agent.Name, msg.CustomName, msg.Agent.Command)
if err != nil {
    // TODO: Show error to user  <-- Error silently ignored!
    return m, nil
}
```

**Impact:** Users have no visibility into failures. Critical operations fail silently, making debugging impossible.

**Recommendation:**
1. Add structured logging (slog or zap)
2. Propagate errors to UI for user display
3. Implement error recovery strategies

---

### 4. Unsafe Type Assertions

**Location:** `internal/infra/adapters.go:11, 25`

**Problem:** Event handlers use unsafe type assertions:
```go
dispatcher.Subscribe("agent.created", func(e domain.Event) {
    event := e.(domain.AgentCreated)  // Panics if wrong type
})
```

**Impact:** One mistake in event publishing causes application crash.

**Recommendation:** Add safe type assertion:
```go
event, ok := e.(domain.AgentCreated)
if !ok {
    log.Printf("unexpected event type: %T", e)
    return
}
```

---

### 5. SQL Query Duplication (DRY Violation)

**Location:** `internal/infra/store/sqlite_store.go:64, 100`

**Problem:** Identical SELECT statement duplicated in `List()` and `Get()`. Scan logic also duplicated at lines 78-88 and 102-112.

**Impact:** Schema changes require updates in 3+ places. Bug fixes must be repeated.

**Recommendation:**
1. Extract query constant: `const agentSelectQuery = "SELECT id, project, ..."`
2. Extract scan method: `scanAgent(scanner) (*Agent, error)`

---

## P1 - High Priority Issues

### 6. Missing `Close()` in Interface (Open/Closed Violation)

**Location:** `internal/domain/interfaces.go`

**Problem:** `IAgentStore` interface lacks `Close()` method but `SQLiteAgentStore` implements it. Cannot use store through interface for resource cleanup.

**Recommendation:** Add to interface:
```go
type IAgentStore interface {
    // ... existing methods ...
    Close() error
}
```

---

### 7. Regex Compiled on Every Call

**Location:** `internal/domain/agent.go:41-60`

**Problem:** `SanitizeName()` compiles regex on every call plus makes 6 string passes:
```go
func SanitizeName(name string) string {
    reg := regexp.MustCompile(`[^a-z0-9-]`)  // Compiled every call!
    // ... 6 passes over string ...
}
```

**Impact:** Called frequently on agent creation, creates garbage and CPU waste.

**Recommendation:**
```go
var sanitizeRegex = regexp.MustCompile(`[^a-z0-9-]`)  // Compile once

func SanitizeName(name string) string {
    name = strings.ToLower(name)
    name = strings.NewReplacer(".", "", ":", "", " ", "-").Replace(name)
    name = sanitizeRegex.ReplaceAllString(name, "")
    name = strings.Trim(strings.ReplaceAll(name, "--", "-"), "-")
    return name
}
```

---

### 8. Inconsistent Error Return Types

**Location:** `internal/infra/store/sqlite_store.go:74-90`

**Problem:** `List()` returns `nil` on error instead of empty slice. Silently skips malformed rows.

**Impact:** Caller can't distinguish "no agents" from "query failed".

**Recommendation:** Return empty slice on error, log failures, consider returning error.

---

### 9. Repeated Message Type Casting Pattern

**Locations:**
- `internal/tui/dashboard.go:81, 146`
- `internal/tui/agent_selector.go:49`
- `internal/tui/side_menu.go:57`
- `internal/tui/name_input.go:39`

**Problem:** Every `Update()` method has identical `switch msg := msg.(type)` pattern.

**Recommendation:** Create `MessageRouter` interface/helper to centralize message dispatch.

---

## P2 - Medium Priority Issues

### 10. Magic Numbers Scattered

**Locations:**
- `internal/tui/dashboard.go:153-160` - dimension calculations
- `internal/tui/quick_commands.go:40-41` - color code "202"
- `internal/tui/side_menu.go:104` - color code "63"
- `internal/tui/content_area.go:61` - color code "86"

**Problem:** Hardcoded values like `bottomHeight := 5`, color codes scattered throughout.

**Recommendation:** Create theme constants:
```go
const (
    ThemeColorBorder    = "63"
    ThemeColorAccent    = "86"
    ThemeSidebarRatio   = 0.25
    ThemeBottomHeight   = 5
)
```

---

### 11. Preview Errors Ignored

**Location:** `internal/tui/dashboard.go:73`

**Problem:** `CaptureOutput` error discarded - user sees empty/stale preview without knowing why.

**Recommendation:** Surface preview capture failures to user.

---

### 12. Reconciliation Mixed Concerns

**Location:** `internal/domain/service.go:108-144`

**Problem:** `Reconcile()` does orphan detection AND cleanup AND status updates in one method. Also has silent error suppression.

**Recommendation:** Split into separate methods with proper error handling.

---

### 13. Over-engineered Modal

**Location:** `internal/tui/modal.go:1-82`

**Problem:** Generic `tea.Model` content but only ever used with 2 specific types (AgentSelector, NameInput).

**Recommendation:** Use specific modal types or add typed methods.

---

## P3 - Low Priority Issues

### 14. Thin Wrapper Methods

**Location:** `internal/domain/service.go:98-106`

**Problem:** `Exists()` and `CaptureOutput()` just delegate to dependencies without adding value.

**Recommendation:** Consider removing or documenting purpose.

---

### 15. Inconsistent Lock Patterns

**Locations:**
- `internal/infra/memory_store.go`
- `internal/infra/event_dispatcher.go`

**Problem:** Mix of `RLock`/`Lock` - correct but error-prone without helper abstractions.

**Recommendation:** Consider lock helper methods for consistency.

---

## Summary

| Category | Count |
|----------|-------|
| SOLID Violations | 5 |
| Code Duplication | 4 |
| Error Handling | 6 |
| Architecture/Coupling | 4 |
| Performance | 2 |
| API Design | 4 |
| **Total** | **25+** |

---

## Recommended Implementation Order

1. Add error logging/propagation (quick win, high impact)
2. Fix unsafe type assertions in `adapters.go` (prevents crashes)
3. Extract SQL query constants and scan helper
4. Add `Close()` to `IAgentStore` interface
5. Create `UIService` interface for dashboard
6. Split Dashboard model into focused components
7. Compile regex once in `SanitizeName()`
8. Create theme constants package
