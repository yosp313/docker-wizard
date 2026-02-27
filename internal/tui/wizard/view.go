package wizard

import (
	"fmt"
	"strings"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/utils"

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
	if isPlainMode() {
		return strings.Join([]string{m.renderHeader(), content, m.renderFooter()}, "\n")
	}
	content = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(content)
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		content,
		m.renderFooter(),
	)
	if m.width > 0 && m.height > 0 {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Background(palettePanel).
			Foreground(paletteText).
			Render(view)
	}
	return view
}

func (m model) renderHeader() string {
	indent := int(4 * m.headerPos)
	padding := strings.Repeat(" ", indent)

	stepText := fmt.Sprintf("Step %d/%d", m.stepIndex(), totalSteps)
	stepName := stepName(m.step)
	langText := "language: detecting"
	if m.langDetected {
		langText = "language: " + languageLabelWithVersion(m.effectiveDetails())
	}
	projectName := baseName(m.root)
	if projectName == "" {
		projectName = "."
	}
	if isPlainMode() {
		return plainHeader(m.stepIndex(), totalSteps, stepName, langText, projectName, progressChips(m.stepIndex(), totalSteps), indent)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(titleColor(m.frame)).Render("Docker Wizard")
	subtitle := mutedStyle().Render("Generate Dockerfile and docker-compose from guided selections")
	project := mutedStyle().Render("project: " + projectName)

	stepBadge := badgeStyle().Render(stepText)
	stepNameBadge := badgeStyle().Render(stepName)
	langBadge := badgeStyle().Render(langText)
	status := lipgloss.JoinHorizontal(lipgloss.Left, stepBadge, "  ", stepNameBadge, "  ", langBadge)
	status = lipgloss.NewStyle().Width(contentWidth(m.width)).Align(lipgloss.Center).Render(status)

	chips := progressChips(m.stepIndex(), totalSteps)
	chips = lipgloss.NewStyle().Width(contentWidth(m.width)).Align(lipgloss.Center).Render(chips)

	heading := lipgloss.JoinVertical(lipgloss.Center, title, subtitle, project)
	heading = lipgloss.NewStyle().Width(contentWidth(m.width)).Align(lipgloss.Center).Render(heading)

	content := lipgloss.JoinVertical(lipgloss.Center, heading, "", status, chips)
	content = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(content)
	box := headerStyle(m.width).Render(padding + content)
	return box
}

func (m model) renderFooter() string {
	hints := formatFooterHints(m.footerKeys())
	if isPlainMode() {
		return hints
	}
	innerWidth := m.width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}
	return footerStyle(m.width).Render(lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hints))
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

func (m model) viewWelcome() string {
	body := []string{
		"Welcome to docker-wizard.",
		"This wizard will detect your project language,",
		"generate a Dockerfile, and create a docker-compose.yml",
		"with the services you choose.",
	}
	return m.renderCard("Welcome", strings.Join(body, "\n"))
}

func (m model) viewDetect() string {
	if !m.detectDone {
		text := "Detecting project language"
		line := fmt.Sprintf("%s %s", m.spinner.View(), text)
		return m.renderCard("Detect", line)
	}

	body := []string{
		"Detected language:",
		languageLabelWithVersion(m.effectiveDetails()),
		"",
		"Press l to choose a different language.",
	}
	return m.renderCard("Detect", strings.Join(body, "\n"))
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
			line += " - " + option.Description
		}
		rendered := serviceLineStyle(i == m.langCursor, selected).Render(line)
		items = append(items, rendered)
	}
	content := strings.Join(items, "\n")
	return m.renderCard("Language", content)
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
			line += " - " + svc.Description
		}
		rendered := serviceLineStyle(i == m.cursor, m.selected[svc.ID]).Render(line)
		items = append(items, rendered)
	}
	content := strings.Join(items, "\n")
	return m.renderCard(stepTitle(current), content)
}

func formatFooterHints(raw string) string {
	parts := strings.Split(raw, "|")
	formatted := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) >= 2 && isKeyToken(fields[0]) {
			if isPlainMode() {
				formatted = append(formatted, "["+fields[0]+"] "+strings.Join(fields[1:], " "))
				continue
			}
			key := keycapStyle().Render(fields[0])
			label := mutedStyle().Render(strings.Join(fields[1:], " "))
			formatted = append(formatted, lipgloss.JoinHorizontal(lipgloss.Left, key, " ", label))
			continue
		}
		if isPlainMode() {
			formatted = append(formatted, part)
		} else {
			formatted = append(formatted, mutedStyle().Render(part))
		}
	}
	if isPlainMode() {
		return strings.Join(formatted, " | ")
	}
	return strings.Join(formatted, "   ")
}

