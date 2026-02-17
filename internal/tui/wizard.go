package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

type step int

const (
	stepWelcome step = iota
	stepDetect
	stepDatabase
	stepMessageQueue
	stepCache
	stepAnalytics
	stepProxy
	stepReview
	stepGenerate
	stepResult
	stepError
)

type tickMsg time.Time

type detectDoneMsg struct {
	details generator.LanguageDetails
	err     error
}

type generateDoneMsg struct {
	output generator.Output
	err    error
}

type serviceChoice struct {
	ID          string
	Label       string
	Selected    bool
	Description string
	Category    string
}

type model struct {
	root string

	step         step
	previousStep step

	width  int
	height int

	spinner spinner.Model

	langDetected bool
	langDetails  generator.LanguageDetails

	services []serviceChoice
	cursor   int
	selected map[string]bool
	warnings []string
	frame    int

	output generator.Output
	err    error

	headerSpring harmonica.Spring
	headerPos    float64
	headerVel    float64
	headerTarget float64
	lastStep     step
}

func RunWizard(root string) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		root = cwd
	}

	services, err := generator.SelectableServices(root)
	if err != nil {
		return err
	}

	m := model{
		root:         root,
		step:         stepWelcome,
		spinner:      spinner.New(),
		headerSpring: harmonica.NewSpring(harmonica.FPS(60), 7.0, 0.6),
		services:     serviceChoicesFromCatalog(services),
		selected:     map[string]bool{},
	}
	setSpinner(&m.spinner)

	program := tea.NewProgram(m, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		m.step = stepDatabase
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
	case tea.KeyMsg:
		return m, m.handleKey(msg)
	}

	return m, nil
}

func (m model) View() string {
	content := ""
	switch m.step {
	case stepWelcome:
		content = m.viewWelcome()
	case stepDetect:
		content = m.viewDetect()
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

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" || key == "q" {
		return tea.Quit
	}

	switch m.step {
	case stepWelcome:
		if key == "enter" {
			m.step = stepDetect
			m.animateHeader()
			return detectCmd(m.root)
		}
	case stepDetect:
		return nil
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		m.cursor = clampCursor(m.cursor, len(m.filteredServices(m.step)))
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.services)-1 {
				m.cursor++
			}
		case " ":
			if len(m.services) > 0 {
				m.toggleCurrentSelection()
			}
		case "enter":
			m.step = m.nextStep()
			m.animateHeader()
		case "b":
			m.step = m.prevStep()
			m.animateHeader()
		}
	case stepReview:
		switch key {
		case "enter":
			if len(m.warnings) > 0 {
				return nil
			}
			m.step = stepGenerate
			m.animateHeader()
			return generateCmd(m.root, selectedServiceIDs(m.services, m.selected))
		case "b":
			m.step = stepProxy
			m.animateHeader()
		}
	case stepGenerate:
		return nil
	case stepResult:
		if key == "enter" {
			return tea.Quit
		}
	case stepError:
		switch key {
		case "r":
			return m.retryFromError()
		case "b":
			m.step = m.previousStep
			m.animateHeader()
			return nil
		}
	}

	return nil
}

func (m *model) retryFromError() tea.Cmd {
	if m.previousStep == stepDetect {
		m.step = stepDetect
		m.animateHeader()
		return detectCmd(m.root)
	}
	if m.previousStep == stepGenerate {
		m.step = stepGenerate
		m.animateHeader()
		return generateCmd(m.root, selectedServiceIDs(m.services, m.selected))
	}
	return nil
}

