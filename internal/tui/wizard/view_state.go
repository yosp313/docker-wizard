package wizard

import (
	"fmt"
	"strings"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/tui/wizard/ui"
	"docker-wizard/internal/utils"
)

func (m model) buildViewState() ui.State {
	projectName := baseName(m.root)
	if projectName == "" {
		projectName = "."
	}

	languageText := "language: detecting"
	if m.langDetected {
		languageText = "language: " + languageLabelWithVersion(m.effectiveDetails())
	}

	s := ui.State{
		Width:            m.width,
		Height:           m.height,
		Frame:            m.frame,
		Step:             mapStep(m.step),
		StepIndex:        m.stepIndex(),
		TotalSteps:       totalSteps,
		HeaderIndent:     int(4 * m.headerPos),
		LanguageText:     languageText,
		ProjectName:      projectName,
		FooterRaw:        m.footerKeys(),
		SpinnerText:      m.spinner.View(),
		DetectDone:       m.detectDone,
		DetectedLanguage: languageLabelWithVersion(m.effectiveDetails()),
		Warnings:         m.warnings,
		Blockers:         m.blockers,
		PreviewReady:     m.previewReady,
	}

	s.LanguageOptions = make([]ui.OptionItem, 0, len(m.langOptions))
	for i, option := range m.langOptions {
		s.LanguageOptions = append(s.LanguageOptions, ui.OptionItem{
			Label:       option.Label,
			Description: option.Description,
			Active:      i == m.langCursor,
			Selected:    m.isLanguageSelected(option),
		})
	}

	if m.step == stepDatabase || m.step == stepMessageQueue || m.step == stepCache || m.step == stepAnalytics || m.step == stepProxy {
		filtered := m.filteredServices(m.step)
		s.ServiceTitle = stepTitle(m.step)
		s.ServiceOptions = make([]ui.OptionItem, 0, len(filtered))
		for i, svc := range filtered {
			s.ServiceOptions = append(s.ServiceOptions, ui.OptionItem{
				Label:       svc.Label,
				Description: svc.Description,
				Active:      i == m.cursor,
				Selected:    m.selected[svc.ID],
			})
		}
	}

	groups := m.selectedByCategory()
	s.ReviewGroups = make([]ui.ReviewGroup, 0, len(utils.CategoryOrder()))
	for _, category := range utils.CategoryOrder() {
		s.ReviewGroups = append(s.ReviewGroups, ui.ReviewGroup{
			Label: utils.CategoryLabel(category),
			Items: groups[category],
		})
	}

	s.ManagedFiles = []string{"- docker-compose.yml", "- Dockerfile"}
	if m.createDockerignore {
		s.ManagedFiles = append(s.ManagedFiles, "- .dockerignore")
	}

	if m.previewReady {
		tab := m.activePreviewTab()
		s.PreviewTabs = make([]ui.PreviewTab, 0, len(m.previewTabItems()))
		for i, item := range m.previewTabItems() {
			s.PreviewTabs = append(s.PreviewTabs, ui.PreviewTab{
				Name:   item.Name,
				Short:  previewTabShortName(item.Name),
				Status: previewStatusLabel(item.File.Status),
				Active: i == m.previewTab,
			})
		}
		if tab.Name != "" {
			s.PreviewFileLine = fmt.Sprintf("file: %s    status: %s", tab.Name, previewStatusLabel(tab.File.Status))
		}

		content := m.previewContent
		if content == "" {
			content = m.activePreviewTabContent()
		}
		totalLines := lineCount(content)
		if totalLines == 0 {
			s.PreviewLineInfo = "No preview available"
		} else {
			start := m.previewViewport.YOffset + 1
			end := m.previewViewport.YOffset + m.previewViewport.Height
			if end > totalLines {
				end = totalLines
			}
			if start > end {
				start = end
			}
			s.PreviewLineInfo = fmt.Sprintf("Lines %d-%d of %d", start, end, totalLines)
		}
		s.PreviewDivider = previewDivider(m.width)
		s.PreviewBody = m.previewViewport.View()
	}

	s.ResultFiles = []string{
		outputLine(m.output.ComposePath, m.output.ComposeStatus),
		outputLine(m.output.DockerfilePath, m.output.DockerfileStatus),
		outputLine(m.output.DockerignorePath, m.output.DockerignoreStatus),
	}
	if m.output.ComposeBackupPath != "" {
		s.ResultBackups = append(s.ResultBackups, "- "+baseName(m.output.ComposeBackupPath))
	}
	if m.output.DockerfileBackupPath != "" {
		s.ResultBackups = append(s.ResultBackups, "- "+baseName(m.output.DockerfileBackupPath))
	}
	s.ResultNextSteps = []string{"- docker compose up"}

	s.ErrorMessage = "Unknown error"
	if m.err != nil {
		s.ErrorMessage = m.err.Error()
	}

	s.SideTitle = "Status"
	s.SideLines = m.sideViewLines()

	return s
}

