package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/infra"
	"github.com/TechnicallyShaun/crAIzy/internal/infra/store"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
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

	// Initialize logging to .craizy directory
	logDir := filepath.Join(workDir, ".craizy")
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

	// Ensure we're in a git repository
	if err := ensureGitRepo(gitClient, workDir); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

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

// ensureGitRepo checks if the directory is a git repository.
// If not, it prompts the user to initialize one.
func ensureGitRepo(git *infra.GitClient, dir string) error {
	if git.IsRepo(dir) {
		return nil
	}

	fmt.Print("This directory is not a git repository. Initialize git? [Y/n] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "" || response == "y" || response == "yes" {
		if err := git.Init(dir); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}
		fmt.Println("Initialized git repository.")
		return nil
	}

	return fmt.Errorf("crAIzy requires a git repository to manage agent worktrees")
}
