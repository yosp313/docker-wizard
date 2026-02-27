package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	paletteBg     = lipgloss.Color("#1a1b26")
	palettePanel  = lipgloss.Color("#1f2335")
	palettePanel2 = lipgloss.Color("#24283b")
	paletteBorder = lipgloss.Color("#3b4261")
	paletteText   = lipgloss.Color("#c0caf5")
	paletteMuted  = lipgloss.Color("#565f89")
	paletteAccent = lipgloss.Color("#7aa2f7")
	paletteCyan   = lipgloss.Color("#7dcfff")
	paletteGreen  = lipgloss.Color("#9ece6a")
	paletteYellow = lipgloss.Color("#e0af68")
	paletteRed    = lipgloss.Color("#f7768e")
	maxContentW   = 120
)

type RenderMode string

const (
	RenderModeStyled RenderMode = "styled"
	RenderModePlain  RenderMode = "plain"
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

func isPlainMode() bool {
	return currentRenderMode == RenderModePlain
}

func contentWidth(width int) int {
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
		BorderStyle(lipgloss.NormalBorder()).
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
		Width(contentWidth(width)).
		Padding(1, 2).
		Margin(1, 1).
		Background(paletteBg).
		Border(lipgloss.NormalBorder()).
		BorderForeground(paletteBorder).
		Foreground(paletteText)
}

func errorStyle(width int) lipgloss.Style {
	if isPlainMode() {
		return cardStyle(width)
	}
	return cardStyle(width).
		BorderForeground(paletteRed).
		Foreground(paletteRed)
}

func badgeStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteCyan).Bold(true)
}

func mutedStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteMuted)
}

func serviceLineStyle(active bool, selected bool) lipgloss.Style {
	if isPlainMode() {
		style := lipgloss.NewStyle()
		if active {
			style = style.Bold(true)
		}
		return style
	}
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

func keycapStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteCyan)
}

func doneChipStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteGreen)
}

func currentChipStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteAccent).Bold(true)
}

func pendingChipStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteMuted)
}

func activePreviewTabStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(paletteAccent)
}

func inactivePreviewTabStyle() lipgloss.Style {
	if isPlainMode() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(paletteMuted)
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

func setSpinner(sp *spinner.Model) {
	sp.Spinner = spinner.Line
	if isPlainMode() {
		return
	}
	sp.Style = lipgloss.NewStyle().Foreground(paletteAccent)
}

func progressChips(current int, total int) string {
	if total <= 0 {
		return ""
	}
	if current < 1 {
		current = 1
	}
	if current > total {
		current = total
	}
	done := current - 1
	left := total - current
	if isPlainMode() {
		return fmt.Sprintf("done %d | now %d/%d | left %d", done, current, total, left)
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		doneChipStyle().Render("done "+strconv.Itoa(done)),
		" | ",
		currentChipStyle().Render("now "+strconv.Itoa(current)+"/"+strconv.Itoa(total)),
		" | ",
		pendingChipStyle().Render("left "+strconv.Itoa(left)),
	)
}

func titleColor(frame int) lipgloss.Color {
	if isPlainMode() {
		return paletteText
	}
	if frame%24 < 12 {
		return paletteAccent
	}
	return paletteCyan
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
