Epic: MVP

# Git Integration

Dependencies: none

## Description

Integrate Git workflows into crAIzy to enable isolated agent workspaces via worktrees. Each agent operates in its own worktree/branch, preventing conflicts between concurrent agents and enabling clean merge workflows back to the main checkout.

## Stories

### Startup Git Check

As a user, when I start crAIzy in a non-git directory, I am prompted to initialize git. If I decline, the application exits with an error.

#### Technical / Architecture

- On startup, before TUI renders:
  1. Check if current directory is a git repo (`git rev-parse --git-dir`)
  2. If not a repo, prompt: "This directory is not a git repository. Initialize git? [Y/n]"
  3. If yes: run `git init`, continue startup
  4. If no: exit with error message explaining git is required

- Implementation in `main.go` before `tea.NewProgram()`:
  ```go
  func ensureGitRepo(dir string) error {
      cmd := exec.Command("git", "rev-parse", "--git-dir")
      cmd.Dir = dir
      if err := cmd.Run(); err != nil {
          // Not a git repo - prompt user
          if !promptGitInit() {
              return fmt.Errorf("crAIzy requires a git repository to manage agent worktrees")
          }
          return exec.Command("git", "init").Run()
      }
      return nil
  }
  ```

- Problematic git states (uncommitted changes, mid-merge, detached HEAD) are ignored at startup - trust the AI CLI to detect and warn if relevant

### New Agent Creates Worktree

As a user, when I create a new agent, a git worktree is created for that agent, and the agent's tmux session starts in that worktree directory.

#### Technical / Architecture

- Worktree location: `.craizy/worktrees/{agent-name}/`
- Branch name: `{agent-name}` (same as the agent name, sanitized)
- Base branch: current checked-out branch at time of agent creation

- Extend `AgentService.Create()`:
  ```go
  func (s *AgentService) Create(project, agentType, name, command string) (*Agent, error) {
      // 1. Validate agent name uniqueness (existing logic)

      // 2. Create worktree
      worktreePath := filepath.Join(".craizy", "worktrees", name)
      baseBranch, _ := s.git.CurrentBranch()

      if err := s.git.CreateWorktree(worktreePath, name, baseBranch); err != nil {
          return nil, fmt.Errorf("failed to create worktree: %w", err)
      }

      // 3. Create tmux session in worktree directory (not root)
      absWorktreePath, _ := filepath.Abs(worktreePath)
      if err := s.tmux.CreateSession(sessionID, command, absWorktreePath); err != nil {
          // Cleanup worktree on failure
          s.git.RemoveWorktree(worktreePath)
          return nil, err
      }

      // 4. Store agent with worktree info
      agent := &Agent{
          ID:          sessionID,
          Name:        name,
          WorkDir:     absWorktreePath,
          Branch:      name,
          BaseBranch:  baseBranch,
          // ... other fields
      }
      s.store.Add(agent)

      return agent, nil
  }
  ```

- New git interface:
  ```go
  type IGitClient interface {
      IsRepo() bool
      Init() error
      CurrentBranch() (string, error)
      CreateWorktree(path, branch, baseBranch string) error
      RemoveWorktree(path string) error
      HasUncommittedChanges(path string) bool
      Merge(branch string) error
      Stash() error
      StashPop() error
  }
  ```

- Git commands used:
  ```bash
  # Create worktree with new branch based on current branch
  git worktree add -b {branch-name} {path} HEAD

  # Remove worktree
  git worktree remove {path}

  # Delete branch after worktree removed
  git branch -d {branch-name}
  ```

- If branch already exists: error modal "Branch '{name}' already exists. Choose a different agent name.", cancel agent creation

- Agent model extended:
  ```go
  type Agent struct {
      // ... existing fields
      Branch     string  // The worktree branch name
      BaseBranch string  // Branch it was created from (for merge target)
  }
  ```

### Merge Agent Work

As a user, when I press `m` on the dashboard, the selected agent's branch is merged into the root checkout's current branch.

#### Technical / Architecture

- New quick command: `m - merge agent`

