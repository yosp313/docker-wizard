package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	paletteBg         = lipgloss.Color("#1a1b26")
	palettePanel      = lipgloss.Color("#1f2335")
	paletteBorder     = lipgloss.Color("#3b4261")
	paletteText       = lipgloss.Color("#c0caf5")
	paletteMuted      = lipgloss.Color("#565f89")
	paletteAccent     = lipgloss.Color("#7aa2f7")
	paletteCyan       = lipgloss.Color("#7dcfff")
	paletteGreen      = lipgloss.Color("#9ece6a")
	paletteYellow     = lipgloss.Color("#e0af68")
	paletteRed        = lipgloss.Color("#f7768e")
	paletteRowFocus   = lipgloss.Color("#252640")
	maxContentW       = 120
	minSideLayout     = 100
	sidePanelW        = 34
	wideSidePanelW    = 38
	wideSideThreshold = 140
)

var currentRenderMode = RenderModeStyled

func SetRenderMode(mode RenderMode) {
	switch mode {
	case RenderModePlain:
		currentRenderMode = RenderModePlain
	default:
		currentRenderMode = RenderModeStyled
	}
}

func ConfigureSpinner(sp *spinner.Model) {
	sp.Spinner = spinner.Line
	if isPlainMode() {
		return
	}
	sp.Style = lipgloss.NewStyle().Foreground(paletteAccent)
}

func ContentWidth(width int) int {
	if width <= 0 {
		return 80
	}
	usable := width - 6
	if usable > maxContentW {
		usable = maxContentW
	}
	if usable < 20 {
		usable = 20
	}
	return usable
}

func isPlainMode() bool {
	return currentRenderMode == RenderModePlain
}

func shouldShowSidePanel(width int) bool {
	return !isPlainMode() && width >= minSideLayout
}

func sidePanelWidth(width int) int {
	if isPlainMode() {
		return 0
	}
	if width >= wideSideThreshold {
		return wideSidePanelW
	}
	return sidePanelW
}

func mainPanelWidth(width int) int {
	if !shouldShowSidePanel(width) {
		return width
	}
	return width - sidePanelWidth(width) - 1
}

// MainContentWidth returns the effective content width accounting for the
// side panel. Use this instead of raw terminal width when computing content
// dimensions that must fit in the main panel.
func MainContentWidth(width int) int {
	return mainPanelWidth(width)
}

func sidePanelStyle(width int, height int) lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1).
		Background(paletteBg).
		Foreground(paletteText).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(paletteBorder)
}

func headerStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Background(palettePanel).
		Foreground(paletteText).
		BorderBottom(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(paletteBorder)
}

func footerStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 2).
		Foreground(paletteMuted).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder)
}

func cardStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().
		Width(ContentWidth(width)).
		Padding(1, 2).
		Margin(1, 1).
		Background(palettePanel).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(paletteBorder).
		Foreground(paletteText)
}

func successCardStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return cardStyle(width)
	}
	return cardStyle(width).BorderForeground(paletteGreen)
}

func errorStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return cardStyle(width)
	}
	return cardStyle(width).
		BorderForeground(paletteRed).
		Foreground(paletteRed)
}

func mutedStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteMuted)
}

func keycapStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().
		Background(paletteCyan).
		Foreground(paletteBg).
		Bold(true).
		Padding(0, 1)
}

func activePreviewTabStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().
		Background(paletteAccent).
		Foreground(paletteBg).
		Bold(true).
		Padding(0, 1)
}

func inactivePreviewTabStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().
		Foreground(paletteMuted).
		Padding(0, 1)
}

func blockerTitle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteRed)
}

func warningTitle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteYellow)
}

func successTitle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteGreen)
}

func sectionTitle(title string) string {
	if isPlainMode() {
		return title
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteAccent).Render(title)
}

func progressBar(current int, total int) string {
	if total <= 0 {
		return ""
	}
	if current < 1 {
		current = 1
	}
	if current > total {
		current = total
	}
	const barWidth = 20
	filled := barWidth * current / total
	if filled < 1 {
		filled = 1
	}
	return fmt.Sprintf("[%s%s] %d/%d", strings.Repeat("#", filled), strings.Repeat(".", barWidth-filled), current, total)
}

func plainHeader(stepIndex int, total int, stepName string, langText string, projectName string, progress string, indent int) string {
	lines := []string{
		"Docker Wizard",
		fmt.Sprintf("step %d/%d (%s) | %s", stepIndex, total, stepName, langText),
		"project: " + projectName,
		progress,
	}
	if indent <= 0 {
		return strings.Join(lines, "\n")
	}
	pad := strings.Repeat(" ", indent)
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}
