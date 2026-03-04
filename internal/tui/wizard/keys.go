package wizard

import (
	"docker-wizard/internal/generator"
	"docker-wizard/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" || key == "q" {
		return tea.Quit
	}

	switch m.step {
	case stepWelcome:
		return m.handleWelcomeKey(key)
	case stepDetect:
		return m.handleDetectKey(key)
	case stepLanguage:
		return m.handleLanguageKey(key)
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		return m.handleServiceStepKey(key)
	case stepReview:
		return m.handleReviewKey(key)
	case stepPreview:
		return m.handlePreviewKey(msg, key)
	case stepGenerate:
		return m.handleGenerateKey(key)
	case stepResult:
		return m.handleResultKey(key)
	case stepError:
		return m.handleErrorKey(key)
	case stepAddService:
		// handled via handleAddServiceMsg in update.go; should not reach here
	}

	return nil
}

func (m *model) handleWelcomeKey(key string) tea.Cmd {
	if key != "enter" {
		return nil
	}
	m.step = stepDetect
	m.animateHeader()
	return detectCmd(m.root)
}

func (m *model) handleDetectKey(key string) tea.Cmd {
	if !m.detectDone {
		return nil
	}

	switch key {
	case "enter":
		m.step = stepDatabase
		m.animateHeader()
	case "l":
		m.langVisited = true
		m.step = stepLanguage
		m.animateHeader()
	case "b":
		m.step = stepWelcome
		m.animateHeader()
	}

	return nil
}

func (m *model) handleLanguageKey(key string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.langCursor > 0 {
			m.langCursor--
		}
	case "down", "j":
		if m.langCursor < len(m.langOptions)-1 {
			m.langCursor++
		}
	case "enter":
		m.applyLanguageChoice()
		m.langVisited = true
		m.step = stepDatabase
		m.animateHeader()
	case "b":
		m.step = stepDetect
		m.animateHeader()
	}

	return nil
}

func (m *model) handleServiceStepKey(key string) tea.Cmd {
	services := m.filteredServices(m.step)
	m.cursor = clampCursor(m.cursor, len(services))

	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(services)-1 {
			m.cursor++
		}
	case " ":
		if len(services) > 0 {
			m.toggleCurrentSelection()
		}
	case "n":
		m.previousStep = m.step
		m.resetAddServiceForm()
		m.step = stepAddService
		m.animateHeader()
	case "enter":
		if m.step == stepProxy {
			if err := m.prepareReview(); err != nil {
				m.err = err
				m.previousStep = stepProxy
				m.step = stepError
				return nil
			}
			m.step = stepReview
			m.animateHeader()
			return nil
		}
		m.step = m.nextStep()
		m.animateHeader()
	case "b":
		m.step = m.prevStep()
		m.animateHeader()
	}

	return nil
}

func (m *model) handleReviewKey(key string) tea.Cmd {
	switch key {
	case "enter":
		if len(m.blockers) > 0 {
			return nil
		}
		m.step = stepGenerate
		m.animateHeader()
		return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	case "p":
		m.previewReady = false
		m.preview = generator.Preview{}
		m.previewTab = 0
		m.previewContent = ""
		m.setPreviewViewportContent("")
		m.step = stepPreview
		m.animateHeader()
		return previewCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	case "b":
		m.step = stepProxy
		m.animateHeader()
	}

	return nil
}

func (m *model) handleGenerateKey(_ string) tea.Cmd {
	return nil
}

func (m *model) handleResultKey(key string) tea.Cmd {
	if key == "enter" {
		return tea.Quit
	}
	return nil
}

func (m *model) handleErrorKey(key string) tea.Cmd {
	switch key {
	case "r":
		return m.retryFromError()
	case "b":
		m.step = m.previousStep
		m.animateHeader()
		return nil
	}
	return nil
}

// handleAddServiceMsg handles all messages for stepAddService, routing
// non-navigation keypresses to the active textinput.
func (m *model) handleAddServiceMsg(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if key == "ctrl+c" || key == "q" {
		return tea.Quit
	}

	switch key {
	case "esc":
		m.step = m.previousStep
		m.animateHeader()
		return nil
	case "tab":
		m.addServiceFocusedField = (m.addServiceFocusedField + 1) % addServiceFieldCount
		m.syncAddServiceFocus()
		return nil
	case "shift+tab":
		m.addServiceFocusedField = (m.addServiceFocusedField + addServiceFieldCount - 1) % addServiceFieldCount
		m.syncAddServiceFocus()
		return nil
	case "up":
		if m.addServiceFocusedField == 2 {
			categories := utils.CategoryOrder()
			m.addServiceCategoryIdx = (m.addServiceCategoryIdx - 1 + len(categories)) % len(categories)
		}
		return nil
	case "down":
		if m.addServiceFocusedField == 2 {
			categories := utils.CategoryOrder()
			m.addServiceCategoryIdx = (m.addServiceCategoryIdx + 1) % len(categories)
		}
		return nil
	case "enter":
		return m.confirmAddService()
	}

	// Forward non-special keys to the active textinput (not category field).
	if m.addServiceFocusedField != 2 {
		idx := addServiceTextInputIndex(m.addServiceFocusedField)
		if idx >= 0 {
			var cmd tea.Cmd
			m.addServiceInputs[idx], cmd = m.addServiceInputs[idx].Update(msg)
			return cmd
		}
	}

	return nil
}
