package main

import (
	"fmt"
	"os"

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

func printUsage() {
	fmt.Println("crAIzy - AI-powered terminal multiplexer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  craizy init <name>    Initialize a new crAIzy project")
	fmt.Println("  craizy start          Start the dashboard")
	fmt.Println("  craizy version        Show version")
	fmt.Println("  craizy help           Show this help")
}
