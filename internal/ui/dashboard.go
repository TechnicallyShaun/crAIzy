package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
)

// Dashboard represents the main UI dashboard
type Dashboard struct {
	config      *config.Config
	tmuxMgr     *tmux.Manager
	aiInstances []*AIInstance
	selectedTab int
}

// AIInstance represents a running AI instance
type AIInstance struct {
	ID      int
	Name    string
	Session *tmux.Session
}

// NewDashboard creates a new dashboard
func NewDashboard(cfg *config.Config, tmuxMgr *tmux.Manager) *Dashboard {
	return &Dashboard{
		config:      cfg,
		tmuxMgr:     tmuxMgr,
		aiInstances: make([]*AIInstance, 0),
		selectedTab: -1,
	}
}

// Start starts the dashboard
func (d *Dashboard) Start() error {
	// Check if tmux is available
	if !tmux.IsTmuxAvailable() {
		return fmt.Errorf("tmux is not installed or not in PATH")
	}

	// Create main dashboard session
	sessionName := fmt.Sprintf("craizy-dashboard-%s", d.config.ProjectName)

	// Check if we're already in a tmux session
	if os.Getenv("TMUX") != "" {
		// We're inside tmux, create a new window
		return d.startInWindow()
	}

	// Create new tmux session for dashboard
	return d.startInSession(sessionName)
}

