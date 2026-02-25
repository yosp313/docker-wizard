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
)

type Options struct {
	Mode Mode
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
	default:
		return "", fmt.Errorf("invalid mode %q (expected styled, plain, or cli)", value)
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
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
}
