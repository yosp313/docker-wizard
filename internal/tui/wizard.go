package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

type step int

const (
	stepWelcome step = iota
	stepDetect
	stepLanguage
	stepDatabase
	stepMessageQueue
	stepCache
	stepAnalytics
	stepProxy
	stepReview
	stepPreview
	stepGenerate
	stepResult
	stepError
)

const totalSteps = 11

type tickMsg time.Time

type detectDoneMsg struct {
	details generator.LanguageDetails
	err     error
}

type generateDoneMsg struct {
	output generator.Output
	err    error
}

type previewDoneMsg struct {
	preview generator.Preview
	err     error
}

type serviceChoice struct {
	ID          string
	Label       string
	Selected    bool
	Description string
	Category    string
}

type languageChoice struct {
	ID          string
	Label       string
	Description string
	Language    generator.Language
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
	langOptions  []languageChoice
	langCursor   int
	overrideLang bool
	overrideType generator.Language
	detectDone   bool
	langVisited  bool

	services           []serviceChoice
	cursor             int
	selected           map[string]bool
	warnings           []string
	blockers           []string
	createDockerignore bool
	previewContent     string
	previewViewport    viewport.Model
	frame              int

	output       generator.Output
	preview      generator.Preview
	previewReady bool
	err          error

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
		langOptions:  defaultLanguageOptions(),
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
		m.syncPreviewViewportSize()
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
		m.langOptions = languageOptionsForDetected(m.langDetails)
		m.detectDone = true
		m.step = stepDetect
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
	case previewDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.previousStep = stepPreview
			m.step = stepError
			return m, nil
		}
		m.preview = msg.preview
		m.previewReady = true
		m.previewContent = strings.Join(buildPreviewLines(msg.preview, m.blockers), "\n")
		m.setPreviewViewportContent(m.previewContent)
		m.step = stepPreview
		m.animateHeader()
		return m, nil
	case tea.KeyMsg:
		return m, m.handleKey(msg)
	}

	return m, nil
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
		if !m.detectDone {
			return nil
		}
		switch key {
		case "enter":
			m.step = stepDatabase
			m.animateHeader()
		case "l":
			m.langVisited = true
			m.step = stepLanguage
			m.animateHeader()
		case "b":
			m.step = stepWelcome
			m.animateHeader()
		}
		return nil
	case stepLanguage:
		switch key {
		case "up", "k":
			if m.langCursor > 0 {
				m.langCursor--
			}
		case "down", "j":
			if m.langCursor < len(m.langOptions)-1 {
				m.langCursor++
			}
		case "enter":
			m.applyLanguageChoice()
			m.langVisited = true
			m.step = stepDatabase
			m.animateHeader()
		case "b":
			m.step = stepDetect
			m.animateHeader()
		}
		return nil
	case stepDatabase, stepMessageQueue, stepCache, stepAnalytics, stepProxy:
		services := m.filteredServices(m.step)
		m.cursor = clampCursor(m.cursor, len(services))
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(services)-1 {
				m.cursor++
			}
		case " ":
			if len(services) > 0 {
				m.toggleCurrentSelection()
			}
		case "enter":
			if m.step == stepProxy {
				if err := m.prepareReview(); err != nil {
					m.err = err
					m.previousStep = stepProxy
					m.step = stepError
					return nil
				}
				m.step = stepReview
				m.animateHeader()
				return nil
			}
			m.step = m.nextStep()
			m.animateHeader()
		case "b":
			m.step = m.prevStep()
			m.animateHeader()
		}
	case stepReview:
		switch key {
		case "enter":
			if len(m.blockers) > 0 {
				return nil
			}
			m.step = stepGenerate
			m.animateHeader()
			return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
		case "p":
			m.previewReady = false
			m.preview = generator.Preview{}
			m.previewContent = ""
			m.setPreviewViewportContent("")
			m.step = stepPreview
			m.animateHeader()
			return previewCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
		case "b":
			m.step = stepProxy
			m.animateHeader()
		}
	case stepPreview:
		if key == "b" {
			m.step = stepReview
			m.animateHeader()
			return nil
		}
		if !m.previewReady {
			return nil
		}
		switch key {
		case "home":
			m.previewViewport.GotoTop()
			return nil
		case "end":
			m.previewViewport.GotoBottom()
			return nil
		case "enter":
			if len(m.blockers) > 0 {
				return nil
			}
			m.step = stepGenerate
			m.animateHeader()
			return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
		}
		var cmd tea.Cmd
		m.previewViewport, cmd = m.previewViewport.Update(msg)
		return cmd
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
	if m.previousStep == stepProxy {
		if err := m.prepareReview(); err != nil {
			m.err = err
			m.step = stepError
			m.animateHeader()
			return nil
		}
		m.step = stepReview
		m.animateHeader()
		return nil
	}
	if m.previousStep == stepGenerate {
		m.step = stepGenerate
		m.animateHeader()
		return generateCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
	}
	if m.previousStep == stepPreview {
		m.step = stepPreview
		m.animateHeader()
		return previewCmd(m.root, selectedServiceIDs(m.services, m.selected), m.overrideLang, m.overrideType)
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

func (m model) effectiveDetails() generator.LanguageDetails {
	if !m.overrideLang {
		return m.langDetails
	}
	updated := m.langDetails
	updated.Type = m.overrideType
	return updated
}

func (m *model) applyLanguageChoice() {
	if m.langCursor < 0 || m.langCursor >= len(m.langOptions) {
		return
	}
	choice := m.langOptions[m.langCursor]
	if choice.ID == "auto" {
		m.overrideLang = false
		m.overrideType = ""
		return
	}
	m.overrideLang = true
	m.overrideType = choice.Language
}

func (m model) isLanguageSelected(option languageChoice) bool {
	if option.ID == "auto" {
		return !m.overrideLang
	}
	if !m.overrideLang {
		return false
	}
	return m.overrideType == option.Language
}
