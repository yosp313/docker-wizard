package wizard

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncPreviewViewportSize()
		return m, nil
	case tickMsg:
		m.frame++
		m.updateHeaderSpring()
		return m, tickCmd()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case detectDoneMsg:
		m.langDetected = msg.err == nil
		m.langDetails = msg.details
		if msg.err != nil {
			m.err = msg.err
			m.previousStep = stepDetect
			m.step = stepError
			return m, nil
		}
		m.langOptions = languageOptionsForDetected(m.langDetails)
		m.detectDone = true
		m.step = stepDetect
		m.animateHeader()
		return m, nil
	case generateDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.previousStep = stepGenerate
			m.step = stepError
			return m, nil
		}
		m.output = msg.output
		m.step = stepResult
		m.animateHeader()
		return m, nil
	case previewDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.previousStep = stepPreview
			m.step = stepError
			return m, nil
		}
		m.preview = msg.preview
		m.previewReady = true
		if m.previewTab < 0 || m.previewTab >= len(m.previewTabItems()) {
			m.previewTab = 0
		}
		m.refreshPreviewTabContent()
		m.step = stepPreview
		m.animateHeader()
		return m, nil
	case tea.KeyMsg:
		if m.step == stepAddService {
			return m, m.handleAddServiceMsg(msg)
		}
		return m, m.handleKey(msg)
	}

	return m, nil
}
