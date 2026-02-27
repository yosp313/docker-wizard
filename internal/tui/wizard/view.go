package wizard

import "docker-wizard/internal/tui/wizard/ui"

func (m model) View() string {
	return ui.Render(m.buildViewState())
}
