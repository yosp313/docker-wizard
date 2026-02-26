package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docker-wizard/internal/generator/compose"
)

func TestSelectionWarningsIncludesInsecureDefaults(t *testing.T) {
	root := t.TempDir()
	writeServicesCatalog(t, root, `{
  "services": [
    {
      "id": "traefik",
      "label": "Traefik",
      "category": "proxy",
      "image": "traefik:v2.11",
      "selectable": true,
      "command": ["--api.insecure=true"],
      "order": 10
    },
    {
      "id": "db",
      "label": "Database",
      "category": "database",
      "image": "postgres:16",
      "selectable": true,
      "env": ["POSTGRES_PASSWORD=example"],
      "order": 20
    }
  ]
}`)

	warnings, err := SelectionWarnings(root, compose.ComposeSelection{Services: []string{"traefik", "db"}})
	if err != nil {
		t.Fatalf("selection warnings: %v", err)
	}

	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "Traefik enables an insecure admin API flag") {
		t.Fatalf("expected insecure command warning, got: %v", warnings)
	}
	if !strings.Contains(joined, "Database includes placeholder environment defaults") {
		t.Fatalf("expected insecure env warning, got: %v", warnings)
	}
}

func TestSelectionWarningsOmitsInsecureDefaultsForSafeValues(t *testing.T) {
	root := t.TempDir()
	writeServicesCatalog(t, root, `{
  "services": [
    {
      "id": "traefik",
      "label": "Traefik",
      "category": "proxy",
      "image": "traefik:v2.11",
      "selectable": true,
      "command": ["--api.dashboard=true"],
      "order": 10
    },
    {
      "id": "db",
      "label": "Database",
      "category": "database",
      "image": "postgres:16",
      "selectable": true,
      "env": ["POSTGRES_PASSWORD=s3cur3-value"],
      "order": 20
    }
  ]
}`)

	warnings, err := SelectionWarnings(root, compose.ComposeSelection{Services: []string{"traefik", "db"}})
	if err != nil {
		t.Fatalf("selection warnings: %v", err)
	}

	joined := strings.Join(warnings, "\n")
	if strings.Contains(joined, "includes placeholder environment defaults") {
		t.Fatalf("did not expect placeholder env warning, got: %v", warnings)
	}
	if strings.Contains(joined, "insecure admin API flag") {
		t.Fatalf("did not expect insecure command warning, got: %v", warnings)
	}
}

func TestSelectionWarningsDoesNotFlagSinglePartPortsAsCollision(t *testing.T) {
	root := t.TempDir()
	writeServicesCatalog(t, root, `{
  "services": [
    {
      "id": "svc-a",
      "label": "Service A",
      "category": "proxy",
      "image": "busybox",
      "ports": ["8080"],
      "public": true,
      "selectable": true,
      "order": 10
    },
    {
      "id": "svc-b",
      "label": "Service B",
      "category": "analytics",
      "image": "busybox",
      "ports": ["8080"],
      "public": true,
      "selectable": true,
      "order": 20
    }
  ]
}`)

	warnings, err := SelectionWarnings(root, compose.ComposeSelection{Services: []string{"svc-a", "svc-b"}})
	if err != nil {
		t.Fatalf("selection warnings: %v", err)
	}

	for _, warning := range warnings {
		if strings.Contains(warning, "host port") {
			t.Fatalf("did not expect host port collision warning for single-part ports: %v", warnings)
		}
	}
}

func writeServicesCatalog(t *testing.T, root string, content string) {
	t.Helper()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write services catalog: %v", err)
	}
}
