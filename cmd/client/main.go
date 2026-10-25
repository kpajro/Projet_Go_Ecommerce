package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pajro/pgo/internal/config"
	"github.com/pajro/pgo/internal/tui/widgets"
)

func main() {
	cfg := config.Load()

	program := tea.NewProgram(
		widgets.NewBoot("E-Go | INTERFACE client", cfg.APIURL),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}
