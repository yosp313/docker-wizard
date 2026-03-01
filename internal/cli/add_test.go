package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdd(t *testing.T) {
	t.Run("add single service dry-run", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		err := RunAdd(root, AddOptions{Services: []string{"mysql"}, Write: false})
		if err != nil {
			t.Fatalf("RunAdd: %v", err)
		}

		// Verify docker-compose.yml does NOT exist in dry-run mode
		composePath := filepath.Join(root, "docker-compose.yml")
		if _, err := os.Stat(composePath); err == nil {
			t.Fatal("docker-compose.yml should not exist in dry-run mode")
		}
	})

	t.Run("add single service write", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		err := RunAdd(root, AddOptions{Services: []string{"mysql"}, Write: true})
		if err != nil {
			t.Fatalf("RunAdd: %v", err)
		}

		// Verify docker-compose.yml exists
		composePath := filepath.Join(root, "docker-compose.yml")
		data, err := os.ReadFile(composePath)
		if err != nil {
			t.Fatalf("read docker-compose.yml: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "mysql:") {
			t.Fatal("compose file should contain mysql service")
		}
		if !strings.Contains(content, "app-net:") {
			t.Fatal("compose file should contain app-net network")
		}
		if strings.Contains(content, "  app:") {
			t.Fatal("compose file should NOT contain app service")
		}
	})

	t.Run("add to existing compose", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		// Write minimal existing compose with redis
		composePath := filepath.Join(root, "docker-compose.yml")
		existingCompose := `version: "3.9"
services:
  redis:
    image: redis:7-alpine
    networks:
      - app-net
networks:
  app-net:
`
		if err := os.WriteFile(composePath, []byte(existingCompose), 0o644); err != nil {
			t.Fatalf("write existing compose: %v", err)
		}

		err := RunAdd(root, AddOptions{Services: []string{"mysql"}, Write: true})
		if err != nil {
			t.Fatalf("RunAdd: %v", err)
		}

		// Verify compose file contains both services
		data, err := os.ReadFile(composePath)
		if err != nil {
			t.Fatalf("read docker-compose.yml: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "mysql:") {
			t.Fatal("compose file should contain mysql service")
		}
		if !strings.Contains(content, "redis:") {
			t.Fatal("compose file should contain redis service")
		}
	})

	t.Run("skip already-present service", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		// Write compose with mysql already present
		composePath := filepath.Join(root, "docker-compose.yml")
		existingCompose := `version: "3.9"
services:
  mysql:
    image: mysql:8.0
    networks:
      - app-net
networks:
  app-net:
`
		if err := os.WriteFile(composePath, []byte(existingCompose), 0o644); err != nil {
			t.Fatalf("write existing compose: %v", err)
		}

		err := RunAdd(root, AddOptions{Services: []string{"mysql"}, Write: true})
		if err != nil {
			t.Fatalf("RunAdd: %v", err)
		}
		// Should succeed (skip notice goes to stdout)
	})

	t.Run("unknown service returns error", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		err := RunAdd(root, AddOptions{Services: []string{"unknown"}})
		if err == nil {
			t.Fatal("expected error for unknown service")
		}
	})

	t.Run("empty services returns error", func(t *testing.T) {
		root := t.TempDir()
		writeServicesCatalog(t, root)

		err := RunAdd(root, AddOptions{Services: []string{}})
		if err == nil {
			t.Fatal("expected error for empty services")
		}
	})
}