func (d *Dashboard) startInSession(sessionName string) error {
	// Create a bash script for the dashboard
	script := d.generateDashboardScript()

	// Create session and run dashboard
	cmd := exec.Command("tmux", "new-session", "-s", sessionName, "bash", "-c", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *Dashboard) startInWindow() error {
	// Create new window in current session
	script := d.generateDashboardScript()
	cmd := exec.Command("tmux", "new-window", "bash", "-c", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *Dashboard) generateDashboardScript() string {
	var sb strings.Builder

	// Wrap everything in a loop
	sb.WriteString("while true; do\n")

	// Clear screen
	sb.WriteString("clear\n")

	// Display header
	sb.WriteString("echo '╔════════════════════════════════════════════════════════════════╗'\n")
	sb.WriteString(fmt.Sprintf("echo '║  crAIzy Dashboard - %s  ║'\n", padRight(d.config.ProjectName, 42)))
	sb.WriteString("echo '╠════════════════════════════════════════════════════════════════╣'\n")
	sb.WriteString("echo '║  Hotkeys: [N] New AI  [Q] Quit  [L] List  [K] Kill           ║'\n")
	sb.WriteString("echo '╚════════════════════════════════════════════════════════════════╝'\n")
	sb.WriteString("echo ''\n")

	// Display AI tabs section
	sb.WriteString("echo '═══ AI Instances ═══'\n")
	if len(d.aiInstances) == 0 {
		sb.WriteString("echo '  No AI instances running. Press N to start one.'\n")
	} else {
		for i, ai := range d.aiInstances {
			sb.WriteString(fmt.Sprintf("echo '  [%d] %s'\n", i+1, ai.Name))
		}
	}
	sb.WriteString("echo ''\n")

	// Display preview section
	sb.WriteString("echo '═══ Preview Window ═══'\n")
	sb.WriteString("echo '  (AI output will appear here)'\n")
	sb.WriteString("echo ''\n")

	// Display available AIs
	sb.WriteString("echo '═══ Available Agents ═══'\n")
	for i, agent := range d.config.Agents {
		sb.WriteString(fmt.Sprintf("echo '  %d. %s - %s'\n", i+1, agent.Name, agent.Command))
	}
	sb.WriteString("echo ''\n")

	// Interactive prompt
	sb.WriteString("echo 'Commands: [n]ew AI, [a]ttach <id>, [k]ill <id>, [l]ist, [q]uit'\n")
	sb.WriteString("read -p '> ' cmd args\n")
	sb.WriteString("case $cmd in\n")
	
	// 'n' command: New AI - show agent menu and spawn selected agent
	sb.WriteString("  n)\n")
	sb.WriteString("    echo ''\n")
	sb.WriteString("    echo 'Select an agent to start:'\n")
	for i, agent := range d.config.Agents {
		sb.WriteString(fmt.Sprintf("    echo '  %d. %s - %s'\n", i+1, agent.Name, agent.Command))
	}
	sb.WriteString("    echo ''\n")
	sb.WriteString("    read -p 'Enter agent number: ' agent_num\n")
	sb.WriteString("    case $agent_num in\n")
	for i, agent := range d.config.Agents {
		// Escape any single quotes in the command
		escapedCmd := strings.ReplaceAll(agent.Command, "'", "'\"'\"'")
		windowName := fmt.Sprintf("craizy-%s", agent.Name)
		sb.WriteString(fmt.Sprintf("      %d)\n", i+1))
		sb.WriteString(fmt.Sprintf("        echo 'Starting %s...'\n", agent.Name))
		sb.WriteString(fmt.Sprintf("        tmux new-window -n '%s' '%s'\n", windowName, escapedCmd))
		sb.WriteString("        sleep 1\n")
		sb.WriteString("        ;;\n")
	}
	sb.WriteString("      *)\n")
	sb.WriteString("        echo 'Invalid selection'\n")
	sb.WriteString("        sleep 2\n")
	sb.WriteString("        ;;\n")
	sb.WriteString("    esac\n")
	sb.WriteString("    ;;\n")
	
	// 'a' command: Attach - list windows and attach to selected one
	sb.WriteString("  a)\n")
	sb.WriteString("    echo ''\n")
	sb.WriteString("    echo 'Available windows:'\n")
	sb.WriteString("    tmux list-windows -F '#I: #W'\n")
	sb.WriteString("    echo ''\n")
	sb.WriteString("    read -p 'Enter window ID to attach: ' window_id\n")
	sb.WriteString("    if tmux select-window -t :$window_id 2>/dev/null; then\n")
	sb.WriteString("      echo \"Switched to window $window_id\"\n")
	sb.WriteString("      sleep 1\n")
	sb.WriteString("    else\n")
	sb.WriteString("      echo \"Failed to switch to window $window_id\"\n")
	sb.WriteString("      sleep 2\n")
	sb.WriteString("    fi\n")
	sb.WriteString("    ;;\n")
	
	// 'k' command: Kill - list windows and kill selected one
	sb.WriteString("  k)\n")
	sb.WriteString("    echo ''\n")
	sb.WriteString("    echo 'Available windows:'\n")
	sb.WriteString("    tmux list-windows -F '#I: #W'\n")
	sb.WriteString("    echo ''\n")
	sb.WriteString("    read -p 'Enter window ID to kill: ' window_id\n")
	sb.WriteString("    if tmux kill-window -t :$window_id 2>/dev/null; then\n")
	sb.WriteString("      echo \"Killed window $window_id\"\n")
	sb.WriteString("      sleep 1\n")
	sb.WriteString("    else\n")
	sb.WriteString("      echo \"Failed to kill window $window_id\"\n")
	sb.WriteString("      sleep 2\n")
	sb.WriteString("    fi\n")
	sb.WriteString("    ;;\n")
	
	sb.WriteString("  l) echo 'Listing AIs...'; sleep 2; ;;\n")
	sb.WriteString("  q) echo 'Goodbye!'; exit 0; ;;\n")
	sb.WriteString("  *) echo 'Unknown command'; sleep 2; ;;\n")
	sb.WriteString("esac\n")

	// Close the loop
	sb.WriteString("done\n")

	return sb.String()
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

// SpawnAI spawns a new AI instance
func (d *Dashboard) SpawnAI(agent config.Agent) (*AIInstance, error) {
	instanceName := fmt.Sprintf("%s-%d", agent.Name, len(d.aiInstances)+1)

	// CreateWindow creates a new tmux window and sends the CLI command to it
	session, err := d.tmuxMgr.CreateWindow(instanceName, agent.Command)
	if err != nil {
		return nil, err
	}

	ai := &AIInstance{
		ID:      len(d.aiInstances) + 1,
		Name:    instanceName,
		Session: session,
	}

	d.aiInstances = append(d.aiInstances, ai)
	return ai, nil
}

// GetAIInstance returns an AI instance by ID
func (d *Dashboard) GetAIInstance(id int) *AIInstance {
	if id < 1 || id > len(d.aiInstances) {
		return nil
	}
	return d.aiInstances[id-1]
}

// StartDetached starts the dashboard in a detached tmux session for testing
func (d *Dashboard) StartDetached(sessionName string) (string, error) {
	// Check if tmux is available
	if !tmux.IsTmuxAvailable() {
		return "", fmt.Errorf("tmux is not installed or not in PATH")
	}

	// Create a bash script for the dashboard
	script := d.generateDashboardScript()

	// Create detached session and run dashboard
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "bash", "-c", script)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start detached dashboard: %w", err)
	}

	return sessionName, nil
}
