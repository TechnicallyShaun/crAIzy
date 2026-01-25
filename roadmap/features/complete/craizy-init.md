# Feature: craizy init Command

## Overview

Implement a `craizy init` subcommand that initializes/upgrades a directory for use with crAIzy. The main `craizy` command should require initialization and direct users to run `craizy init` if not set up.

## User Stories

### Story 1: Initialize Fresh Directory
As a user, I want to run `craizy init` in a new project directory so that crAIzy is properly configured.

**Acceptance Criteria:**
- Running `craizy init` in an uninitialized directory performs all setup steps
- Each step is logged using the logging framework
- User is prompted for git init (if not already a repo)
- Process exits cleanly if user declines git init

### Story 2: Run craizy Without Init
As a user, I want a clear error message if I run `craizy` without initializing first.

**Acceptance Criteria:**
- Running `craizy` in an uninitialized directory shows: `This directory is not initialized. Run 'craizy init' first.`
- Does not prompt inline - just exits with the message

### Story 3: Idempotent Init (Upgrade Path)
As a user, I want to run `craizy init` multiple times safely so that I can upgrade after updating crAIzy.

**Acceptance Criteria:**
- Running `craizy init` in an already-initialized directory skips completed steps
- Only performs actions that haven't been done yet
- Reports what was skipped vs what was done

## Technical Specification

### Init Steps (in order)

Each step must detect if already done and skip if so:

| Step | Check | Action if needed |
|------|-------|------------------|
| 1. Git repo | `git rev-parse --git-dir` succeeds | Prompt user, run `git init` if yes, exit if no |
| 2. .gitignore entry | `.craizy/` line exists in .gitignore | Append `.craizy/` to .gitignore (create if needed) |
| 3. .craizy directory | Directory exists | Create `.craizy/` |
| 4. AGENTS.yml | `.craizy/AGENTS.yml` exists | Copy embedded default AGENTS.yml |
| 5. Initial commit | Has at least one commit | `git commit --allow-empty -m "crAIzy init"` |

### Default AGENTS.yml

Embed the project's root `AGENTS.yml` at build time. During init, copy this to `.craizy/AGENTS.yml` if it doesn't exist.

Current default content:
```yaml
agents:
  - name: Claude
    command: claude --dangerously-skip-permissions
  - name: Gemini
    command: gemini --yolo
  - name: Copilot
    command: copilot --allow-all-tools
```

### Config Path Change

Update `internal/config/config.go` to look for AGENTS.yml at `.craizy/AGENTS.yml` instead of project root.

### Command Structure

```
craizy init    # Initialize/upgrade crAIzy in current directory
craizy         # Run the TUI (requires initialization)
```

Use Go's flag package or a simple subcommand check in main.go.

### Expected Output

```
$ craizy init
Checking git repository... not found
Initialize git repository? [Y/n] y
✓ Initialized git repository

Checking .gitignore... adding .craizy/
✓ Updated .gitignore

Checking .craizy directory... not found
✓ Created .craizy/

Checking .craizy/AGENTS.yml... not found
✓ Created default AGENTS.yml

Creating initial commit...
✓ Created commit: "crAIzy init"

Ready! Run 'craizy' to start.
```

Already initialized:
```
$ craizy init
Checking git repository... ✓ exists
Checking .gitignore... ✓ already configured
Checking .craizy directory... ✓ exists
Checking .craizy/AGENTS.yml... ✓ exists
Checking git commits... ✓ has commits

Already initialized. Nothing to do.
```

### Files to Modify/Create

| File | Action |
|------|--------|
| `cmd/craizy/main.go` | Add subcommand handling, remove inline git prompt |
| `cmd/craizy/init.go` | New - implement init command |
| `internal/config/config.go` | Change AGENTS.yml path to `.craizy/AGENTS.yml` |
| `internal/config/embed.go` | New - embed default AGENTS.yml |
| `AGENTS.yml` | Keep in root as source for embedding |

### Logging

All init steps should use the logging framework:
```go
logging.Info("git repository initialized")
logging.Debug("skipping .gitignore, already configured")
```

## Out of Scope

- `craizy doctor` command (future feature)
- Migration of existing root AGENTS.yml to .craizy/ (users move manually if needed)
- Remote/shared configuration

## Verification

1. `go test ./...` passes
2. Fresh directory: `craizy` shows init message, `craizy init` works
3. Re-run `craizy init` - skips all steps
4. `craizy` launches TUI after init
5. Agent creation works with new AGENTS.yml location
