package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/tmux"
	"github.com/TechnicallyShaun/crAIzy/internal/ui"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Error: project name required")
			fmt.Println("Usage: craizy init <name>")
			os.Exit(1)
		}
		handleInit(os.Args[2])
	case "start":
		handleStart()
	case "agent":
		if len(os.Args) < 3 {
			fmt.Println("Error: agent subcommand required")
			fmt.Println("Usage: craizy agent <add|list|remove>")
			os.Exit(1)
		}
		handleAgent(os.Args[2:])
	case "version":
		fmt.Printf("crAIzy v%s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit(name string) {
	if err := config.InitProject(name); err != nil {
		fmt.Printf("Error initializing project: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Initialized crAIzy project: %s\n", name)
}

func handleStart() {
	// Check if .craizy directory exists
	if !config.IsInitialized() {
		fmt.Println("Error: not in a crAIzy project")
		fmt.Println("Run 'craizy init <name>' first")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create tmux manager
	tmuxMgr := tmux.NewManager()

	// Start the dashboard
	dashboard := ui.NewDashboard(cfg, tmuxMgr)
	if err := dashboard.Start(); err != nil {
		fmt.Printf("Error starting dashboard: %v\n", err)
		os.Exit(1)
	}
}

func handleAgent(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: agent subcommand required")
		fmt.Println("Usage: craizy agent <add|list|remove>")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "add":
		if len(args) < 2 {
			fmt.Println("Error: command required")
			fmt.Println("Usage: craizy agent add <command>")
			fmt.Println("Example: craizy agent add \"claude --dangerously-skip-permissions\"")
			os.Exit(1)
		}
		handleAgentAdd(args[1])
	case "list":
		handleAgentList()
	case "remove":
		if len(args) < 2 {
			fmt.Println("Error: agent name required")
			fmt.Println("Usage: craizy agent remove <name>")
			os.Exit(1)
		}
		handleAgentRemove(args[1])
	default:
		fmt.Printf("Unknown agent subcommand: %s\n", subcommand)
		fmt.Println("Usage: craizy agent <add|list|remove>")
		os.Exit(1)
	}
}

func handleAgentAdd(command string) {
	// Parse command to extract name (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		fmt.Println("Error: invalid command")
		os.Exit(1)
	}

	name := strings.Title(parts[0])

	if err := config.AddAgent(name, command); err != nil {
		fmt.Printf("Error adding agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added agent: %s\n", name)
	fmt.Printf("  Command: %s\n", command)
}

func handleAgentList() {
	agents, err := config.ListAgents()
	if err != nil {
		fmt.Printf("Error listing agents: %v\n", err)
		os.Exit(1)
	}

	if len(agents) == 0 {
		fmt.Println("No agents configured")
		return
	}

	fmt.Println("Configured agents:")
	for i, agent := range agents {
		fmt.Printf("  %d. %s\n", i+1, agent.Name)
		fmt.Printf("     Command: %s\n", agent.Command)
	}
}

func handleAgentRemove(name string) {
	if err := config.RemoveAgent(name); err != nil {
		fmt.Printf("Error removing agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed agent: %s\n", name)
}

func printUsage() {
	fmt.Println("crAIzy - AI-powered terminal multiplexer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  craizy init <name>              Initialize a new crAIzy project")
	fmt.Println("  craizy start                    Start the dashboard")
	fmt.Println("  craizy agent add <command>      Add a new agent")
	fmt.Println("  craizy agent list               List all agents")
	fmt.Println("  craizy agent remove <name>      Remove an agent")
	fmt.Println("  craizy version                  Show version")
	fmt.Println("  craizy help                     Show this help")
}
