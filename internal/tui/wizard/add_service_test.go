package wizard

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func makeModelWithCatalog(t *testing.T) (model, string) {
	t.Helper()
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	catalog := `{"services":[{"id":"redis","name":"Redis","label":"Redis","category":"cache","selectable":true,"order":1}]}`
	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}
	m := model{
		root:             root,
		step:             stepDatabase,
		selected:         map[string]bool{},
		addServiceInputs: initAddServiceInputs(),
		services: []serviceChoice{
			{ID: "redis", Label: "Redis", Category: "cache"},
		},
	}
	return m, root
}

// Task 5.1: pressing n on a service-selection step opens stepAddService,
// and Escape returns to the originating step.
func TestAddService_NKeyOpensForm(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepDatabase

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if m.step != stepAddService {
		t.Fatalf("expected stepAddService after n, got %v", m.step)
	}
	if m.previousStep != stepDatabase {
		t.Fatalf("expected previousStep=stepDatabase, got %v", m.previousStep)
	}
}

func TestAddService_EscapeReturnsToOriginatingStep(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepDatabase
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	// Now on stepAddService; press Escape
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyEscape})

	if m.step != stepDatabase {
		t.Fatalf("expected stepDatabase after Escape, got %v", m.step)
	}
}

func TestAddService_EscapeFromMessageQueueStep(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepMessageQueue
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyEscape})

	if m.step != stepMessageQueue {
		t.Fatalf("expected stepMessageQueue after Escape, got %v", m.step)
	}
}

// Task 5.2: confirming the form with valid input calls AppendService and
// the new service appears in m.services.
func TestAddService_ConfirmValidInputAddsService(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepDatabase
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Type into name field (field 0)
	for _, r := range "My Kafka" {
		m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Tab to image field (field 1)
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "confluentinc/cp-kafka:7" {
		m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Tab to category (field 2), then Down to select "message-queue"
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyTab})
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyDown}) // database(0)->message-queue(1)

	// Confirm
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepDatabase {
		t.Fatalf("expected to return to stepDatabase, got %v (formError: %q)", m.step, m.addServiceFormError)
	}

	found := false
	for _, svc := range m.services {
		if svc.ID == "my-kafka" {
			found = true
			break
		}
	}
	if !found {
		ids := make([]string, 0, len(m.services))
		for _, s := range m.services {
			ids = append(ids, s.ID)
		}
		t.Fatalf("expected my-kafka in m.services, got %v", ids)
	}
}

// Task 5.3: confirming with empty name or empty image shows a form error
// and does not call AppendService.
func TestAddService_EmptyNameShowsError(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepDatabase
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Leave name blank, tab to image and fill it
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "redis:7" {
		m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Try to confirm
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepAddService {
		t.Fatalf("expected to stay on stepAddService, got %v", m.step)
	}
	if m.addServiceFormError == "" {
		t.Fatal("expected formError to be set for empty name")
	}
}

func TestAddService_EmptyImageShowsError(t *testing.T) {
	m, _ := makeModelWithCatalog(t)
	m.step = stepDatabase
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Type name but leave image blank
	for _, r := range "My Service" {
		m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Try to confirm (image is still empty)
	m.handleAddServiceMsg(tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepAddService {
		t.Fatalf("expected to stay on stepAddService, got %v", m.step)
	}
	if m.addServiceFormError == "" {
		t.Fatal("expected formError to be set for empty image")
	}
}
