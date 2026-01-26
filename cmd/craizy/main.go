package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/TechnicallyShaun/crAIzy/internal/config"
	"github.com/TechnicallyShaun/crAIzy/internal/domain"
	"github.com/TechnicallyShaun/crAIzy/internal/infra"
	"github.com/TechnicallyShaun/crAIzy/internal/infra/store"
	"github.com/TechnicallyShaun/crAIzy/internal/logging"
	"github.com/TechnicallyShaun/crAIzy/internal/tui"
)

func main() {
	// Check for subcommands first (before flag parsing)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInitCommand()
			return
		case "msg":
			runMsgCommand()
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
	fmt.Println("  msg         Messaging commands (send, list, read, count)")
	fmt.Println("  help        Show this help message")
	fmt.Println()
	fmt.Println("Run 'craizy' without arguments to start the TUI.")
	fmt.Println("Run 'craizy msg help' for messaging commands.")
}

func runInitCommand() {
	exitCode := runInitCommandInner()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func runInitCommandInner() int {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		return 1
	}

	// Initialize logging (create .craizy dir first if needed for logging)
	logDir := filepath.Join(workDir, ".craizy")
	_ = os.MkdirAll(logDir, 0o755) // Ignore error, init will create it properly
	if err := logging.Init(logDir); err != nil {
		// Don't fail on logging init during init command
		fmt.Printf("Warning: logging not available: %v\n", err)
	}
	defer logging.Close()

	logging.Info("craizy init starting, workDir=%s", workDir)

	if err := runInit(workDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		logging.Error(err, "command", "init")
		return 1
	}
	return 0
}

func runTUI() {
	exitCode := runTUIInner()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func runTUIInner() int {
	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		return 1
	}

	// Check if initialized
	if !isInitialized(workDir) {
		fmt.Println("This directory is not initialized. Run 'craizy init' first.")
		return 1
	}

	// Detect project name (parent folder of cwd)
	project := filepath.Base(workDir)

	// Initialize logging to .craizy directory
	logDir := config.CraizyDirPath(workDir)
	if initErr := logging.Init(logDir); initErr != nil {
		fmt.Printf("Failed to initialize logging: %v\n", initErr)
		return 1
	}
	defer logging.Close()
	logging.Info("crAIzy starting, project=%s, workDir=%s", project, workDir)

	// Create database directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get home directory: %v\n", err)
		return 1
	}
	dbDir := filepath.Join(homeDir, ".craizy")
	if mkdirErr := os.MkdirAll(dbDir, 0o755); mkdirErr != nil {
		fmt.Printf("Failed to create database directory: %v\n", mkdirErr)
		return 1
	}
	dbPath := filepath.Join(dbDir, "craizy.db")

	// Initialize infrastructure
	tmuxClient := infra.NewTmuxClient()
	gitClient := infra.NewGitClient(workDir)

	// Initialize SQLite store
	agentStore, err := store.NewSQLiteAgentStore(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return 1
	}
	defer agentStore.Close()

	// Initialize event dispatcher and wire adapters
	dispatcher := infra.NewEventDispatcher()
	infra.WireAdapters(dispatcher, agentStore, tmuxClient, gitClient)

	// Initialize message store and service
	messageStore := store.NewSQLiteMessageStore(agentStore.DB())
	messageService := domain.NewMessageService(messageStore, tmuxClient, agentStore)

	// Initialize agent service
	agentService := domain.NewAgentService(tmuxClient, agentStore, dispatcher, gitClient, project, workDir)
	agentService.SetMessageService(messageService)

	// Reconcile any zombie sessions before starting
	_ = agentService.Reconcile()

	// Start TUI with the agent service
	p := tea.NewProgram(tui.NewModel(agentService))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		return 1
	}
	return 0
}

// runMsgCommand handles the msg subcommand and its subcommands.
func runMsgCommand() {
	if len(os.Args) < 3 {
		printMsgHelp()
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "send":
		runMsgSend()
	case "list", "ls":
		runMsgList()
	case "read":
		runMsgRead()
	case "count":
		runMsgCount()
	case "help", "--help", "-h":
		printMsgHelp()
	default:
		fmt.Printf("Unknown msg subcommand: %s\n", subCmd)
		printMsgHelp()
		os.Exit(1)
	}
}

func printMsgHelp() {
	fmt.Println("Usage: craizy msg <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  send    Send a message")
	fmt.Println("  list    List messages (alias: ls)")
	fmt.Println("  read    Read a specific message")
	fmt.Println("  count   Count unread messages")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  craizy msg send --from worker-001 --to lead-001 --type question --content \"Which auth library?\"")
	fmt.Println("  craizy msg list --for worker-001")
	fmt.Println("  craizy msg list --for human --unread")
	fmt.Println("  craizy msg read <message-id>")
	fmt.Println("  craizy msg count --for human")
}