func (m *model) animateHeader() {
	m.headerPos = 1
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

func (m model) renderHeader() string {
	indent := int(4 * m.headerPos)
	padding := strings.Repeat(" ", indent)

	stepText := fmt.Sprintf("Step %d/9", m.stepIndex())
	langText := "language: detecting"
	if m.langDetected {
		langText = "language: " + languageLabelWithVersion(m.langDetails)
	}
	progress := progressBar(m.stepIndex(), 9, 22)

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
		return "detecting language..."
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		return "up/down move | space toggle | enter next | b back | q quit"
	case stepReview:
		if len(m.warnings) > 0 {
			return "resolve warnings to continue | b back | q quit"
		}
		return "enter generate | b back | q quit"
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
	text := "Detecting project language"
	line := fmt.Sprintf("%s %s", m.spinner.View(), text)
	return cardStyle(m.width).Render(sectionTitle("Detect") + "\n\n" + line)
}

func (m model) viewServices(step step) string {
	filtered := m.filteredServices(step)
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
	return cardStyle(m.width).Render(sectionTitle(stepTitle(step)) + "\n\n" + content)
}

func (m model) viewReview() string {
	groups := m.selectedByCategory()
	body := []string{
		"Review your selections:",
		"",
		fmt.Sprintf("Detected language: %s", languageLabelWithVersion(m.langDetails)),
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

	body = append(body,
		"Files to be generated:",
		"- docker-compose.yml",
		"- Dockerfile",
	)
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

func (m model) viewGenerate() string {
	line := fmt.Sprintf("%s Generating docker-compose.yml and Dockerfile", m.spinner.View())
	return cardStyle(m.width).Render(sectionTitle("Generate") + "\n\n" + line)
}

func (m model) viewResult() string {
	body := []string{
		successTitle().Render("All set."),
		"",
		"Generated files:",
		"- " + m.output.ComposePath,
		"- " + m.output.DockerfilePath,
		"",
		"Next steps:",
		"- docker compose up",
	}
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

func (m *model) toggleCurrentSelection() {
	services := m.filteredServices(m.step)
	if len(services) == 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(services) {
		return
	}
	id := services[m.cursor].ID
	if m.selected[id] {
		delete(m.selected, id)
		return
	}
	m.selected[id] = true
}

func (m model) filteredServices(step step) []serviceChoice {
	category := stepCategory(step)
	if category == "" {
		return nil
	}
	filtered := make([]serviceChoice, 0, len(m.services))
	for _, svc := range m.services {
		if svc.Category == category {
			filtered = append(filtered, svc)
		}
	}
	return filtered
}

func (m model) selectedByCategory() map[string][]string {
	grouped := map[string][]string{}
	for _, svc := range m.services {
		if !m.selected[svc.ID] {
			continue
		}
		grouped[svc.Category] = append(grouped[svc.Category], svc.Label)
	}
	return grouped
}

func stepCategory(step step) string {
	switch step {
	case stepDatabase:
		return "database"
	case stepMessageQueue:
		return "message-queue"
	case stepCache:
		return "cache"
	case stepAnalytics:
		return "analytics"
	case stepProxy:
		return "proxy"
	default:
		return ""
	}
}

func stepTitle(step step) string {
	switch step {
	case stepDatabase:
		return "Databases"
	case stepMessageQueue:
		return "Message Queues"
	case stepCache:
		return "Caching"
	case stepAnalytics:
		return "Analytics"
	case stepProxy:
		return "Webservers / Proxies"
	default:
		return "Services"
	}
}

func categoryLabel(category string) string {
	switch category {
	case "database":
		return "Databases"
	case "message-queue":
		return "Message Queues"
	case "cache":
		return "Caching"
	case "analytics":
		return "Analytics"
	case "proxy":
		return "Webservers / Proxies"
	default:
		return category
	}
}

func categoryOrder() []string {
	return []string{"database", "message-queue", "cache", "analytics", "proxy"}
}

func clampCursor(cursor int, length int) int {
	if length <= 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func (m *model) nextStep() step {
	switch m.step {
	case stepDatabase:
		return stepMessageQueue
	case stepMessageQueue:
		return stepCache
	case stepCache:
		return stepAnalytics
	case stepAnalytics:
		return stepProxy
	case stepProxy:
		m.warnings = m.validateWarnings()
		return stepReview
	default:
		return m.step
	}
}

func (m *model) prevStep() step {
	switch m.step {
	case stepDatabase:
		return stepWelcome
	case stepMessageQueue:
		return stepDatabase
	case stepCache:
		return stepMessageQueue
	case stepAnalytics:
		return stepCache
	case stepProxy:
		return stepAnalytics
	default:
		return m.step
	}
}

func (m model) stepIndex() int {
	switch m.step {
	case stepWelcome:
		return 1
	case stepDetect:
		return 2
	case stepDatabase:
		return 3
	case stepMessageQueue:
		return 4
	case stepCache:
		return 5
	case stepAnalytics:
		return 6
	case stepProxy:
		return 7
	case stepReview:
		return 8
	case stepGenerate:
		return 9
	case stepResult:
		return 9
	case stepError:
		return 9
	default:
		return 1
	}
}

func detectCmd(root string) tea.Cmd {
	return func() tea.Msg {
		details, err := generator.DetectLanguage(root)
		return detectDoneMsg{details: details, err: err}
	}
}

func generateCmd(root string, services []string) tea.Cmd {
	return func() tea.Msg {
		lang, err := generator.DetectLanguage(root)
		if err != nil {
			return generateDoneMsg{err: err}
		}
		dockerfile, err := generator.Dockerfile(lang)
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

func serviceChoicesFromCatalog(services []generator.ServiceSpec) []serviceChoice {
	choices := make([]serviceChoice, 0, len(services))
	for _, svc := range services {
		choices = append(choices, serviceChoice{
			ID:          svc.ID,
			Label:       svc.Label,
			Description: svc.Description,
			Category:    svc.Category,
		})
	}
	return choices
}

func selectedServiceIDs(services []serviceChoice, selected map[string]bool) []string {
	ids := make([]string, 0, len(services))
	for _, svc := range services {
		if selected[svc.ID] {
			ids = append(ids, svc.ID)
		}
	}
	return ids
}

func languageLabel(lang generator.Language) string {
	switch lang {
	case generator.LanguageGo:
		return "Go"
	case generator.LanguageNode:
		return "Node"
	case generator.LanguagePython:
		return "Python"
	case generator.LanguageRuby:
		return "Ruby"
	case generator.LanguagePHP:
		return "PHP"
	case generator.LanguageJava:
		return "Java"
	case generator.LanguageDotNet:
		return ".NET"
	case generator.LanguageUnknown:
		return "Unknown"
	default:
		return string(lang)
	}
}

func languageLabelWithVersion(details generator.LanguageDetails) string {
	base := languageLabel(details.Type)
	version := ""
	switch details.Type {
	case generator.LanguageGo:
		version = details.GoVersion
	case generator.LanguageNode:
		version = details.NodeVersion
	case generator.LanguagePython:
		version = details.PythonVersion
	case generator.LanguageRuby:
		version = details.RubyVersion
	case generator.LanguagePHP:
		version = details.PHPVersion
	case generator.LanguageJava:
		version = details.JavaVersion
	case generator.LanguageDotNet:
		version = details.DotNetVersion
	}
	if version == "" {
		return base
	}
	return base + " " + version
}

func (m model) validateWarnings() []string {
	warnings := []string{}
	composePath := filepath.Join(m.root, generator.ComposeFileName)
	dockerfilePath := filepath.Join(m.root, generator.DockerfileFileName)
	if fileExists(composePath) {
		warnings = append(warnings, "docker-compose.yml already exists")
	}
	if fileExists(dockerfilePath) {
		warnings = append(warnings, "Dockerfile already exists")
	}
	return warnings
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

const (
	paletteBg     = lipgloss.Color("#1a1b26")
	palettePanel  = lipgloss.Color("#1f2335")
	paletteBorder = lipgloss.Color("#3b4261")
	paletteText   = lipgloss.Color("#c0caf5")
	paletteMuted  = lipgloss.Color("#565f89")
	paletteAccent = lipgloss.Color("#7aa2f7")
	paletteCyan   = lipgloss.Color("#7dcfff")
	paletteGreen  = lipgloss.Color("#9ece6a")
	paletteYellow = lipgloss.Color("#e0af68")
	paletteRed    = lipgloss.Color("#f7768e")
)

func headerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Background(palettePanel).
		Foreground(paletteText).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder)
}

func footerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		Foreground(paletteMuted).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder)
}

func cardStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Margin(1, 2).
		Background(paletteBg).
		Border(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder).
		Foreground(paletteText)
}

func errorStyle(width int) lipgloss.Style {
	return cardStyle(width).
		BorderForeground(paletteRed).
		Foreground(paletteRed)
}

func badgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder).
		Foreground(paletteCyan)
}

func mutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(paletteMuted)
}

