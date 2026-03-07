package wizard

import (
	"fmt"
	"os"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/tui/wizard/ui"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

func Run(root string) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		root = cwd
	}

	services, err := generator.SelectableServices(root)
	if err != nil {
		return err
	}

	m := model{
		root:             root,
		step:             stepWelcome,
		spinner:          spinner.New(),
		headerSpring:     harmonica.NewSpring(harmonica.FPS(60), 7.0, 0.6),
		services:         serviceChoicesFromCatalog(services),
		selected:         map[string]bool{},
		langOptions:      defaultLanguageOptions(),
		addServiceInputs: initAddServiceInputs(),
	}
	ui.ConfigureSpinner(&m.spinner)

	program := tea.NewProgram(m, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.spinner.Tick)
}
