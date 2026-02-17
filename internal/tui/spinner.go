package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

type workDoneMsg struct {
	err error
}

type spinnerModel struct {
	spinner    spinner.Model
	message    string
	work       func() error
	done       bool
	err        error
	lastToggle time.Time
	pos        float64
	vel        float64
	target     float64
	spring     harmonica.Spring
}

func RunWithSpinner(message string, work func() error) error {
	m := spinnerModel{
		spinner: spinner.New(),
		message: message,
		work:    work,
		spring:  harmonica.NewSpring(harmonica.FPS(60), 5.0, 0.6),
	}
	m.spinner.Spinner = spinner.Dot

	program := tea.NewProgram(m)
	model, err := program.Run()
	if err != nil {
		return err
	}
	if finalModel, ok := model.(spinnerModel); ok {
		if finalModel.err != nil {
			return finalModel.err
		}
	}
	return nil
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return workDoneMsg{err: m.runWork()}
	})
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.updateSpring()
		return m, cmd
	case workDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.err = fmt.Errorf("cancelled")
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}

	indent := int(math.Round(m.pos * 2))
	pad := strings.Repeat(" ", indent)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return style.Render(pad+m.spinner.View()+" "+m.message) + "\n"
}

func (m *spinnerModel) runWork() error {
	if m.work == nil {
		return nil
	}
	return m.work()
}

func (m *spinnerModel) updateSpring() {
	now := time.Now()
	if m.lastToggle.IsZero() {
		m.lastToggle = now
	}
	if now.Sub(m.lastToggle) > 300*time.Millisecond {
		if m.target < 0.5 {
			m.target = 1
		} else {
			m.target = 0
		}
		m.lastToggle = now
	}
	m.pos, m.vel = m.spring.Update(m.pos, m.vel, m.target)
}