func (m model) sideViewLines() []string {
	lines := []string{
		fmt.Sprintf("Step %d/%d", m.stepIndex(), totalSteps),
		fmt.Sprintf("Stage: %s", stepTitle(m.step)),
		"",
	}
	if m.langDetected {
		lines = append(lines, "Language: "+languageLabelWithVersion(m.effectiveDetails()))
	} else {
		lines = append(lines, "Language: detecting...")
	}
	lines = append(lines, "")
	count := len(selectedServiceIDs(m.services, m.selected))
	lines = append(lines, fmt.Sprintf("Services: %d selected", count))
	if len(m.warnings) > 0 {
		lines = append(lines, fmt.Sprintf("Warnings: %d", len(m.warnings)))
	}
	if len(m.blockers) > 0 {
		lines = append(lines, fmt.Sprintf("Blockers: %d", len(m.blockers)))
	}
	lines = append(lines, "", tipForStep(m.step))
	return lines
}

func tipForStep(s step) string {
	switch s {
	case stepWelcome:
		return "Press enter to begin"
	case stepDetect:
		return "Detecting your project..."
	case stepLanguage:
		return "Pick your language"
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		return "Space to toggle, enter to continue"
	case stepReview:
		return "Review before generating"
	case stepPreview:
		return "Scroll to inspect files"
	case stepGenerate:
		return "Generating files..."
	case stepResult:
		return "All done!"
	case stepError:
		return "Something went wrong"
	default:
		return ""
	}
}

func (m model) footerKeys() string {
	switch m.step {
	case stepWelcome:
		return "enter next | q quit"
	case stepDetect:
		if m.detectDone {
			return "enter next | l choose language | b back | q quit"
		}
		return "detecting language..."
	case stepLanguage:
		return "up/down move | enter select | b back | q quit"
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		return "up/down move | space toggle | enter next | b back | q quit"
	case stepReview:
		if len(m.blockers) > 0 {
			return "resolve blockers to continue | p preview | b back | q quit"
		}
		return "enter generate | p preview | b back | q quit"
	case stepPreview:
		if !m.previewReady {
			return "preparing preview..."
		}
		if len(m.blockers) > 0 {
			return "left/right tab | 1/2/3 file | up/down scroll | b back | q quit"
		}
		return "left/right tab | 1/2/3 file | up/down scroll | enter generate | b back | q quit"
	case stepGenerate:
		return "generating..."
	case stepResult:
		return "enter finish | q quit"
	case stepError:
		return "r retry | b back | q quit"
	default:
		return "q quit"
	}
}

func mapStep(current step) ui.Step {
	switch current {
	case stepWelcome:
		return ui.StepWelcome
	case stepDetect:
		return ui.StepDetect
	case stepLanguage:
		return ui.StepLanguage
	case stepDatabase:
		return ui.StepDatabases
	case stepMessageQueue:
		return ui.StepMessageQueue
	case stepCache:
		return ui.StepCache
	case stepAnalytics:
		return ui.StepAnalytics
	case stepProxy:
		return ui.StepProxies
	case stepReview:
		return ui.StepReview
	case stepPreview:
		return ui.StepPreview
	case stepGenerate:
		return ui.StepGenerate
	case stepResult:
		return ui.StepResult
	case stepError:
		return ui.StepError
	default:
		return ui.StepWelcome
	}
}

func previewTabShortName(name string) string {
	switch name {
	case generator.ComposeFileName:
		return "compose"
	case generator.DockerfileFileName:
		return "dockerfile"
	case generator.DockerignoreFileName:
		return "dockerignore"
	default:
		return name
	}
}

func previewStatusLabel(status generator.FileStatus) string {
	switch status {
	case generator.FileStatusNew:
		return "new"
	case generator.FileStatusSame:
		return "matches existing"
	case generator.FileStatusDifferent:
		return "differs from existing"
	case generator.FileStatusExists:
		return "exists"
	default:
		return string(status)
	}
}

func previewDivider(width int) string {
	w := ui.ContentWidth(width) - 10
	if w < 24 {
		w = 24
	}
	if w > 96 {
		w = 96
	}
	return strings.Repeat("-", w)
}
