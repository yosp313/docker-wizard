package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docker-wizard/internal/generator/catalog"
)

func TestWriteServiceRendersExposeForInternalServicesWithoutPorts(t *testing.T) {
	b := &strings.Builder{}
	writeService(b, catalog.ServiceSpec{
		ID:     "internalapi",
		Name:   "internalapi",
		Image:  "busybox",
		Public: false,
		Expose: []string{"9090"},
	})

	output := b.String()
	if !strings.Contains(output, "    expose:\n") {
		t.Fatalf("expected expose block in service output: %q", output)
	}
	if !strings.Contains(output, "      - \"9090\"\n") {
		t.Fatalf("expected expose port 9090 in service output: %q", output)
	}
	if strings.Contains(output, "    ports:\n") {
		t.Fatalf("did not expect published ports for internal service: %q", output)
	}
}

func writeTestCatalog(t *testing.T, root string) {
	t.Helper()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	content := `{
  "services": [
    {
      "id": "mysql",
      "label": "MySQL",
      "category": "database",
      "image": "mysql:8.0",
      "selectable": true,
      "order": 10
    },
    {
      "id": "redis",
      "label": "Redis",
      "category": "cache",
      "image": "redis:7-alpine",
      "selectable": true,
      "order": 20
    },
    {
      "id": "kafka",
      "label": "Kafka",
      "category": "message-queue",
      "image": "confluentinc/cp-kafka:7.5",
      "selectable": true,
      "order": 30,
      "requires": ["zookeeper"],
      "depends_on": ["zookeeper"]
    },
    {
      "id": "zookeeper",
      "label": "Zookeeper",
      "category": "message-queue",
      "image": "confluentinc/cp-zookeeper:7.5",
      "selectable": false,
      "order": 31
    }
  ]
}`

	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write services catalog: %v", err)
	}
}

func TestComposeFragment(t *testing.T) {
	root := t.TempDir()
	writeTestCatalog(t, root)

	t.Run("single service", func(t *testing.T) {
		output, expanded, err := ComposeFragment(root, []string{"mysql"})
		if err != nil {
			t.Fatalf("ComposeFragment: %v", err)
		}
		if expanded != nil && len(expanded) > 0 {
			t.Fatalf("expected no expanded services, got %v", expanded)
		}
		if !strings.Contains(output, "mysql:") {
			t.Fatalf("expected 'mysql:' service block in output")
		}
		if strings.Contains(output, "app:") {
			t.Fatalf("did not expect 'app:' service in fragment")
		}
		if !strings.Contains(output, "networks:") || !strings.Contains(output, "app-net:") {
			t.Fatalf("expected networks section with app-net")
		}
	})

	t.Run("multiple services", func(t *testing.T) {
		output, expanded, err := ComposeFragment(root, []string{"mysql", "redis"})
		if err != nil {
			t.Fatalf("ComposeFragment: %v", err)
		}
		if expanded != nil && len(expanded) > 0 {
			t.Fatalf("expected no expanded services, got %v", expanded)
		}
		if !strings.Contains(output, "mysql:") {
			t.Fatalf("expected 'mysql:' service block")
		}
		if !strings.Contains(output, "redis:") {
			t.Fatalf("expected 'redis:' service block")
		}
	})

	t.Run("auto-expands dependencies", func(t *testing.T) {
		output, expanded, err := ComposeFragment(root, []string{"kafka"})
		if err != nil {
			t.Fatalf("ComposeFragment: %v", err)
		}
		if len(expanded) != 1 || expanded[0] != "zookeeper" {
			t.Fatalf("expected expanded=['zookeeper'], got %v", expanded)
		}
		if !strings.Contains(output, "kafka:") {
			t.Fatalf("expected 'kafka:' service block")
		}
		if !strings.Contains(output, "zookeeper:") {
			t.Fatalf("expected 'zookeeper:' service block")
		}
	})

	t.Run("unknown service returns error", func(t *testing.T) {
		_, _, err := ComposeFragment(root, []string{"unknown"})
		if err == nil {
			t.Fatal("expected error for unknown service")
		}
	})

	t.Run("empty services returns error", func(t *testing.T) {
		_, _, err := ComposeFragment(root, []string{})
		if err == nil {
			t.Fatal("expected error for empty services")
		}
	})
}

func TestExistingComposeServices(t *testing.T) {
	t.Run("parses services", func(t *testing.T) {
		content := `version: "3.9"
services:
  mysql:
    image: mysql:8
  redis:
    image: redis:7
`
		result, err := ExistingComposeServices(content)
		if err != nil {
			t.Fatalf("ExistingComposeServices: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 services, got %d", len(result))
		}
		if !result["mysql"] {
			t.Fatal("expected mysql service")
		}
		if !result["redis"] {
			t.Fatal("expected redis service")
		}
	})

	t.Run("empty services", func(t *testing.T) {
		content := `version: "3.9"
services:
`
		result, err := ExistingComposeServices(content)
		if err != nil {
			t.Fatalf("ExistingComposeServices: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected empty map, got %d services", len(result))
		}
	})

	t.Run("invalid yaml returns error", func(t *testing.T) {
		content := `{{invalid`
		_, err := ExistingComposeServices(content)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})
}
