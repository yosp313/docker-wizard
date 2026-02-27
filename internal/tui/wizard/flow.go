package wizard

import "docker-wizard/internal/generator"

import tea "github.com/charmbracelet/bubbletea"

func (m *model) retryFromError() tea.Cmd {
	if m.previousStep == stepDetect {
		m.step = stepDetect
		m.animateHeader()
		return detectCmd(m.root)
	}
	if m.previousStep == stepProxy {
		if err := m.prepareReview(); err != nil {
			m.err = err
			m.step = stepError
			m.animateHeader()
			return nil
		}
		m.step = stepReview
		m.animateHeader()
		return nil
	}
	if m.previousStep == stepGenerate {
		m.step = stepGenerate
		m.animateHeader()
		return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	}
	if m.previousStep == stepPreview {
		m.step = stepPreview
		m.animateHeader()
		return previewCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	}
	return nil
}

func (m *model) animateHeader() {
	m.headerPos = 0
	m.headerVel = 0
	m.headerTarget = 0
	if m.lastStep != m.step {
		m.lastStep = m.step
	}
}

func (m *model) updateHeaderSpring() {
	if m.headerPos == 0 && m.headerVel == 0 {
		return
	}
	m.headerPos, m.headerVel = m.headerSpring.Update(m.headerPos, m.headerVel, m.headerTarget)
	if m.headerPos < 0.01 {
		m.headerPos = 0
		m.headerVel = 0
	}
}

func (m model) effectiveDetails() generator.LanguageDetails {
	if !m.overrideLang {
		return m.langDetails
	}
	updated := m.langDetails
	updated.Type = m.overrideType
	return updated
}

func (m *model) applyLanguageChoice() {
	if m.langCursor < 0 || m.langCursor >= len(m.langOptions) {
		return
	}
	choice := m.langOptions[m.langCursor]
	if choice.ID == "auto" {
		m.overrideLang = false
		m.overrideType = ""
		return
	}
	m.overrideLang = true
	m.overrideType = choice.Language
}

func (m model) isLanguageSelected(option languageChoice) bool {
	if option.ID == "auto" {
		return !m.overrideLang
	}
	if !m.overrideLang {
		return false
	}
	return m.overrideType == option.Language
}
