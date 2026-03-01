package app

import (
	"fmt"
	"os"
	"strings"

	cliwizard "docker-wizard/internal/cli"
	"docker-wizard/internal/tui"
)

type Mode string

const (
	ModeStyled Mode = "styled"
	ModePlain  Mode = "plain"
	ModeCLI    Mode = "cli"
	ModeBatch  Mode = "batch"
)

type AutomationOptions struct {
	Services []string
	Language string
	Write    bool
	DryRun   bool
}

type Options struct {
	Mode       Mode
	Automation AutomationOptions
}

func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case "", ModeStyled:
		return ModeStyled, nil
	case ModePlain:
		return ModePlain, nil
	case ModeCLI:
		return ModeCLI, nil
	case ModeBatch:
		return ModeBatch, nil
	default:
		return "", fmt.Errorf("invalid mode %q (expected styled, plain, cli, or batch)", value)
	}
}

func Run() error {
	return RunWithOptions(Options{Mode: ModeStyled})
}

func RunWithOptions(options Options) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	mode := options.Mode
	if mode == "" {
		mode = ModeStyled
	}

	switch mode {
	case ModeStyled:
		tui.SetRenderMode(tui.RenderModeStyled)
		return tui.RunWizard(root)
	case ModePlain:
		tui.SetRenderMode(tui.RenderModePlain)
		return tui.RunWizard(root)
	case ModeCLI:
		return cliwizard.RunInteractive(root)
	case ModeBatch:
		return cliwizard.RunNonInteractive(root, cliwizard.NonInteractiveOptions{
			Services: options.Automation.Services,
			Language: options.Automation.Language,
			Write:    options.Automation.Write,
			DryRun:   options.Automation.DryRun,
		})
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
}

type AddOptions = cliwizard.AddOptions

func RunAdd(options AddOptions) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return cliwizard.RunAdd(root, options)
}

func RunList() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return cliwizard.RunList(root)
}
