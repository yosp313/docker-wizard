package tui

import "docker-wizard/internal/tui/wizard"

type RenderMode = wizard.RenderMode

const (
	RenderModeStyled = wizard.RenderModeStyled
	RenderModePlain  = wizard.RenderModePlain
)

func SetRenderMode(mode RenderMode) {
	wizard.SetRenderMode(mode)
}

func RunWizard(root string) error {
	return wizard.Run(root)
}