func serviceLineStyle(active bool, selected bool) lipgloss.Style {
	style := lipgloss.NewStyle()
	if selected {
		style = style.Foreground(paletteGreen)
	} else {
		style = style.Foreground(paletteText)
	}
	if active {
		style = style.Bold(true).Foreground(paletteAccent)
	}
	return style
}

func warningTitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(paletteYellow)
}

func successTitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(paletteGreen)
}

func sectionTitle(title string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(paletteAccent).Render(title)
}

func setSpinner(sp *spinner.Model) {
	sp.Spinner = spinner.Line
	sp.Style = lipgloss.NewStyle().Foreground(paletteAccent)
}

func progressBar(current, total, width int) string {
	if total <= 0 {
		return "[----------]"
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	filled := int(float64(width) * (float64(current) / float64(total)))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
	return "[" + bar + "]"
}

func titleArt() []string {
	return []string{
		"██████╗  ██████╗  ██████╗██╗  ██╗███████╗██████╗",
		"██╔══██╗██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗",
		"██║  ██║██║   ██║██║     █████╔╝ █████╗  ██████╔╝",
		"██║  ██║██║   ██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗",
		"██████╔╝╚██████╔╝╚██████╗██║  ██╗███████╗██║  ██║",
		"╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝",
		"██╗    ██╗██╗███████╗ █████╗ ██████╗ ██████╗",
		"██║    ██║██║╚══███╔╝██╔══██╗██╔══██╗██╔══██╗",
		"██║ █╗ ██║██║  ███╔╝ ███████║██████╔╝██║  ██║",
		"██║███╗██║██║ ███╔╝  ██╔══██║██╔══██╗██║  ██║",
		"╚███╔███╔╝██║███████╗██║  ██║██║  ██║██████╔╝",
		" ╚══╝╚══╝ ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝",
	}
}

func titleColor(frame int) lipgloss.Color {
	if frame%24 < 12 {
		return paletteAccent
	}
	return paletteCyan
}
