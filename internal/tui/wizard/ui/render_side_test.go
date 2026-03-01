package ui

import (
	"strings"
	"testing"
)

func TestShouldShowSidePanel(t *testing.T) {
	tests := []struct {
		name   string
		mode   RenderMode
		width  int
		expect bool
	}{
		{"styled wide enough", RenderModeStyled, 120, true},
		{"styled exact minimum", RenderModeStyled, minSideLayout, true},
		{"styled too narrow", RenderModeStyled, minSideLayout - 1, false},
		{"plain wide", RenderModePlain, 200, false},
		{"plain narrow", RenderModePlain, 80, false},
		{"styled zero width", RenderModeStyled, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := currentRenderMode
			defer func() { currentRenderMode = prev }()
			currentRenderMode = tt.mode

			got := shouldShowSidePanel(tt.width)
			if got != tt.expect {
				t.Fatalf("shouldShowSidePanel(%d) = %v, want %v", tt.width, got, tt.expect)
			}
		})
	}
}

func TestSidePanelWidth(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModePlain
	if got := sidePanelWidth(200); got != 0 {
		t.Fatalf("plain mode: sidePanelWidth(200) = %d, want 0", got)
	}

	currentRenderMode = RenderModeStyled
	if got := sidePanelWidth(120); got != sidePanelW {
		t.Fatalf("styled normal: sidePanelWidth(120) = %d, want %d", got, sidePanelW)
	}
	if got := sidePanelWidth(wideSideThreshold); got != wideSidePanelW {
		t.Fatalf("styled wide: sidePanelWidth(%d) = %d, want %d", wideSideThreshold, got, wideSidePanelW)
	}
	if got := sidePanelWidth(wideSideThreshold + 20); got != wideSidePanelW {
		t.Fatalf("styled extra wide: sidePanelWidth(%d) = %d, want %d", wideSideThreshold+20, got, wideSidePanelW)
	}
}

func TestMainPanelWidth(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModeStyled
	// When side panel is shown, main = width - side - 1
	w := 120
	expected := w - sidePanelW - 1
	if got := mainPanelWidth(w); got != expected {
		t.Fatalf("mainPanelWidth(%d) = %d, want %d", w, got, expected)
	}

	// When too narrow for side panel, main = full width
	narrow := minSideLayout - 1
	if got := mainPanelWidth(narrow); got != narrow {
		t.Fatalf("mainPanelWidth(%d) = %d, want %d", narrow, got, narrow)
	}

	// Plain mode returns full width
	currentRenderMode = RenderModePlain
	if got := mainPanelWidth(200); got != 200 {
		t.Fatalf("plain mode: mainPanelWidth(200) = %d, want 200", got)
	}
}

func TestMainContentWidth(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModeStyled

	// MainContentWidth should equal mainPanelWidth
	for _, w := range []int{80, 100, 120, 150} {
		got := MainContentWidth(w)
		expected := mainPanelWidth(w)
		if got != expected {
			t.Fatalf("MainContentWidth(%d) = %d, want %d", w, got, expected)
		}
	}
}

func TestRenderPlainModeNoSidePanel(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModePlain
	s := State{
		Width:      200,
		Height:     40,
		Step:       StepWelcome,
		StepIndex:  1,
		TotalSteps: 7,
		SideTitle:  "Status",
		SideLines:  []string{"Step 1/7", "Stage: welcome"},
	}

	output := Render(s)
	if strings.Contains(output, "Status") && strings.Contains(output, "Stage: welcome") {
		t.Fatal("plain mode should not render side panel content")
	}
}

func TestRenderStyledNarrowNoSidePanel(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModeStyled
	s := State{
		Width:       minSideLayout - 1,
		Height:      40,
		Step:        StepWelcome,
		StepIndex:   1,
		TotalSteps:  7,
		ProjectName: "test-project",
		SideTitle:   "Status",
		SideLines:   []string{"Step 1/7", "Stage: welcome"},
	}

	output := Render(s)
	// Side panel content should NOT appear in the output
	if strings.Contains(output, "Stage: welcome") {
		t.Fatal("narrow styled mode should not render side panel")
	}
}

func TestRenderStyledWideHasSidePanel(t *testing.T) {
	prev := currentRenderMode
	defer func() { currentRenderMode = prev }()

	currentRenderMode = RenderModeStyled
	s := State{
		Width:       120,
		Height:      40,
		Step:        StepWelcome,
		StepIndex:   1,
		TotalSteps:  7,
		ProjectName: "test-project",
		SideTitle:   "Status",
		SideLines:   []string{"Step 1/7", "Stage: welcome", "", "Tip: press enter"},
	}

	output := Render(s)
	// Side panel content should appear in the output
	if !strings.Contains(output, "Status") {
		t.Fatal("wide styled mode should render side panel with title")
	}
	if !strings.Contains(output, "Stage: welcome") {
		t.Fatal("wide styled mode should render side panel lines")
	}
}
