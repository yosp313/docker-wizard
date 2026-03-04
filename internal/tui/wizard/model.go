package wizard

import (
	"time"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	stepAddService
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
	previewTab         int
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

	// add-service form state
	addServiceFocusedField int
	addServiceInputs       [5]textinput.Model // name, image, ports, env vars, volume mounts
	addServiceCategoryIdx  int
	addServiceFormError    string
}
