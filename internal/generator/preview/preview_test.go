package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docker-wizard/internal/generator/write"
)

func TestPreviewFilesUsesComposeMergeOutput(t *testing.T) {
	root := t.TempDir()

	existingCompose := "" +
		"version: \"3.9\"\n" +
		"services:\n" +
		"  custom:\n" +
		"    image: busybox\n" +
		"x-local:\n" +
		"  note: keep\n"
	if err := os.WriteFile(filepath.Join(root, write.ComposeFileName), []byte(existingCompose), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}

	generatedCompose := "" +
		"version: \"3.9\"\n" +
		"services:\n" +
		"  app:\n" +
		"    image: app:latest\n" +
		"networks:\n" +
		"  app-net:\n"

	preview, err := PreviewFiles(root, generatedCompose, "FROM alpine:3.20\n")
	if err != nil {
		t.Fatalf("preview files: %v", err)
	}

	if preview.Compose.Status != FileStatusDifferent {
		t.Fatalf("expected compose status different, got %s", preview.Compose.Status)
	}
	if preview.Compose.Content == generatedCompose {
		t.Fatalf("expected preview compose content to reflect merged output, not raw generated")
	}
	if !strings.Contains(preview.Compose.Content, "custom:") {
		t.Fatalf("expected preview compose content to preserve existing custom service")
	}
}

func TestPreviewFilesMarksSameWhenMergedDockerfileUnchanged(t *testing.T) {
	root := t.TempDir()

	existingDockerfile := "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"CMD [\"./run.sh\"]\n" +
		"ENV APP_START_CMD=\"/app/app\"\n"
	if err := os.WriteFile(filepath.Join(root, write.DockerfileFileName), []byte(existingDockerfile), 0o644); err != nil {
		t.Fatalf("write dockerfile: %v", err)
	}

	generatedDockerfile := "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"ENV APP_START_CMD=\"/app/app\"\n" +
		"CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]\n"

	preview, err := PreviewFiles(root, "version: \"3.9\"\nservices:\n", generatedDockerfile)
	if err != nil {
		t.Fatalf("preview files: %v", err)
	}

	if preview.Dockerfile.Status != FileStatusSame {
		t.Fatalf("expected dockerfile status same after merge, got %s", preview.Dockerfile.Status)
	}
}
