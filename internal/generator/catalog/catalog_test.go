package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendService_HappyPath(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	initial := `{"services":[{"id":"redis","name":"Redis","label":"Redis","category":"cache","selectable":true}]}`
	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial catalog: %v", err)
	}

	svc := ServiceSpec{
		Name:     "My Postgres",
		Label:    "My Postgres",
		Image:    "postgres:16",
		Category: "database",
	}
	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "services.json"))
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	var catalog ServiceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(catalog.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(catalog.Services))
	}
	added := catalog.Services[1]
	if added.ID != "my-postgres" {
		t.Errorf("expected id my-postgres, got %q", added.ID)
	}
	if !added.Selectable {
		t.Error("expected selectable=true")
	}
	if added.Public {
		t.Error("expected public=false")
	}
	if added.Order != 100 {
		t.Errorf("expected order=100, got %d", added.Order)
	}
}

func TestAppendService_IDCollisionResolution(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	initial := `{"services":[{"id":"my-redis","name":"Redis","label":"Redis","category":"cache","selectable":true}]}`
	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial catalog: %v", err)
	}

	svc := ServiceSpec{
		Name:     "My Redis",
		Image:    "redis:7",
		Category: "cache",
	}
	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(configDir, "services.json"))
	var catalog ServiceCatalog
	_ = json.Unmarshal(data, &catalog)

	if len(catalog.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(catalog.Services))
	}
	if catalog.Services[1].ID != "my-redis-2" {
		t.Errorf("expected id my-redis-2, got %q", catalog.Services[1].ID)
	}
}

func TestAppendService_CreatesMissingConfigDir(t *testing.T) {
	root := t.TempDir()
	// No config/ directory created

	svc := ServiceSpec{
		Name:     "Kafka",
		Image:    "confluentinc/cp-kafka:7",
		Category: "message-queue",
	}
	if err := AppendService(root, svc); err != nil {
		t.Fatalf("AppendService: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "config", "services.json"))
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	var catalog ServiceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(catalog.Services) != 1 || catalog.Services[0].ID != "kafka" {
		t.Errorf("unexpected catalog: %+v", catalog.Services)
	}
}

func TestAppendService_InvalidCategoryRejected(t *testing.T) {
	root := t.TempDir()

	svc := ServiceSpec{
		Name:     "Bad",
		Image:    "bad:latest",
		Category: "invalid-category",
	}
	err := AppendService(root, svc)
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
}
