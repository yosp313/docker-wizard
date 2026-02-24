package tui

import (
	"fmt"
	"strings"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	content := ""
	switch m.step {
	case stepWelcome:
		content = m.viewWelcome()
	case stepDetect:
		content = m.viewDetect()
	case stepLanguage:
		content = m.viewLanguage()
	case stepDatabase:
		content = m.viewServices(stepDatabase)
	case stepMessageQueue:
		content = m.viewServices(stepMessageQueue)
	case stepCache:
		content = m.viewServices(stepCache)
	case stepAnalytics:
		content = m.viewServices(stepAnalytics)
	case stepProxy:
		content = m.viewServices(stepProxy)
	case stepReview:
		content = m.viewReview()
	case stepPreview:
		content = m.viewPreview()
	case stepGenerate:
		content = m.viewGenerate()
	case stepResult:
		content = m.viewResult()
	case stepError:
		content = m.viewError()
	default:
		content = ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		content,
		m.renderFooter(),
	)
}

func (m model) renderHeader() string {
	indent := int(4 * m.headerPos)
	padding := strings.Repeat(" ", indent)

	stepText := fmt.Sprintf("Step %d/%d", m.stepIndex(), totalSteps)
	langText := "language: detecting"
	if m.langDetected {
		langText = "language: " + languageLabelWithVersion(m.effectiveDetails())
	}
	progress := progressBar(m.stepIndex(), totalSteps, 22)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(titleColor(m.frame))
	art := titleArt()
	for i := range art {
		art[i] = titleStyle.Render(art[i])
	}
	artBlock := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(strings.Join(art, "\n"))

	stepBadge := badgeStyle().Render(stepText)
	langBadge := badgeStyle().Render(langText)
	progressBadge := badgeStyle().Render(progress)

	bar := lipgloss.JoinHorizontal(lipgloss.Center, stepBadge, "  ", progressBadge, "  ", langBadge)
	bar = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(bar)

	content := lipgloss.JoinVertical(lipgloss.Center, artBlock, "", bar)
	box := headerStyle(m.width).Render(padding + content)
	return box
}

func (m model) renderFooter() string {
	keys := m.footerKeys()
	return footerStyle(m.width).Render(keys)
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
			return "up/down scroll | b back | q quit"
		}
		return "up/down scroll | enter generate | b back | q quit"
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

func (m model) viewWelcome() string {
	body := []string{
		"Welcome to docker-wizard.",
		"This wizard will detect your project language,",
		"generate a Dockerfile, and create a docker-compose.yml",
		"with the services you choose.",
	}
	return cardStyle(m.width).Render(sectionTitle("Welcome") + "\n\n" + strings.Join(body, "\n"))
}

func (m model) viewDetect() string {
	if !m.detectDone {
		text := "Detecting project language"
		line := fmt.Sprintf("%s %s", m.spinner.View(), text)
		return cardStyle(m.width).Render(sectionTitle("Detect") + "\n\n" + line)
	}

	body := []string{
		"Detected language:",
		languageLabelWithVersion(m.effectiveDetails()),
		"",
		"Press l to choose a different language.",
	}
	return cardStyle(m.width).Render(sectionTitle("Detect") + "\n\n" + strings.Join(body, "\n"))
}

func (m model) viewLanguage() string {
	items := make([]string, 0, len(m.langOptions))
	for i, option := range m.langOptions {
		cursor := " "
		if i == m.langCursor {
			cursor = ">"
		}
		selected := m.isLanguageSelected(option)
		check := "[ ]"
		if selected {
			check = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, check, option.Label)
		if option.Description != "" {
			line += "  " + mutedStyle().Render(option.Description)
		}
		items = append(items, serviceLineStyle(i == m.langCursor, selected).Render(line))
	}
	content := strings.Join(items, "\n")
	return cardStyle(m.width).Render(sectionTitle("Language") + "\n\n" + content)
}

func (m model) viewServices(current step) string {
	filtered := m.filteredServices(current)
	items := make([]string, 0, len(filtered))
	for i, svc := range filtered {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		check := "[ ]"
		if m.selected[svc.ID] {
			check = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, check, svc.Label)
		if svc.Description != "" {
			line += "  " + mutedStyle().Render(svc.Description)
		}
		items = append(items, serviceLineStyle(i == m.cursor, m.selected[svc.ID]).Render(line))
	}
	content := strings.Join(items, "\n")
	return cardStyle(m.width).Render(sectionTitle(stepTitle(current)) + "\n\n" + content)
}

