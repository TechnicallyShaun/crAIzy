Epic: Agent Orchestration

# Hook-Based Identity Injection Prototype

Dependencies: craizy-init

## Overview

Prototype the hook-based identity injection system. When crAIzy spawns an agent, that agent should automatically receive its identity (role, hierarchy position, work assignment) via CLI hooks, without manual intervention.

This is a **prototype** to validate the approach before full implementation.

## User Story

As a user, when I create a new agent through crAIzy, that agent boots up and automatically knows its identity and role based on a variable passed during spawn.

**Acceptance Criteria:**
- Agent boots up successfully
- When agent/AI has loaded and is ready, a hook fires
- Hook receives an identifier variable from the spawning process
- Agent's context is injected with role-appropriate instructions
- This works for each CLI defined in AGENTS.yml (Claude, Gemini, Copilot)
- Hook configuration is NOT global - installed via `craizy init`
- Identity can be derived from the passed variable (role, parent, work item)

---

## Technical Exploration

### Variable Passing Options

We need to pass identity from orchestrator → hook. Options to prototype:

#### Option 1: Environment Variable

```go
// In AgentService.Create()
cmd := exec.Command(agentCommand)
cmd.Env = append(os.Environ(),
    "CRAIZY_AGENT_ID=director/lead.a/worker.2:story.5",
)
```

Hook reads:
```bash
#!/bin/bash
ID="${CRAIZY_AGENT_ID:-unknown}"
# Parse and inject
```

**Test:** Does the env var propagate through tmux to the CLI's hook?

#### Option 2: tmux Environment

```go
// Set env on the tmux session itself
tmux.SetEnv(sessionID, "CRAIZY_AGENT_ID", agentID)
```

Hook reads from tmux:
```bash
#!/bin/bash
ID=$(tmux show-environment -t "$TMUX_PANE" CRAIZY_AGENT_ID 2>/dev/null | cut -d= -f2)
```

**Test:** Can hooks access tmux session environment?

#### Option 3: File-Based

```go
// Write identity file before spawn
identityFile := filepath.Join(".craizy", "agents", sessionID, "identity")
os.WriteFile(identityFile, []byte(agentID), 0600)
```

Hook reads:
```bash
#!/bin/bash
SESSION=$(tmux display-message -p '#{session_name}')
ID=$(cat ".craizy/agents/${SESSION}/identity" 2>/dev/null)
```

**Test:** Can hook reliably determine its session and find the file?

#### Option 4: Session Name as ID

```go
// Use the agent ID as the tmux session name
sessionID := "director/lead.a/worker.2:story.5"  // or sanitized version
```

Hook extracts from session name:
```bash
#!/bin/bash
ID=$(tmux display-message -p '#{session_name}')
```

**Test:** What characters are valid in tmux session names?

---

## Stories

### Story 1: Hook Infrastructure in craizy init

As a developer, `craizy init` installs hook configurations for each supported CLI.

**Acceptance Criteria:**
- `craizy init` creates `.craizy/hooks/inject-identity.sh`
- `craizy init` creates/updates `.claude/settings.json` with SessionStart hook
- `craizy init` creates/updates `.gemini/settings.json` with SessionStart hook
- `craizy init` creates/updates `.github/hooks/hooks.json` with sessionStart hook
- Existing user hook configurations are preserved (merged, not overwritten)
- Running `craizy init` again is idempotent

**Technical Notes:**

```
.craizy/
├── hooks/
│   └── inject-identity.sh    # Shared hook script
├── roles/
│   ├── director.md
│   ├── lead.md
│   └── worker.md
└── AGENTS.yml

.claude/
└── settings.json             # Claude hook config

.gemini/
└── settings.json             # Gemini hook config

.github/
└── hooks/
    └── hooks.json            # Copilot hook config
```

Hook config examples:

```json
// .claude/settings.json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": ".craizy/hooks/inject-identity.sh"
    }]
  }
}
```

