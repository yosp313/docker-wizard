package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func detectCmd(root string) tea.Cmd {
	return func() tea.Msg {
		details, err := generator.DetectLanguage(root)
		return detectDoneMsg{details: details, err: err}
	}
}

func resolveLanguage(root string, overrideLang bool, overrideType generator.Language) (generator.LanguageDetails, error) {
	details, err := generator.DetectLanguage(root)
	if err != nil {
		return generator.LanguageDetails{}, err
	}
	if overrideLang {
		details.Type = overrideType
	}
	return details, nil
}

func previewCmd(root string, services []string, overrideLang bool, overrideType generator.Language) tea.Cmd {
	return func() tea.Msg {
		details, err := resolveLanguage(root, overrideLang, overrideType)
		if err != nil {
			return previewDoneMsg{err: err}
		}
		dockerfile, err := generator.Dockerfile(root, details)
		if err != nil {
			return previewDoneMsg{err: err}
		}
		compose, err := generator.Compose(root, generator.ComposeSelection{Services: services})
		if err != nil {
			return previewDoneMsg{err: err}
		}
		preview, err := generator.PreviewFiles(root, compose, dockerfile)
		return previewDoneMsg{preview: preview, err: err}
	}
}

func generateCmd(root string, services []string, overrideLang bool, overrideType generator.Language) tea.Cmd {
	return func() tea.Msg {
		details, err := resolveLanguage(root, overrideLang, overrideType)
		if err != nil {
			return generateDoneMsg{err: err}
		}
		dockerfile, err := generator.Dockerfile(root, details)
		if err != nil {
			return generateDoneMsg{err: err}
		}
		compose, err := generator.Compose(root, generator.ComposeSelection{Services: services})
		if err != nil {
			return generateDoneMsg{err: err}
		}
		output, err := generator.WriteFiles(root, compose, dockerfile)
		return generateDoneMsg{output: output, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *model) prepareReview() error {
	selection := generator.ComposeSelection{Services: selectedServiceIDs(m.services, m.selected)}
	warnings, err := generator.SelectionWarnings(m.root, selection)
	if err != nil {
		return err
	}

	details, err := resolveLanguage(m.root, m.overrideLang, m.overrideType)
	if err != nil {
		return err
	}

	dockerfile, err := generator.Dockerfile(m.root, details)
	if err != nil {
		return err
	}

	compose, err := generator.Compose(m.root, selection)
	if err != nil {
		return err
	}

	preview, err := generator.PreviewFiles(m.root, compose, dockerfile)
	if err != nil {
		return err
	}

	if preview.Compose.Status == generator.FileStatusDifferent {
		warnings = append(warnings, "docker-compose.yml differs from generated output and will be merged (backup: docker-compose.yml.bak)")
	}
	if preview.Dockerfile.Status == generator.FileStatusDifferent {
		warnings = append(warnings, "Dockerfile differs from generated output and will be merged (backup: Dockerfile.bak)")
	}

	sort.Strings(warnings)
	m.warnings = warnings
	m.blockers = nil
	m.createDockerignore = preview.Dockerignore.Status == generator.FileStatusNew
	m.previewReady = false
	m.preview = generator.Preview{}
	m.previewContent = ""
	m.setPreviewViewportContent("")

	return nil
}

func (m *model) previewContentHeight() int {
	available := m.contentAreaHeight() - 8
	if available > 18 {
		available = 18
	}
	if available < 6 {
		return 6
	}
	return available
}

func (m *model) previewContentWidth() int {
	available := contentWidth(m.width) - 6
	if available < 20 {
		return 20
	}
	return available
}

func (m *model) setPreviewViewportContent(content string) {
	if m.previewViewport.Width == 0 && m.previewViewport.Height == 0 {
		m.previewViewport = viewport.New(m.previewContentWidth(), m.previewContentHeight())
	} else {
		m.syncPreviewViewportSize()
	}
	m.previewViewport.SetContent(content)
	m.previewViewport.GotoTop()
}

func (m *model) syncPreviewViewportSize() {
	if m.previewViewport.Width == 0 && m.previewViewport.Height == 0 {
		return
	}
	m.previewViewport.Width = m.previewContentWidth()
	m.previewViewport.Height = m.previewContentHeight()
}

func lineCount(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}
func baseName(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Base(path)
}

func outputLine(path string, status generator.WriteStatus) string {
	name := baseName(path)
	if name == "" {
		name = path
	}
	return fmt.Sprintf("- %s (%s)", name, status)
}
