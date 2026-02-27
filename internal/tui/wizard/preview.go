package wizard

import (
	"docker-wizard/internal/generator"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handlePreviewKey(msg tea.KeyMsg, key string) tea.Cmd {
	if key == "b" {
		m.step = stepReview
		m.animateHeader()
		return nil
	}
	if !m.previewReady {
		return nil
	}

	switch key {
	case "left", "h":
		m.switchPreviewTab(-1)
		return nil
	case "right", "l":
		m.switchPreviewTab(1)
		return nil
	case "1":
		m.setPreviewTab(0)
		return nil
	case "2":
		m.setPreviewTab(1)
		return nil
	case "3":
		m.setPreviewTab(2)
		return nil
	case "home":
		m.previewViewport.GotoTop()
		return nil
	case "end":
		m.previewViewport.GotoBottom()
		return nil
	case "enter":
		if len(m.blockers) > 0 {
			return nil
		}
		m.step = stepGenerate
		m.animateHeader()
		return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	}

	var cmd tea.Cmd
	m.previewViewport, cmd = m.previewViewport.Update(msg)
	return cmd
}

type previewTabItem struct {
	Name string
	File generator.FilePreview
}

func (m model) previewTabItems() []previewTabItem {
	return []previewTabItem{
		{Name: generator.ComposeFileName, File: m.preview.Compose},
		{Name: generator.DockerfileFileName, File: m.preview.Dockerfile},
		{Name: generator.DockerignoreFileName, File: m.preview.Dockerignore},
	}
}

func (m model) activePreviewTab() previewTabItem {
	items := m.previewTabItems()
	if len(items) == 0 {
		return previewTabItem{}
	}
	index := m.previewTab
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}
	return items[index]
}

func (m model) activePreviewTabContent() string {
	tab := m.activePreviewTab()
	if tab.Name == "" {
		return ""
	}
	if tab.File.Status == generator.FileStatusExists {
		return "existing file will be kept"
	}
	if tab.File.Content == "" {
		return "no preview available"
	}
	return tab.File.Content
}

func (m *model) refreshPreviewTabContent() {
	m.previewContent = m.activePreviewTabContent()
	m.setPreviewViewportContent(m.previewContent)
}

func (m *model) setPreviewTab(index int) {
	items := m.previewTabItems()
	if len(items) == 0 {
		m.previewTab = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}
	if m.previewTab == index {
		return
	}
	m.previewTab = index
	if m.previewReady {
		m.refreshPreviewTabContent()
	}
}

func (m *model) switchPreviewTab(delta int) {
	items := m.previewTabItems()
	if len(items) == 0 || delta == 0 {
		return
	}
	m.setPreviewTab(m.previewTab + delta)
}
