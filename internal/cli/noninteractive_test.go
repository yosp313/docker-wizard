package cli

import (
	"os"
	"path/filepath"
	"testing"

	"docker-wizard/internal/generator"
)

func TestParseLanguageOption(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      generator.Language
		wantSet   bool
		wantError bool
	}{
		{name: "auto default", input: "", wantSet: false},
		{name: "auto explicit", input: "auto", wantSet: false},
		{name: "go", input: "go", want: generator.LanguageGo, wantSet: true},
		{name: "dotnet alias", input: ".NET", want: generator.LanguageDotNet, wantSet: true},
		{name: "invalid", input: "rust", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, set, err := parseLanguageOption(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected language %q, got %q", tt.want, got)
			}
			if set != tt.wantSet {
				t.Fatalf("expected set %v, got %v", tt.wantSet, set)
			}
		})
	}
}

func TestResolveServices(t *testing.T) {
	root := t.TempDir()
	writeServicesCatalog(t, root)

	t.Run("all selects selectable services", func(t *testing.T) {
		got, err := resolveServices(root, []string{"all"})
		if err != nil {
			t.Fatalf("resolve services: %v", err)
		}
		want := []string{"mysql", "redis"}
		if len(got) != len(want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("expected %v, got %v", want, got)
			}
		}
	})

	t.Run("custom list is deduped and ordered", func(t *testing.T) {
		got, err := resolveServices(root, []string{"redis", "mysql", "redis"})
		if err != nil {
			t.Fatalf("resolve services: %v", err)
		}
		want := []string{"mysql", "redis"}
		if len(got) != len(want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("expected %v, got %v", want, got)
			}
		}
	})

	t.Run("unknown service returns error", func(t *testing.T) {
		_, err := resolveServices(root, []string{"mysql", "unknown"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func writeServicesCatalog(t *testing.T, root string) {
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
      "id": "plausible-db",
      "label": "Plausible DB",
      "category": "analytics",
      "image": "clickhouse/clickhouse-server:24.3",
      "selectable": false,
      "order": 30
    }
  ]
}`

	if err := os.WriteFile(filepath.Join(configDir, "services.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write services catalog: %v", err)
	}
}
