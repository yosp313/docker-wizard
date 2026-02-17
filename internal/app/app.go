package app

import (
	"fmt"
	"os"

	"docker-wizard/internal/tui"
)

func Run() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	return tui.RunWizard(root)
}
