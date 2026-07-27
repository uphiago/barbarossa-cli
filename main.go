package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/uphiago/barbarossa-cli/internal/config"
	"github.com/uphiago/barbarossa-cli/internal/docker"
	"github.com/uphiago/barbarossa-cli/internal/tui"
)

func main() {
	cfg := config.Load()

	client, err := docker.NewClientWithHost(cfg.Docker.Host)
	if err != nil {
		fmt.Printf("%s failed to connect to Docker: %v\n", tui.Error().Render("barbarossa:"), err)
		fmt.Println("Starting with mock data instead.")
		client = nil
	}

	model := tui.NewApp(client, cfg)
	prog := tea.NewProgram(model)

	if _, err := prog.Run(); err != nil {
		fmt.Printf("%s %v\n", tui.Error().Render("barbarossa:"), err)
		os.Exit(1)
	}

	fmt.Println(lipgloss.NewStyle().Render(""))
}
