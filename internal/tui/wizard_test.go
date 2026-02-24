package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
