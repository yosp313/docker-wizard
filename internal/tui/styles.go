package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

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

func blockerTitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(paletteRed)
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

func progressBar(current int, total int, width int) string {
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
