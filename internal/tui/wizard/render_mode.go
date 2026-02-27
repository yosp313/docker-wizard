package wizard

import "docker-wizard/internal/tui/wizard/ui"

type RenderMode = ui.RenderMode

const (
	RenderModeStyled = ui.RenderModeStyled
	RenderModePlain  = ui.RenderModePlain
)

func SetRenderMode(mode RenderMode) {
	ui.SetRenderMode(mode)
}
