package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeServices writes a ServiceCatalog JSON file to <root>/config/services.json.
func writeServices(t *testing.T, root string, cat ServiceCatalog) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatalf("setup: mkdir: %v", err)
	}
	data, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		t.Fatalf("setup: marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "services.json"), data, 0o644); err != nil {
		t.Fatalf("setup: write: %v", err)
	}
}

// loadServices reads and returns the ServiceCatalog from <root>/config/services.json.
func loadServices(t *testing.T, root string) ServiceCatalog {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "config", "services.json"))
	if err != nil {
		t.Fatalf("loadServices: %v", err)
	}
	var cat ServiceCatalog
	if err := json.Unmarshal(data, &cat); err != nil {
		t.Fatalf("loadServices unmarshal: %v", err)
	}
	return cat
}

// TestAppendService_HappyPath verifies that a valid service is persisted and can be reloaded.
func TestAppendService_HappyPath(t *testing.T) {
	root := t.TempDir()

	svc := ServiceSpec{
		Name:     "Redis",
		Label:    "Redis",
		Category: "cache",
		Image:    "redis:7",
	}

	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	cat := loadServices(t, root)
	if len(cat.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cat.Services))
	}
	got := cat.Services[0]
	if got.Category != "cache" {
		t.Errorf("category: got %q, want %q", got.Category, "cache")
	}
	if got.Image != "redis:7" {
		t.Errorf("image: got %q, want %q", got.Image, "redis:7")
	}
}

// TestAppendService_IDAutoGeneration checks that "My Redis" becomes id "my-redis".
func TestAppendService_IDAutoGeneration(t *testing.T) {
	root := t.TempDir()

	svc := ServiceSpec{
		Name:     "My Redis",
		Label:    "My Redis",
		Category: "cache",
		Image:    "redis:7",
	}

	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	cat := loadServices(t, root)
	if len(cat.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cat.Services))
	}
	if got := cat.Services[0].ID; got != "my-redis" {
		t.Errorf("id: got %q, want %q", got, "my-redis")
	}
}

// TestAppendService_IDCollisionResolution checks that two services with the same slug
// receive ids "my-redis" and "my-redis-2".
func TestAppendService_IDCollisionResolution(t *testing.T) {
	root := t.TempDir()

	svc := ServiceSpec{
		Name:     "My Redis",
		Label:    "My Redis",
		Category: "cache",
		Image:    "redis:7",
	}

	if err := AppendService(root, svc); err != nil {
		t.Fatalf("first AppendService: %v", err)
	}
	if err := AppendService(root, svc); err != nil {
		t.Fatalf("second AppendService: %v", err)
	}

	cat := loadServices(t, root)
	if len(cat.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cat.Services))
	}
	if id := cat.Services[0].ID; id != "my-redis" {
		t.Errorf("first id: got %q, want %q", id, "my-redis")
	}
	if id := cat.Services[1].ID; id != "my-redis-2" {
		t.Errorf("second id: got %q, want %q", id, "my-redis-2")
	}
}

// TestAppendService_MissingConfigDir verifies that AppendService creates the config/
// directory when it does not exist.
func TestAppendService_MissingConfigDir(t *testing.T) {
	root := t.TempDir()
	// Deliberately do NOT create root/config.

	svc := ServiceSpec{
		Name:     "Postgres",
		Label:    "Postgres",
		Category: "database",
		Image:    "postgres:16",
	}

	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "config", "services.json")); err != nil {
		t.Errorf("services.json not created: %v", err)
	}
}

// TestAppendService_ExistingServicesPreserved checks that appending a new service
// does not remove already-existing entries.
func TestAppendService_ExistingServicesPreserved(t *testing.T) {
	root := t.TempDir()

	existing := ServiceCatalog{
		Services: []ServiceSpec{
			{
				ID:         "postgres",
				Name:       "Postgres",
				Label:      "Postgres",
				Category:   "database",
				Image:      "postgres:16",
				Selectable: true,
				Order:      10,
			},
		},
	}
	writeServices(t, root, existing)

	newSvc := ServiceSpec{
		Name:     "Redis",
		Label:    "Redis",
		Category: "cache",
		Image:    "redis:7",
	}

	if err := AppendService(root, newSvc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	cat := loadServices(t, root)
	if len(cat.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cat.Services))
	}
	if cat.Services[0].ID != "postgres" {
		t.Errorf("first service id: got %q, want %q", cat.Services[0].ID, "postgres")
	}
	if cat.Services[1].ID != "redis" {
		t.Errorf("second service id: got %q, want %q", cat.Services[1].ID, "redis")
	}
}

// TestAppendService_InvalidCategory verifies that an unknown category is rejected.
func TestAppendService_InvalidCategory(t *testing.T) {
	root := t.TempDir()

	svc := ServiceSpec{
		Name:     "FancyThing",
		Label:    "FancyThing",
		Category: "not-a-real-category",
		Image:    "fancy:latest",
	}

	err := AppendService(root, svc)
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
}

// TestSlugify exercises the slugify helper directly.
func TestSlugify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"My Redis", "my-redis"},
		{"Hello World!", "hello-world"},
		{"PostgreSQL 16", "postgresql-16"},
		{"already-slug", "already-slug"},
		{"  spaces  ", "--spaces--"},
	}
	for _, tc := range cases {
		got := slugify(tc.input)
		if got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
