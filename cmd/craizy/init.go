package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
)

// InitResult tracks what actions were taken during init.
type InitResult struct {
	GitInitialized     bool
	GitIgnoreUpdated   bool
	CraizyDirCreated   bool
	AgentsYMLCreated   bool
	InitialCommitMade  bool
	AlreadyInitialized bool
}

// runInit performs the craizy init command.
func runInit(workDir string) error {
	fmt.Println("Initializing crAIzy...")
	fmt.Println()

	result := &InitResult{}

	// Step 1: Check/init git repository
	if err := initGitRepo(workDir, result); err != nil {
		return err
	}

	// Step 2: Check/update .gitignore
	if err := initGitIgnore(workDir, result); err != nil {
		return err
	}

	// Step 3: Check/create .craizy directory
	if err := initCraizyDir(workDir, result); err != nil {
		return err
	}

	// Step 4: Check/create AGENTS.yml
	if err := initAgentsYML(workDir, result); err != nil {
		return err
	}

	// Step 5: Check/create initial commit
	if err := initInitialCommit(workDir, result); err != nil {
		return err
	}

	// Summary
	fmt.Println()
	if result.AlreadyInitialized {
		fmt.Println("Already initialized. Nothing to do.")
	} else {
		fmt.Println("Ready! Run 'craizy' to start.")
	}

	return nil
}

// initGitRepo checks for and optionally initializes a git repository.
func initGitRepo(workDir string, result *InitResult) error {
	fmt.Print("Checking git repository... ")

	cmd := exec.Command("git", "-C", workDir, "rev-parse", "--git-dir")
	if cmd.Run() == nil {
		fmt.Println("exists")
		logging.Debug("git repository already exists")
		return nil
	}

	fmt.Println("not found")
	fmt.Print("Initialize git repository? [Y/n] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" && response != "yes" {
		return fmt.Errorf("crAIzy requires a git repository. Initialization canceled")
	}

	cmd = exec.Command("git", "init", workDir)
	if err := cmd.Run(); err != nil {
		logging.Error(err, "action", "git init")
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	fmt.Println("Initialized git repository")
	logging.Info("git repository initialized, workDir=%s", workDir)
	result.GitInitialized = true
	return nil
}

// initGitIgnore ensures .craizy/ is in .gitignore.
func initGitIgnore(workDir string, result *InitResult) error {
	fmt.Print("Checking .gitignore... ")

	gitignorePath := filepath.Join(workDir, ".gitignore")
	entry := ".craizy/"

	// Check if .gitignore exists and contains the entry
	content, err := os.ReadFile(gitignorePath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == entry {
				fmt.Println("already configured")
				logging.Debug(".gitignore already has .craizy/ entry")
				return nil
			}
		}
	}

	// Add entry to .gitignore
	fmt.Println("adding .craizy/")

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		logging.Error(err, "action", "open .gitignore")
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	// Add newline before entry if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			logging.Error(err, "action", "write newline to .gitignore")
			return fmt.Errorf("failed to write to .gitignore: %w", err)
		}
	}

	if _, err := f.WriteString(entry + "\n"); err != nil {
		logging.Error(err, "action", "write .craizy/ to .gitignore")
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	fmt.Println("Updated .gitignore")
	logging.Info(".gitignore updated with .craizy/ entry")
	result.GitIgnoreUpdated = true
	return nil
}

// initCraizyDir ensures the .craizy directory exists.
func initCraizyDir(workDir string, result *InitResult) error {
	fmt.Print("Checking .craizy directory... ")

	craizyDir := config.CraizyDirPath(workDir)

	if info, err := os.Stat(craizyDir); err == nil && info.IsDir() {
		fmt.Println("exists")
		logging.Debug(".craizy directory already exists")
		return nil
	}

	fmt.Println("not found")

	if err := os.MkdirAll(craizyDir, 0o755); err != nil {
		logging.Error(err, "action", "create .craizy directory")
		return fmt.Errorf("failed to create .craizy directory: %w", err)
	}

	fmt.Println("Created .craizy/")
	logging.Info(".craizy directory created, path=%s", craizyDir)
	result.CraizyDirCreated = true
	return nil
}

// initAgentsYML ensures AGENTS.yml exists in .craizy/.
func initAgentsYML(workDir string, result *InitResult) error {
	fmt.Print("Checking .craizy/AGENTS.yml... ")

	agentsPath := config.AgentsPath(workDir)

	if _, err := os.Stat(agentsPath); err == nil {
		fmt.Println("exists")
		logging.Debug("AGENTS.yml already exists")
		return nil
	}

	fmt.Println("not found")

	if err := os.WriteFile(agentsPath, config.DefaultAgentsYML, 0o644); err != nil {
		logging.Error(err, "action", "create AGENTS.yml")
		return fmt.Errorf("failed to create AGENTS.yml: %w", err)
	}

	fmt.Println("Created default AGENTS.yml")
	logging.Info("AGENTS.yml created, path=%s", agentsPath)
	result.AgentsYMLCreated = true
	return nil
}

// initInitialCommit creates an initial commit if the repo has none.
func initInitialCommit(workDir string, result *InitResult) error {
	fmt.Print("Checking git commits... ")

	// Check if there are any commits
	cmd := exec.Command("git", "-C", workDir, "rev-parse", "HEAD")
	if cmd.Run() == nil {
		fmt.Println("has commits")
		logging.Debug("repository already has commits")
		result.AlreadyInitialized = !result.GitInitialized && !result.GitIgnoreUpdated &&
			!result.CraizyDirCreated && !result.AgentsYMLCreated
		return nil
	}

	fmt.Println("no commits")
	fmt.Print("Creating initial commit... ")

	// Stage .gitignore if it was created/modified
	cmd = exec.Command("git", "-C", workDir, "add", ".gitignore")
	_ = cmd.Run() // Ignore error if .gitignore doesn't exist

	// Create initial commit
	cmd = exec.Command("git", "-C", workDir, "commit", "--allow-empty", "-m", "crAIzy init")
	if err := cmd.Run(); err != nil {
		logging.Error(err, "action", "initial commit")
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	fmt.Println("done")
	logging.Info("initial commit created")
	result.InitialCommitMade = true
	return nil
}

// isInitialized checks if crAIzy has been initialized in the given directory.
func isInitialized(workDir string) bool {
	// Check for .craizy directory
	craizyDir := config.CraizyDirPath(workDir)
	if _, err := os.Stat(craizyDir); os.IsNotExist(err) {
		return false
	}

	// Check for AGENTS.yml
	agentsPath := config.AgentsPath(workDir)
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return false
	}

	// Check for git repo
	cmd := exec.Command("git", "-C", workDir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
