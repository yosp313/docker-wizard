package wizard

import "docker-wizard/internal/tui/wizard/ui"

func (m model) contentAreaHeight() int {
	return ui.ContentAreaHeight(m.buildViewState())
}