```json
// .gemini/settings.json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": ".craizy/hooks/inject-identity.sh"
    }]
  }
}
```

```json
// .github/hooks/hooks.json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{
      "type": "command",
      "bash": ".craizy/hooks/inject-identity.sh",
      "timeoutSec": 10
    }]
  }
}
```

---

### Story 2: Identity Variable Propagation

As a developer, when AgentService spawns an agent, the identity variable is accessible to the hook.

**Acceptance Criteria:**
- `CRAIZY_AGENT_ID` env var is set before spawning
- Hook can read the variable and output its value
- Works for Claude, Gemini, and Copilot CLIs
- If variable is missing, hook outputs a sensible default/error

**Technical Notes:**

Modify `internal/agent/service.go`:

```go
func (s *AgentService) Create(project, agentType, name, command, workDir string) (*Agent, error) {
    sessionID := buildSessionID(project, agentType, name)

    // For prototype: simple ID format
    // Future: full hierarchy path + work assignment
    agentID := fmt.Sprintf("%s/%s", agentType, name)

    // Pass identity via environment
    env := append(os.Environ(), "CRAIZY_AGENT_ID="+agentID)

    if err := s.tmux.CreateSessionWithEnv(sessionID, command, workDir, env); err != nil {
        return nil, err
    }
    // ...
}
```

Modify `internal/tmux/client.go`:

```go
func (c *Client) CreateSessionWithEnv(id, command, workDir string, env []string) error {
    // Option A: Pass env to the shell that runs the command
    // Option B: Use tmux set-environment
    // Prototype both and see which works with hooks
}
```

---

### Story 3: Hook Script Implementation

As a developer, the hook script parses the identity variable and outputs role-appropriate instructions.

**Acceptance Criteria:**
- Script reads `CRAIZY_AGENT_ID` environment variable
- Script determines role from the ID (director/lead/worker)
- Script outputs the appropriate role markdown file
- Script appends instance-specific context (ID, parent, work item)
- Output is valid for hook injection (stdout only, no debug to stdout)

**Technical Notes:**

`.craizy/hooks/inject-identity.sh`:
```bash
#!/bin/bash

# Get identity from environment
ID="${CRAIZY_AGENT_ID:-}"

if [ -z "$ID" ]; then
    echo "# Unknown Agent"
    echo "No identity provided. Running in standalone mode."
    exit 0
fi

# Parse ID format: hierarchy:work_item
# Example: director/lead.a/worker.2:story.5
HIERARCHY="${ID%:*}"
WORK_ITEM="${ID#*:}"

# Handle case where there's no work item (no colon)
if [ "$HIERARCHY" = "$ID" ]; then
    WORK_ITEM=""
fi

# Get role from last segment of hierarchy
AGENT_SEG=$(basename "$HIERARCHY")
ROLE=$(echo "$AGENT_SEG" | sed 's/[.0-9a-z]*$//')

# Get parent (everything except last segment)
PARENT=$(dirname "$HIERARCHY")
if [ "$PARENT" = "." ]; then
    PARENT=""
fi

# Determine role file
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
ROLE_FILE="${PROJECT_ROOT}/.craizy/roles/${ROLE}.md"

# Output role instructions
if [ -f "$ROLE_FILE" ]; then
    cat "$ROLE_FILE"
else
    echo "# Role: ${ROLE}"
    echo "No role definition found at ${ROLE_FILE}"
fi

# Append instance context
echo ""
echo "## Instance Context"
echo "- Agent ID: ${ID}"
echo "- Role: ${ROLE}"
echo "- Reports To: ${PARENT:-none (top level)}"
if [ -n "$WORK_ITEM" ]; then
    echo "- Assignment: ${WORK_ITEM}"
    echo ""
    echo "## Getting Started"
    echo "Retrieve your assignment: \`craizy work get ${WORK_ITEM}\`"
fi
```