- On `m` keypress:
  1. Get currently selected agent
  2. Check if root checkout has uncommitted changes
     - If yes: stash changes, show alert "Your uncommitted changes have been stashed"
  3. Merge agent's branch into root checkout's current branch
  4. If conflict: show modal with instructions, abort merge, unstash if needed
  5. If success: show success modal, worktree is now "clean" (can be killed without warning)

- Merge flow:
  ```go
  func (s *AgentService) MergeAgent(agent *Agent) error {
      // 1. Check for uncommitted changes in root
      if s.git.HasUncommittedChanges(".") {
          s.git.Stash()
          // Alert user via returned message
      }

      // 2. Merge the agent's branch
      if err := s.git.Merge(agent.Branch); err != nil {
          // Conflict or other error
          s.git.MergeAbort()
          if stashed {
              s.git.StashPop()
          }
          return fmt.Errorf("merge conflict - resolve manually in root checkout")
      }

      // 3. Success - stash pop if needed
      if stashed {
          s.git.StashPop()
      }

      return nil
  }
  ```

- Git commands used:
  ```bash
  # Check for uncommitted changes
  git status --porcelain

  # Stash
  git stash push -m "craizy: auto-stash before merge"

  # Merge (regular merge commit)
  git merge {branch-name}

  # On conflict
  git merge --abort

  # Restore stash
  git stash pop
  ```

- Merge uses regular merge commit (not squash) to preserve history
- User controls granularity by choosing which branch to checkout before spawning agents

### Kill Agent with Uncommitted Changes

As a user, when I press `k` to kill an agent that has uncommitted changes in its worktree, I am prompted to Keep, Discard, or Cancel.

#### Technical / Architecture

- Modify kill flow:
  1. Check `git status --porcelain` in agent's worktree
  2. If clean: kill agent, remove worktree, delete branch (existing behavior)
  3. If dirty: show modal with options

- Modal options:
  - **Keep**: Cancel the kill, return to dashboard (agent stays alive)
  - **Discard**: Discard all changes (`git checkout .`), then kill agent, remove worktree, delete branch
  - **Cancel**: Same as Keep (dismiss modal, no action)

- Implementation:
  ```go
  func (s *AgentService) Kill(agent *Agent) (requiresConfirm bool, err error) {
      if s.git.HasUncommittedChanges(agent.WorkDir) {
          return true, nil // Signal TUI to show confirmation modal
      }
      return false, s.forceKill(agent)
  }

  func (s *AgentService) ForceKill(agent *Agent, discardChanges bool) error {
      if discardChanges {
          s.git.DiscardChanges(agent.WorkDir)
      }

      // 1. Kill tmux session
      s.tmux.KillSession(agent.ID)

      // 2. Remove worktree
      s.git.RemoveWorktree(agent.WorkDir)

      // 3. Delete branch
      s.git.DeleteBranch(agent.Branch)

      // 4. Remove from store
      s.store.Remove(agent.ID)

      return nil
  }
  ```

- Confirmation modal:
  ```go
  type KillConfirmModal struct {
      agent   *Agent
      message string  // "This worktree has uncommitted changes"
      options []string // ["Keep", "Discard", "Cancel"]
  }
  ```

### Developer Debug Mode

As a developer working on crAIzy, I can build a development version that can be tested from any git repository on the filesystem.

#### Technical / Architecture

- Add Makefile target:
  ```makefile
  # Install development build globally
  install-dev:
  	go build -o $(shell go env GOPATH)/bin/craizy-dev ./cmd/craizy
  	@echo "Installed craizy-dev to $(shell go env GOPATH)/bin/"
  	@echo "Run 'craizy-dev' from any git repository to test"
  ```

- Development workflow:
  1. Make changes in a worktree (e.g., `.craizy/worktrees/new-feature/`)
  2. From that worktree, run `make install-dev`
  3. Navigate to any test git repository
  4. Run `craizy-dev` to test the new version
  5. Repeat for other worktrees as needed

- The installed binary uses the code from whichever worktree it was built in
- Test repositories can be throwaway GitHub repos cloned anywhere on the filesystem

## Out of Scope

- GitHub/GitLab API integration for PR creation
- Multi-repository support (agents across different repos)
- Automatic conflict resolution
- Rebase workflows (merge commits only)
- Read-only agents without worktrees (deferred to future feature)
