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
	showAIModal bool
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
	sb.WriteString("echo '═══ Available AIs ═══'\n")
	for i, ai := range d.config.AIs {
		sb.WriteString(fmt.Sprintf("echo '  %d. %s - %s'\n", i+1, ai.Name, ai.Command))
	}
	sb.WriteString("echo ''\n")

	// Interactive prompt
	sb.WriteString("echo 'Commands: [n]ew AI, [a]ttach <id>, [k]ill <id>, [l]ist, [q]uit'\n")
	sb.WriteString("read -p '> ' cmd args\n")
	sb.WriteString("case $cmd in\n")
	sb.WriteString("  n) echo 'Starting new AI...'; ;;\n")
	sb.WriteString("  a) echo 'Attaching to AI...'; ;;\n")
	sb.WriteString("  k) echo 'Killing AI...'; ;;\n")
	sb.WriteString("  l) echo 'Listing AIs...'; ;;\n")
	sb.WriteString("  q) exit 0; ;;\n")
	sb.WriteString("  *) echo 'Unknown command'; ;;\n")
	sb.WriteString("esac\n")
	sb.WriteString("sleep 2\n")

	return sb.String()
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

// SpawnAI spawns a new AI instance
func (d *Dashboard) SpawnAI(aiSpec config.AISpec) (*AIInstance, error) {
	instanceName := fmt.Sprintf("%s-%d", aiSpec.Name, len(d.aiInstances)+1)

	session, err := d.tmuxMgr.CreateSession(instanceName, aiSpec.Command)
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
