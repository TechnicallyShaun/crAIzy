# Feature Implementation Workflow

You are working on the feature file provided alongside this prompt. Follow this autonomous workflow:

## 1. Setup

- Read and understand the feature specification completely
- Create a git worktree for this feature:
  ```bash
  git worktree add -b feature/<feature-name> .craizy/worktrees/<feature-name> HEAD
  ```
- Change your working directory to the worktree

## 2. Plan

- Analyze the feature's stories and technical requirements
- Identify affected files and packages
- Create a implementation plan with discrete, testable steps
- Note any dependencies or prerequisites

## 3. Test First

- Write tests for the feature before implementing
- Tests should cover:
  - Happy path for each story
  - Edge cases mentioned in the spec
  - Error conditions
- Ensure tests fail initially (red phase)

## 4. Implement

- Work through your plan systematically
- Implement one story at a time
- Commit frequently with clear messages
- Follow existing code patterns in the codebase

## 5. Verify

Run verification in this order:

1. **Run tests**
   ```bash
   go test ./...
   ```

2. **Run the application** (as background task to verify it builds and starts)
   ```bash
   go build -o /tmp/craizy-test ./cmd/craizy && /tmp/craizy-test &
   ```
   Then kill the background process after confirming startup.

3. **Lint/vet** (if configured)
   ```bash
   go vet ./...
   ```

## 6. Report

When complete, report back with:
- Summary of what was implemented
- Any deviations from the spec (and why)
- Test results
- Any open questions or follow-up work identified

**Wait for human confirmation before merging.**

## 7. Finalize (after human approval)

Once approved:
1. From the main checkout, merge the feature branch:
   ```bash
   git merge feature/<feature-name>
   ```
2. Remove the worktree:
   ```bash
   git worktree remove .craizy/worktrees/<feature-name>
   ```
3. Delete the feature branch:
   ```bash
   git branch -d feature/<feature-name>
   ```
4. Move the feature spec to complete:
   ```bash
   mv roadmap/features/<feature-name>.md roadmap/features/complete/
   ```
5. Commit the move

---

**Start now: Read the feature file provided and begin with Step 1.**
