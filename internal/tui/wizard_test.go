package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
