package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docker-wizard/internal/generator"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const minimalDockerfileCatalogJSON = `{"dockerfiles":[{"language":"unknown","templateLines":["FROM alpine:3.20","WORKDIR /app","COPY . .","ENV APP_START_CMD=\"sh\"","CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]"]}]}`

func TestHandleKey_ServiceCursorUsesFilteredLength(t *testing.T) {
	m := model{
		step: stepDatabase,
		services: []serviceChoice{
			{ID: "mysql", Category: "database"},
			{ID: "redis", Category: "cache"},
		},
		selected: map[string]bool{},
		cursor:   0,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})

	if m.cursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestRetryFromErrorProxyTransitionsToReview(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(`{"services":[]}`), 0o644); err != nil {
		t.Fatalf("write services catalog: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "dockerfiles.json"), []byte(minimalDockerfileCatalogJSON), 0o644); err != nil {
		t.Fatalf("write dockerfile catalog: %v", err)
	}

	m := model{
		root:         root,
		step:         stepError,
		previousStep: stepProxy,
		selected:     map[string]bool{},
	}

	cmd := m.retryFromError()

	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if m.step != stepReview {
		t.Fatalf("expected step %v, got %v", stepReview, m.step)
	}
	if m.err != nil {
		t.Fatalf("expected nil error, got %v", m.err)
	}
}

func TestRetryFromErrorProxyKeepsErrorStepOnFailure(t *testing.T) {
	m := model{
		root:         "",
		step:         stepError,
		previousStep: stepProxy,
		selected:     map[string]bool{},
	}

	cmd := m.retryFromError()

	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if m.step != stepError {
		t.Fatalf("expected step %v, got %v", stepError, m.step)
	}
	if m.err == nil {
		t.Fatal("expected error to be set")
	}
}

func TestHandleKey_PreviewHomeAndEnd(t *testing.T) {
	content := strings.Join([]string{
		"1", "2", "3", "4", "5", "6", "7", "8",
	}, "\n")

	m := model{
		step:         stepPreview,
		previewReady: true,
		previewViewport: viewport.Model{
			Width:  10,
			Height: 3,
		},
	}
	m.previewViewport.SetContent(content)

	m.handleKey(tea.KeyMsg{Type: tea.KeyEnd})
	if m.previewViewport.YOffset == 0 {
		t.Fatal("expected preview offset to move to bottom")
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyHome})
	if m.previewViewport.YOffset != 0 {
		t.Fatalf("expected preview offset at top, got %d", m.previewViewport.YOffset)
	}
}

func TestHandleKey_PreviewTabSwitchArrowsAndNumbers(t *testing.T) {
	m := model{
		step:         stepPreview,
		previewReady: true,
		preview: generator.Preview{
			Compose:    generator.FilePreview{Status: generator.FileStatusNew, Content: "compose\nline2\nline3\nline4"},
			Dockerfile: generator.FilePreview{Status: generator.FileStatusNew, Content: "dockerfile\nline2\nline3\nline4"},
			Dockerignore: generator.FilePreview{Status: generator.FileStatusNew,
				Content: "dockerignore\nline2\nline3\nline4",
			},
		},
		previewTab: 0,
		width:      100,
		height:     40,
	}
	m.refreshPreviewTabContent()

	m.previewViewport.YOffset = 2
	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if m.previewTab != 1 {
		t.Fatalf("expected preview tab 1, got %d", m.previewTab)
	}
	if !strings.Contains(m.previewContent, "dockerfile") {
		t.Fatalf("expected dockerfile content, got %q", m.previewContent)
	}
	if m.previewViewport.YOffset != 0 {
		t.Fatalf("expected preview offset reset to top, got %d", m.previewViewport.YOffset)
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.previewTab != 2 {
		t.Fatalf("expected preview tab 2, got %d", m.previewTab)
	}
	if !strings.Contains(m.previewContent, "dockerignore") {
		t.Fatalf("expected dockerignore content, got %q", m.previewContent)
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if m.previewTab != 1 {
		t.Fatalf("expected preview tab 1 after left, got %d", m.previewTab)
	}
}

func TestActivePreviewTabContentExistsStatus(t *testing.T) {
	m := model{
		preview: generator.Preview{
			Compose:      generator.FilePreview{Status: generator.FileStatusNew, Content: "compose"},
			Dockerfile:   generator.FilePreview{Status: generator.FileStatusNew, Content: "dockerfile"},
			Dockerignore: generator.FilePreview{Status: generator.FileStatusExists},
		},
		previewTab: 2,
	}

	got := m.activePreviewTabContent()
	if got != "existing file will be kept" {
		t.Fatalf("expected exists message, got %q", got)
	}
}

func TestHandleKey_ReviewEnterBlockedByBlockers(t *testing.T) {
	m := model{
		step:     stepReview,
		blockers: []string{"missing required field"},
		selected: map[string]bool{},
	}

	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Fatalf("expected no command while blockers exist")
	}
	if m.step != stepReview {
		t.Fatalf("expected step to remain review, got %v", m.step)
	}
}
