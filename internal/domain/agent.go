package domain

import (
	"regexp"
	"strings"
	"time"
)

// AgentStatus represents the lifecycle state of an agent.
type AgentStatus string

const (
	AgentStatusPending    AgentStatus = "pending"
	AgentStatusActive     AgentStatus = "active"
	AgentStatusTerminated AgentStatus = "terminated"
)

// Agent represents a running agent session in tmux.
type Agent struct {
	ID           string       // tmux session ID: craizy-{project}-{agent}-{name}
	Project      string       // parent folder name
	AgentType    string       // from AGENTS.yml (lowercase)
	Name         string       // user-entered name (sanitized)
	Command      string       // agent command to run
	WorkDir      string       // working directory
	Status       AgentStatus  // current lifecycle status
	CreatedAt    time.Time
	TerminatedAt *time.Time   // when the agent was terminated (nil if still active)
	Branch       string       // worktree branch name
	BaseBranch   string       // branch it was created from
}

// BuildSessionID creates a unique tmux session ID from the components.
func BuildSessionID(project, agentType, name string) string {
	return "craizy-" + SanitizeName(project) + "-" + SanitizeName(agentType) + "-" + SanitizeName(name)
}

// SanitizeName converts a name to a tmux-safe format.
// - Converts to lowercase
// - Removes periods and colons
// - Replaces spaces with hyphens
// - Removes any non-alphanumeric characters except hyphens
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any character that's not alphanumeric or hyphen
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Remove consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	return name
}
