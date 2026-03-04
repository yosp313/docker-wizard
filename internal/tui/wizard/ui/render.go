package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Render(s State) string {
	content := renderContent(s)

	if isPlainMode() {
		return strings.Join([]string{renderHeader(s), content, renderFooter(s)}, "\n")
	}

	if shouldShowSidePanel(s.Width) {
		cs := s
		cs.Width = mainPanelWidth(s.Width)
		content = renderContent(cs)
		mw := mainPanelWidth(s.Width)
		content = lipgloss.NewStyle().Width(mw).Align(lipgloss.Center).Render(content)
		sidePanel := renderSidePanel(s)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, " ", sidePanel)
	} else {
		content = lipgloss.NewStyle().Width(s.Width).Align(lipgloss.Center).Render(content)
	}

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		renderHeader(s),
		content,
		renderFooter(s),
	)
	if s.Width > 0 && s.Height > 0 {
		return lipgloss.NewStyle().
			Width(s.Width).
			Height(s.Height).
			Background(palettePanel).
			Foreground(paletteText).
			Render(view)
	}
	return view
}

func renderContent(s State) string {
	switch s.Step {
	case StepWelcome:
		return viewWelcome(s)
	case StepDetect:
		return viewDetect(s)
	case StepLanguage:
		return viewLanguage(s)
	case StepDatabases, StepMessageQueue, StepCache, StepAnalytics, StepProxies:
		return viewServices(s)
	case StepAddService:
		return viewAddService(s)
	case StepReview:
		return viewReview(s)
	case StepPreview:
		return viewPreview(s)
	case StepGenerate:
		return viewGenerate(s)
	case StepResult:
		return viewResult(s)
	case StepError:
		return viewError(s)
	default:
		return ""
	}
}

func renderSidePanel(s State) string {
	headerH := lipgloss.Height(renderHeader(s))
	footerH := lipgloss.Height(renderFooter(s))
	available := s.Height - headerH - footerH - 4
	if available < 4 {
		available = 4
	}

	lines := []string{sectionTitle(s.SideTitle), ""}
	for _, line := range s.SideLines {
		if line == "" {
			lines = append(lines, "")
		} else {
			lines = append(lines, mutedStyle().Render(line))
		}
	}
	body := strings.Join(lines, "\n")
	spw := sidePanelWidth(s.Width) - 4
	if spw < 10 {
		spw = 10
	}
	return sidePanelStyle(spw, available).Render(body)
}

func renderHeader(s State) string {
	padding := strings.Repeat(" ", s.HeaderIndent)
	if isPlainMode() {
		return plainHeader(s.StepIndex, s.TotalSteps, string(s.Step), s.LanguageText, s.ProjectName, progressBar(s.StepIndex, s.TotalSteps), s.HeaderIndent)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(titleColor(s.Frame)).Render("Docker Wizard")
	subtitle := mutedStyle().Render("Generate Dockerfile and docker-compose from guided selections")
	project := mutedStyle().Render("project: " + s.ProjectName)

	bar := progressBar(s.StepIndex, s.TotalSteps)
	bar = lipgloss.NewStyle().Width(ContentWidth(s.Width)).Align(lipgloss.Center).Render(bar)

	heading := lipgloss.JoinVertical(lipgloss.Center, title, subtitle, project)
	heading = lipgloss.NewStyle().Width(ContentWidth(s.Width)).Align(lipgloss.Center).Render(heading)

	content := lipgloss.JoinVertical(lipgloss.Center, heading, "", bar)
	content = lipgloss.NewStyle().Width(s.Width).Align(lipgloss.Center).Render(content)
	return headerStyle(s.Width).Render(padding + content)
}

func renderFooter(s State) string {
	hints := formatFooterHints(s.FooterRaw)
	if isPlainMode() {
		return hints
	}
	innerWidth := s.Width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}
	return footerStyle(s.Width).Render(lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(hints))
}

func viewWelcome(s State) string {
	body := []string{
		"Welcome to docker-wizard.",
		"This wizard will detect your project language,",
		"generate a Dockerfile, and create a docker-compose.yml",
		"with the services you choose.",
	}
	return renderCard(s.Width, "Welcome", strings.Join(body, "\n"))
}

func viewDetect(s State) string {
	if !s.DetectDone {
		line := fmt.Sprintf("%s Detecting project language", s.SpinnerText)
		return renderCard(s.Width, "Detect", line)
	}
	body := []string{
		"Detected language:",
		s.DetectedLanguage,
		"",
		"Press l to choose a different language.",
	}
	return renderCard(s.Width, "Detect", strings.Join(body, "\n"))
}

func viewLanguage(s State) string {
	items := make([]string, 0, len(s.LanguageOptions))
	for _, option := range s.LanguageOptions {
		cursor := " "
		if option.Active {
			cursor = ">"
		}
		check := "[ ]"
		if option.Selected {
			check = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, check, option.Label)
		if option.Description != "" {
			line += " - " + option.Description
		}
		items = append(items, serviceLineStyle(option.Active, option.Selected).Render(line))
	}
	return renderCard(s.Width, "Language", strings.Join(items, "\n"))
}