---

### Story 4: Role Definition Files

As a developer, role definition files provide role-specific instructions for agents.

**Acceptance Criteria:**
- `.craizy/roles/director.md` exists with Director instructions
- `.craizy/roles/lead.md` exists with Lead instructions
- `.craizy/roles/worker.md` exists with Worker instructions
- Files are created by `craizy init`
- Files can be customized by user (not overwritten on re-init)

**Technical Notes:**

Embed default role files in the binary, copy during init if not exists.

Example `.craizy/roles/worker.md`:
```markdown
# Role: Worker

You are a Worker agent in the crAIzy orchestration system.

## Your Responsibilities
- Execute the assigned work item directly (write code, run commands)
- Ask questions when blocked (escalate to your Lead)
- Report completion when the work item is done

## Your Constraints
- Do NOT decompose work into smaller pieces (that's Lead's job)
- Do NOT make architectural decisions (escalate to Lead)
- Do NOT spawn other agents
- Do NOT work on things outside your assignment

## Communication Protocol
- **Questions/Blockers:** Use `craizy escalate "your question here"`
- **Completion:** Use `craizy done "summary of what was done"`
- **Status:** Use `craizy status "what you're currently doing"`

## Starting Work
1. Read the assignment details provided below
2. Understand what "done" looks like (acceptance criteria)
3. Execute the work
4. Report completion
```

---

### Story 5: Validation Test Suite

As a developer, I can run a test that validates the full hook → identity → injection flow.

**Acceptance Criteria:**
- Test spawns a Claude agent with a known identity
- Test verifies the hook fired (check log or output)
- Test verifies the agent received the identity (agent can echo it back)
- Test covers Claude, Gemini, Copilot (or documents which work)
- Test can be run manually: `go test ./... -run TestHookInjection`

**Technical Notes:**

This may need to be a manual/integration test rather than unit test, since it requires actual CLI binaries.

```go
func TestHookInjection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup: ensure craizy init has been run
    // Spawn agent with known ID
    // Send message asking agent to report its identity
    // Verify response contains expected values
}
```

---

## Prototype Scope

**In Scope:**
- Hook installation via `craizy init`
- Environment variable propagation
- Hook script that parses ID and outputs role instructions
- Basic role definition files
- Manual testing with each CLI

**Out of Scope:**
- Full hierarchy path (prototype uses simple `type/name` format)
- Work item assignment (prototype uses placeholder)
- Database integration
- Automatic role derivation from hierarchy depth
- Error handling/retry for hook failures
- tmux send-keys fallback for CLIs without hook support

---

## Verification

1. Run `craizy init` in a test directory
2. Verify hook configs created for Claude, Gemini, Copilot
3. Verify hook script is executable
4. Verify role files exist
5. Manually spawn each agent type and verify identity injection:
   ```bash
   # Set the env var manually to test
   CRAIZY_AGENT_ID="worker/test-worker:story.1" claude
   # Agent should show identity in its initial context
   ```
6. Spawn agent through crAIzy UI and verify hook fires

---

## Open Questions for Prototype

1. **Environment propagation through tmux** - Does `cmd.Env` propagate through `tmux new-session` to the subprocess? Need to test.

2. **Hook stdout injection** - Does each CLI actually inject hook stdout into context, or just run the hook silently? Need to test.

3. **Working directory for hooks** - What is the CWD when the hook runs? Project root? Home? Need to verify for each CLI.

4. **Hook execution timing** - Does `SessionStart` fire before or after the CLI shows its prompt? Need to verify.

5. **Multiple hooks** - If user has existing hooks, does our hook merge correctly? Need to test merge logic.

---

## Related Documents

- `visions/hook-identity-injection.md` - Detailed exploration of approaches
- `visions/hierarchical-agents.md` - Agent hierarchy model
- `visions/process-model.md` - Work decomposition model
- `features/complete/craizy-init.md` - Current init implementation
