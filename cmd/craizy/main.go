package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/infra"
	"github.com/TechnicallyShaun/crAIzy/internal/infra/store"
	"github.com/TechnicallyShaun/crAIzy/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message")
	flag.Parse()

	if *help {
		fmt.Println("Usage: craizy [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		return
	}

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Detect project name (parent folder of cwd)
	project := filepath.Base(workDir)

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

	// Initialize SQLite store
	agentStore, err := store.NewSQLiteAgentStore(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer agentStore.Close()

	// Initialize event dispatcher and wire adapters
	dispatcher := infra.NewEventDispatcher()
	infra.WireAdapters(dispatcher, agentStore, tmuxClient)

	// Initialize service
	agentService := domain.NewAgentService(tmuxClient, agentStore, dispatcher, project, workDir)

	// Reconcile any zombie sessions before starting
	_ = agentService.Reconcile()

	// Start TUI with the agent service
	p := tea.NewProgram(tui.NewModel(agentService))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
