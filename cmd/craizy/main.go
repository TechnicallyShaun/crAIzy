package main

import (
	"flag"
	"fmt"
	"os"

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

	p := tea.NewProgram(tui.NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
