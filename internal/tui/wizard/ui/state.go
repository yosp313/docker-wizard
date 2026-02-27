package ui

type RenderMode string

const (
	RenderModeStyled RenderMode = "styled"
	RenderModePlain  RenderMode = "plain"
)

type Step string

const (
	StepWelcome      Step = "welcome"
	StepDetect       Step = "detect"
	StepLanguage     Step = "language"
	StepDatabases    Step = "databases"
	StepMessageQueue Step = "message queues"
	StepCache        Step = "cache"
	StepAnalytics    Step = "analytics"
	StepProxies      Step = "proxies"
	StepReview       Step = "review"
	StepPreview      Step = "preview"
	StepGenerate     Step = "generate"
	StepResult       Step = "result"
	StepError        Step = "error"
)

type OptionItem struct {
	Label       string
	Description string
	Active      bool
	Selected    bool
}

type ReviewGroup struct {
	Label string
	Items []string
}

type PreviewTab struct {
	Name   string
	Short  string
	Status string
	Active bool
}

type State struct {
	Width      int
	Height     int
	Frame      int
	Step       Step
	StepIndex  int
	TotalSteps int

	HeaderIndent int
	LanguageText string
	ProjectName  string
	FooterRaw    string

	SpinnerText      string
	DetectDone       bool
	DetectedLanguage string

	LanguageOptions []OptionItem
	ServiceTitle    string
	ServiceOptions  []OptionItem

	ReviewGroups []ReviewGroup
	ManagedFiles []string
	Warnings     []string
	Blockers     []string
	CreateIgnore bool

	PreviewReady    bool
	PreviewTabs     []PreviewTab
	PreviewFileLine string
	PreviewLineInfo string
	PreviewDivider  string
	PreviewBody     string

	ResultFiles     []string
	ResultBackups   []string
	ResultNextSteps []string

	ErrorMessage string
}