func isKeyToken(token string) bool {
	switch token {
	case "enter", "q", "b", "l", "p", "r", "space", "up/down", "left/right", "1/2/3", "home", "end":
		return true
	default:
		return false
	}
}

func stepName(current step) string {
	switch current {
	case stepWelcome:
		return "welcome"
	case stepDetect:
		return "detect"
	case stepLanguage:
		return "language"
	case stepDatabase:
		return "databases"
	case stepMessageQueue:
		return "message queues"
	case stepCache:
		return "cache"
	case stepAnalytics:
		return "analytics"
	case stepProxy:
		return "proxies"
	case stepReview:
		return "review"
	case stepPreview:
		return "preview"
	case stepGenerate:
		return "generate"
	case stepResult:
		return "result"
	case stepError:
		return "error"
	default:
		return "wizard"
	}
}

func (m model) viewReview() string {
	groups := m.selectedByCategory()
	body := []string{
		"Review your selections:",
		"",
		fmt.Sprintf("Detected language: %s", languageLabelWithVersion(m.effectiveDetails())),
		"",
	}

	for _, category := range utils.CategoryOrder() {
		labels := groups[category]
		if len(labels) == 0 {
			labels = []string{"none"}
		}
		body = append(body, fmt.Sprintf("%s:", utils.CategoryLabel(category)))
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
	return m.renderCard("Review", strings.Join(body, "\n"))
}

func (m model) viewPreview() string {
	if !m.previewReady {
		line := fmt.Sprintf("%s Preparing preview", m.spinner.View())
		return m.renderCard("Preview", line)
	}

	tab := m.activePreviewTab()
	content := m.previewContent
	if content == "" {
		content = m.activePreviewTabContent()
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

	body := []string{m.renderPreviewTabs(), ""}
	body = append(body, mutedStyle().Render("tab: left/right or 1/2/3"), "")
	if tab.Name != "" {
		body = append(body, fmt.Sprintf("file: %s    status: %s", tab.Name, previewStatusLabel(tab.File.Status)))
	}
	body = append(body, mutedStyle().Render(lineInfo), mutedStyle().Render(previewDivider(m.width)), "")
	if len(m.blockers) > 0 {
		body = append(body, blockerTitle().Render("Blocking issues"))
		for _, blocker := range m.blockers {
			body = append(body, "- "+blocker)
		}
		body = append(body, "")
	}
	body = append(body, m.previewViewport.View())
	return m.renderCompactCard("Preview", strings.Join(body, "\n"))
}

func (m model) renderPreviewTabs() string {
	items := m.previewTabItems()
	parts := make([]string, 0, len(items))
	for i, tab := range items {
		label := fmt.Sprintf("%d %s", i+1, previewTabShortName(tab.Name))
		if i == m.previewTab {
			parts = append(parts, activePreviewTabStyle().Render("[> "+label+"]"))
		} else {
			parts = append(parts, inactivePreviewTabStyle().Render("[  "+label+"]"))
		}
	}
	return strings.Join(parts, "  ")
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

func previewDivider(width int) string {
	w := contentWidth(width) - 10
	if w < 24 {
		w = 24
	}
	if w > 96 {
		w = 96
	}
	return strings.Repeat("-", w)
}

func (m model) viewGenerate() string {
	line := fmt.Sprintf("%s Generating docker-compose.yml and Dockerfile", m.spinner.View())
	return m.renderCard("Generate", line)
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

	return m.renderCard("Result", strings.Join(body, "\n"))
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
	return m.renderErrorCard("Error", strings.Join(body, "\n"))
}

func (m model) contentAreaHeight() int {
	if m.height <= 0 {
		return 0
	}
	headerHeight := lipgloss.Height(m.renderHeader())
	footerHeight := lipgloss.Height(m.renderFooter())
	available := m.height - headerHeight - footerHeight
	if available < 1 {
		return 1
	}
	return available
}

func (m model) renderCard(title string, body string) string {
	style := cardStyle(m.width)
	return style.Render(sectionTitle(title) + "\n\n" + body)
}

func (m model) renderErrorCard(title string, body string) string {
	style := errorStyle(m.width)
	return style.Render(sectionTitle(title) + "\n\n" + body)
}

func (m model) renderCompactCard(title string, body string) string {
	style := cardStyle(m.width)
	return style.Render(sectionTitle(title) + "\n" + body)
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