// initMsgServices initializes the services needed for messaging commands.
func initMsgServices() (*domain.MessageService, func(), error) {
	// Get database path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	dbDir := filepath.Join(homeDir, ".craizy")
	if mkdirErr := os.MkdirAll(dbDir, 0o755); mkdirErr != nil {
		return nil, nil, fmt.Errorf("failed to create database directory: %w", mkdirErr)
	}
	dbPath := filepath.Join(dbDir, "craizy.db")

	// Initialize stores
	agentStore, err := store.NewSQLiteAgentStore(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	messageStore := store.NewSQLiteMessageStore(agentStore.DB())
	tmuxClient := infra.NewTmuxClient()

	messageSvc := domain.NewMessageService(messageStore, tmuxClient, agentStore)

	cleanup := func() {
		agentStore.Close()
	}

	return messageSvc, cleanup, nil
}

func runMsgSend() {
	// Parse flags starting from os.Args[3:]
	fs := flag.NewFlagSet("msg send", flag.ExitOnError)
	from := fs.String("from", "", "Sender ID (required)")
	to := fs.String("to", "", "Recipient ID (required)")
	msgType := fs.String("type", "", "Message type: question, answer, assignment, completion, status, info (required)")
	content := fs.String("content", "", "Message content (required)")
	relatedWork := fs.String("related", "", "Related work item (optional)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Validate required flags
	if *from == "" || *to == "" || *msgType == "" || *content == "" {
		fmt.Println("Error: --from, --to, --type, and --content are required")
		fmt.Println()
		fmt.Println("Usage: craizy msg send --from <sender> --to <recipient> --type <type> --content \"message\"")
		os.Exit(1)
	}

	// Validate message type
	if !domain.IsValidMessageType(*msgType) {
		fmt.Printf("Error: invalid message type: %s\n", *msgType)
		fmt.Println("Valid types: question, answer, assignment, completion, status, info")
		os.Exit(1)
	}

	svc, cleanup, err := initMsgServices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	var relatedWorkPtr *string
	if *relatedWork != "" {
		relatedWorkPtr = relatedWork
	}

	msg, err := svc.Send(*from, *to, domain.MessageType(*msgType), *content, relatedWorkPtr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Message sent: %s\n", msg.ID)
}

func runMsgList() {
	fs := flag.NewFlagSet("msg list", flag.ExitOnError)
	forAgent := fs.String("for", "", "Recipient ID to list messages for (required)")
	unreadOnly := fs.Bool("unread", false, "Show only unread messages")

	if err := fs.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *forAgent == "" {
		fmt.Println("Error: --for is required")
		fmt.Println()
		fmt.Println("Usage: craizy msg list --for <recipient> [--unread]")
		os.Exit(1)
	}

	svc, cleanup, err := initMsgServices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	var messages []*domain.Message
	if *unreadOnly {
		messages, err = svc.ListUnread(*forAgent)
	} else {
		messages, err = svc.List(*forAgent, 0)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(messages) == 0 {
		if *unreadOnly {
			fmt.Println("No unread messages")
		} else {
			fmt.Println("No messages")
		}
		return
	}

	// Print messages in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFROM\tTYPE\tTIME\tCONTENT")

	var unreadCount int
	for _, msg := range messages {
		if !msg.Read {
			unreadCount++
		}
		// Truncate content for display
		content := msg.Content
		if len(content) > 40 {
			content = content[:37] + "..."
		}
		content = strings.ReplaceAll(content, "\n", " ")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			msg.ID[:8], // Show first 8 chars of ID
			msg.From,
			msg.Type,
			msg.CreatedAt.Format(time.DateTime),
			content,
		)
	}
	w.Flush()

	fmt.Printf("\n%d messages", len(messages))
	if unreadCount > 0 {
		fmt.Printf(" (%d unread)", unreadCount)
	}
	fmt.Println()
}

func runMsgRead() {
	if len(os.Args) < 4 {
		fmt.Println("Error: message ID required")
		fmt.Println()
		fmt.Println("Usage: craizy msg read <message-id>")
		os.Exit(1)
	}

	messageID := os.Args[3]

	svc, cleanup, err := initMsgServices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	msg, err := svc.Read(messageID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print message details
	fmt.Printf("From:    %s\n", msg.From)
	fmt.Printf("To:      %s\n", msg.To)
	fmt.Printf("Type:    %s\n", msg.Type)
	fmt.Printf("Time:    %s\n", msg.CreatedAt.Format(time.DateTime))
	if msg.RelatedWork != nil {
		fmt.Printf("Related: %s\n", *msg.RelatedWork)
	}

	fmt.Println()
	fmt.Println("Content:")
	fmt.Println(strings.Repeat("─", 35))
	fmt.Println(msg.Content)
	fmt.Println(strings.Repeat("─", 35))
	fmt.Println()
	fmt.Println("[Marked as read]")
}

func runMsgCount() {
	fs := flag.NewFlagSet("msg count", flag.ExitOnError)
	forAgent := fs.String("for", "", "Recipient ID to count messages for (required)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *forAgent == "" {
		fmt.Println("Error: --for is required")
		fmt.Println()
		fmt.Println("Usage: craizy msg count --for <recipient>")
		os.Exit(1)
	}

	svc, cleanup, err := initMsgServices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	count, err := svc.UnreadCount(*forAgent)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if count == 1 {
		fmt.Println("1 unread message")
	} else {
		fmt.Printf("%d unread messages\n", count)
	}
}
