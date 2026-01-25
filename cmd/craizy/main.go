package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/infra"
	"github.com/TechnicallyShaun/crAIzy/internal/infra/store"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
	"github.com/TechnicallyShaun/crAIzy/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check for subcommands first (before flag parsing)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInitCommand()
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}

	// Parse flags for the main TUI command
	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message")
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Run the main TUI
	runTUI()
}

func printHelp() {
	fmt.Println("Usage: craizy [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init        Initialize crAIzy in the current directory")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Run 'craizy' without arguments to start the TUI.")
}

func runInitCommand() {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging (create .craizy dir first if needed for logging)
	logDir := filepath.Join(workDir, ".craizy")
	_ = os.MkdirAll(logDir, 0755) // Ignore error, init will create it properly
	if err := logging.Init(logDir); err != nil {
		// Don't fail on logging init during init command
		fmt.Printf("Warning: logging not available: %v\n", err)
	}
	defer logging.Close()

	logging.Info("craizy init starting, workDir=%s", workDir)

	if err := runInit(workDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		logging.Error(err, "command", "init")
		os.Exit(1)
	}
}

func runTUI() {
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Check if initialized
	if !isInitialized(workDir) {
		fmt.Println("This directory is not initialized. Run 'craizy init' first.")
		os.Exit(1)
	}

	// Detect project name (parent folder of cwd)
	project := filepath.Base(workDir)

	// Initialize logging to .craizy directory
	logDir := config.CraizyDirPath(workDir)
	if err := logging.Init(logDir); err != nil {
		fmt.Printf("Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	defer logging.Close()
	logging.Info("crAIzy starting, project=%s, workDir=%s", project, workDir)

	// Create database directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get home directory: %v\n", err)
		os.Exit(1)
	}
	dbDir := filepath.Join(homeDir, ".craizy")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Printf("Failed to create database directory: %v\n", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(dbDir, "craizy.db")

	// Initialize infrastructure
	tmuxClient := infra.NewTmuxClient()
	gitClient := infra.NewGitClient(workDir)

	// Initialize SQLite store
	agentStore, err := store.NewSQLiteAgentStore(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer agentStore.Close()

	// Initialize event dispatcher and wire adapters
	dispatcher := infra.NewEventDispatcher()
	infra.WireAdapters(dispatcher, agentStore, tmuxClient, gitClient)

	// Initialize service
	agentService := domain.NewAgentService(tmuxClient, agentStore, dispatcher, gitClient, project, workDir)

	// Reconcile any zombie sessions before starting
	_ = agentService.Reconcile()

	// Start TUI with the agent service
	p := tea.NewProgram(tui.NewModel(agentService))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