func (m model) viewReview() string {
	groups := m.selectedByCategory()
	body := []string{
		"Review your selections:",
		"",
		fmt.Sprintf("Detected language: %s", languageLabelWithVersion(m.effectiveDetails())),
		"",
	}

	for _, category := range categoryOrder() {
		labels := groups[category]
		if len(labels) == 0 {
			labels = []string{"none"}
		}
		body = append(body, fmt.Sprintf("%s:", categoryLabel(category)))
		body = append(body, "- "+strings.Join(labels, "\n- "))
		body = append(body, "")
	}

	files := []string{"- docker-compose.yml", "- Dockerfile"}
	if m.createDockerignore {
		files = append(files, "- .dockerignore")
	}
	body = append(body,
		"Managed files:",
		strings.Join(files, "\n"),
		"Existing differing files are merged and backed up as *.bak.",
	)

	if len(m.blockers) > 0 {
		blockerBlock := []string{
			"",
			blockerTitle().Render("Blocking issues"),
			"- " + strings.Join(m.blockers, "\n- "),
		}
		body = append(body, blockerBlock...)
	}
	if len(m.warnings) > 0 {
		warningBlock := []string{
			"",
			warningTitle().Render("Warnings"),
			"- " + strings.Join(m.warnings, "\n- "),
		}
		body = append(body, warningBlock...)
	}
	return cardStyle(m.width).Render(sectionTitle("Review") + "\n\n" + strings.Join(body, "\n"))
}

func (m model) viewPreview() string {
	if !m.previewReady {
		line := fmt.Sprintf("%s Preparing preview", m.spinner.View())
		return cardStyle(m.width).Render(sectionTitle("Preview") + "\n\n" + line)
	}

	content := m.previewContent
	if content == "" {
		content = strings.Join(buildPreviewLines(m.preview, m.blockers), "\n")
	}

	totalLines := lineCount(content)
	lineInfo := ""
	if totalLines == 0 {
		lineInfo = "No preview available"
	} else {
		start := m.previewViewport.YOffset + 1
		end := m.previewViewport.YOffset + m.previewViewport.Height
		if end > totalLines {
			end = totalLines
		}
		if start > end {
			start = end
		}
		lineInfo = fmt.Sprintf("Lines %d-%d of %d", start, end, totalLines)
	}
	body := []string{
		mutedStyle().Render(lineInfo),
		"",
		m.previewViewport.View(),
	}
	return cardStyle(m.width).Render(sectionTitle("Preview") + "\n\n" + strings.Join(body, "\n"))
}

func (m model) viewGenerate() string {
	line := fmt.Sprintf("%s Generating docker-compose.yml and Dockerfile", m.spinner.View())
	return cardStyle(m.width).Render(sectionTitle("Generate") + "\n\n" + line)
}

func (m model) viewResult() string {
	files := []string{
		outputLine(m.output.ComposePath, m.output.ComposeStatus),
		outputLine(m.output.DockerfilePath, m.output.DockerfileStatus),
		outputLine(m.output.DockerignorePath, m.output.DockerignoreStatus),
	}
	body := []string{
		successTitle().Render("All set."),
		"",
		"Write results:",
		strings.Join(files, "\n"),
	}

	backups := []string{}
	if m.output.ComposeBackupPath != "" {
		backups = append(backups, "- "+baseName(m.output.ComposeBackupPath))
	}
	if m.output.DockerfileBackupPath != "" {
		backups = append(backups, "- "+baseName(m.output.DockerfileBackupPath))
	}
	if len(backups) > 0 {
		body = append(body,
			"",
			"Backups:",
			strings.Join(backups, "\n"),
		)
	}

	body = append(body,
		"",
		"Next steps:",
		"- docker compose up",
	)

	return cardStyle(m.width).Render(sectionTitle("Result") + "\n\n" + strings.Join(body, "\n"))
}

func (m model) viewError() string {
	message := "Unknown error"
	if m.err != nil {
		message = m.err.Error()
	}
	body := []string{
		"Something went wrong.",
		"",
		message,
	}
	return errorStyle(m.width).Render(sectionTitle("Error") + "\n\n" + strings.Join(body, "\n"))
}

func buildPreviewLines(preview generator.Preview, blockers []string) []string {
	lines := []string{}
	lines = appendPreviewBlock(lines, generator.ComposeFileName, preview.Compose, true)
	lines = append(lines, "")
	lines = appendPreviewBlock(lines, generator.DockerfileFileName, preview.Dockerfile, true)
	lines = append(lines, "")
	showDockerignore := preview.Dockerignore.Status != generator.FileStatusExists
	lines = appendPreviewBlock(lines, generator.DockerignoreFileName, preview.Dockerignore, showDockerignore)
	if len(blockers) > 0 {
		lines = append(lines, "", "Blocking issues:")
		for _, blocker := range blockers {
			lines = append(lines, "- "+blocker)
		}
	}
	return lines
}

func appendPreviewBlock(lines []string, name string, preview generator.FilePreview, showContent bool) []string {
	status := previewStatusLabel(preview.Status)
	lines = append(lines, fmt.Sprintf("%s (%s)", name, status))
	if showContent && preview.Content != "" {
		lines = append(lines, "  ---")
		lines = append(lines, indentLines(preview.Content, "  ")...)
		return lines
	}
	if preview.Status == generator.FileStatusExists {
		lines = append(lines, "  existing file will be kept")
	}
	return lines
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

func indentLines(value string, indent string) []string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = indent + lines[i]
	}
	return lines
}
