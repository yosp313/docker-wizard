package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFilesSuccess(t *testing.T) {
	root := t.TempDir()
	compose := "version: \"3.9\"\nservices:\n"
	dockerfile := "FROM alpine:3.20\n"

	out, err := WriteFiles(root, compose, dockerfile)
	if err != nil {
		t.Fatalf("write files: %v", err)
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)

	if out.ComposePath != composePath {
		t.Fatalf("unexpected compose path: %s", out.ComposePath)
	}
	if out.DockerfilePath != dockerfilePath {
		t.Fatalf("unexpected dockerfile path: %s", out.DockerfilePath)
	}
	if out.DockerignorePath == "" {
		t.Fatal("expected dockerignore path to be set")
	}

	composeBytes, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	if string(composeBytes) != compose {
		t.Fatalf("unexpected compose content: %q", string(composeBytes))
	}

	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read dockerfile: %v", err)
	}
	if string(dockerfileBytes) != dockerfile {
		t.Fatalf("unexpected dockerfile content: %q", string(dockerfileBytes))
	}
}

func TestWriteFilesRollsBackComposeWhenDockerfileMoveFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, DockerfileFileName), 0o755); err != nil {
		t.Fatalf("create dockerfile directory: %v", err)
	}

	_, err := WriteFiles(root, "services:\n", "FROM alpine\n")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "move dockerfile into place") {
		t.Fatalf("unexpected error: %v", err)
	}

	_, statErr := os.Stat(filepath.Join(root, ComposeFileName))
	if !os.IsNotExist(statErr) {
		t.Fatalf("expected compose file to be rolled back, got stat err: %v", statErr)
	}

	info, statErr := os.Stat(filepath.Join(root, DockerfileFileName))
	if statErr != nil {
		t.Fatalf("expected dockerfile directory to remain: %v", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("expected dockerfile path to remain a directory")
	}
}