func viewServices(s State) string {
	items := make([]string, 0, len(s.ServiceOptions))
	for _, option := range s.ServiceOptions {
		cursor := " "
		if option.Active {
			cursor = ">"
		}
		check := "[ ]"
		if option.Selected {
			check = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, check, option.Label)
		if option.Description != "" {
			line += " - " + option.Description
		}
		items = append(items, serviceLineStyle(option.Active, option.Selected).Render(line))
	}
	return renderCard(s.Width, s.ServiceTitle, strings.Join(items, "\n"))
}

func viewReview(s State) string {
	body := []string{
		"Review your selections:",
		"",
		fmt.Sprintf("Detected language: %s", s.DetectedLanguage),
		"",
	}
	for _, group := range s.ReviewGroups {
		items := group.Items
		if len(items) == 0 {
			items = []string{"none"}
		}
		body = append(body, group.Label+":")
		body = append(body, "- "+strings.Join(items, "\n- "))
		body = append(body, "")
	}
	body = append(body,
		"Managed files:",
		strings.Join(s.ManagedFiles, "\n"),
		"Existing differing files are merged and backed up as *.bak.",
	)
	if len(s.Blockers) > 0 {
		body = append(body, "", blockerTitle().Render("Blocking issues"), "- "+strings.Join(s.Blockers, "\n- "))
	}
	if len(s.Warnings) > 0 {
		body = append(body, "", warningTitle().Render("Warnings"), "- "+strings.Join(s.Warnings, "\n- "))
	}
	return renderCard(s.Width, "Review", strings.Join(body, "\n"))
}

func viewPreview(s State) string {
	if !s.PreviewReady {
		line := fmt.Sprintf("%s Preparing preview", s.SpinnerText)
		return renderCard(s.Width, "Preview", line)
	}
	body := []string{renderPreviewTabs(s.PreviewTabs), ""}
	body = append(body, mutedStyle().Render("tab: left/right or 1/2/3"), "")
	if s.PreviewFileLine != "" {
		body = append(body, s.PreviewFileLine)
	}
	body = append(body, mutedStyle().Render(s.PreviewLineInfo), mutedStyle().Render(s.PreviewDivider), "")
	if len(s.Blockers) > 0 {
		body = append(body, blockerTitle().Render("Blocking issues"))
		for _, blocker := range s.Blockers {
			body = append(body, "- "+blocker)
		}
		body = append(body, "")
	}
	body = append(body, s.PreviewBody)
	return renderCompactCard(s.Width, "Preview", strings.Join(body, "\n"))
}

func renderPreviewTabs(tabs []PreviewTab) string {
	parts := make([]string, 0, len(tabs))
	for i, tab := range tabs {
		label := fmt.Sprintf("%d %s", i+1, tab.Short)
		if tab.Active {
			parts = append(parts, activePreviewTabStyle().Render("[> "+label+"]"))
		} else {
			parts = append(parts, inactivePreviewTabStyle().Render("[  "+label+"]"))
		}
	}
	return strings.Join(parts, "  ")
}

func viewGenerate(s State) string {
	line := fmt.Sprintf("%s Generating docker-compose.yml and Dockerfile", s.SpinnerText)
	return renderCard(s.Width, "Generate", line)
}

func viewResult(s State) string {
	body := []string{
		successTitle().Render("All set."),
		"",
		"Write results:",
		strings.Join(s.ResultFiles, "\n"),
	}
	if len(s.ResultBackups) > 0 {
		body = append(body, "", "Backups:", strings.Join(s.ResultBackups, "\n"))
	}
	if len(s.ResultNextSteps) > 0 {
		body = append(body, "", "Next steps:", strings.Join(s.ResultNextSteps, "\n"))
	}
	return renderCard(s.Width, "Result", strings.Join(body, "\n"))
}

func viewError(s State) string {
	body := []string{"Something went wrong.", "", s.ErrorMessage}
	return renderErrorCard(s.Width, "Error", strings.Join(body, "\n"))
}

func viewAddService(s State) string {
	return renderCard(s.Width, "Add Custom Service", s.AddServiceBody)
}

func renderCard(width int, title string, body string) string {
	return cardStyle(width).Render(sectionTitle(title) + "\n\n" + body)
}

func renderErrorCard(width int, title string, body string) string {
	return errorStyle(width).Render(sectionTitle(title) + "\n\n" + body)
}

func renderCompactCard(width int, title string, body string) string {
	return cardStyle(width).Render(sectionTitle(title) + "\n" + body)
}

func ContentAreaHeight(s State) int {
	if s.Height <= 0 {
		return 0
	}
	headerHeight := lipgloss.Height(renderHeader(s))
	footerHeight := lipgloss.Height(renderFooter(s))
	available := s.Height - headerHeight - footerHeight
	if available < 1 {
		return 1
	}
	return available
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
	case "enter", "q", "b", "l", "n", "p", "r", "space", "up/down", "left/right", "1/2/3", "home", "end", "tab", "shift+tab", "esc":
		return true
	default:
		return false
	}
}
