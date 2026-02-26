package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFilesCreatesMissingFiles(t *testing.T) {
	root := t.TempDir()
	compose := "version: \"3.9\"\nservices:\n"
	dockerfile := "FROM alpine:3.20\n"

	out, err := WriteFiles(root, compose, dockerfile)
	if err != nil {
		t.Fatalf("write files: %v", err)
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)
	dockerignorePath := filepath.Join(root, DockerignoreFileName)

	if out.ComposePath != composePath {
		t.Fatalf("unexpected compose path: %s", out.ComposePath)
	}
	if out.DockerfilePath != dockerfilePath {
		t.Fatalf("unexpected dockerfile path: %s", out.DockerfilePath)
	}
	if out.DockerignorePath != dockerignorePath {
		t.Fatalf("unexpected dockerignore path: %s", out.DockerignorePath)
	}

	if out.ComposeStatus != WriteStatusCreated {
		t.Fatalf("expected compose status created, got %s", out.ComposeStatus)
	}
	if out.DockerfileStatus != WriteStatusCreated {
		t.Fatalf("expected dockerfile status created, got %s", out.DockerfileStatus)
	}
	if out.DockerignoreStatus != WriteStatusCreated {
		t.Fatalf("expected dockerignore status created, got %s", out.DockerignoreStatus)
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

func TestWriteFilesMergesDifferingExistingFiles(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)
	existingCompose := "" +
		"version: \"3.9\"\n" +
		"x-local: keep\n" +
		"services:\n" +
		"  custom:\n" +
		"    image: busybox\n" +
		"    command:\n" +
		"      - sleep\n" +
		"      - \"3600\"\n"
	if err := os.WriteFile(composePath, []byte(existingCompose), 0o644); err != nil {
		t.Fatalf("write existing compose: %v", err)
	}

	existingDockerfile := "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"CMD [\"./run.sh\"]\n"
	if err := os.WriteFile(dockerfilePath, []byte(existingDockerfile), 0o644); err != nil {
		t.Fatalf("write existing dockerfile: %v", err)
	}

	generatedCompose := "" +
		"version: \"3.9\"\n" +
		"services:\n" +
		"  app:\n" +
		"    image: app:latest\n" +
		"networks:\n" +
		"  app-net:\n"
	generatedDockerfile := "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"ENV APP_START_CMD=\"/app/app\"\n" +
		"CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]\n"

	out, err := WriteFiles(root, generatedCompose, generatedDockerfile)
	if err != nil {
		t.Fatalf("write files: %v", err)
	}

	if out.ComposeStatus != WriteStatusUpdated {
		t.Fatalf("expected compose updated, got %s", out.ComposeStatus)
	}
	if out.DockerfileStatus != WriteStatusUpdated {
		t.Fatalf("expected dockerfile updated, got %s", out.DockerfileStatus)
	}
	if out.ComposeBackupPath == "" {
		t.Fatal("expected compose backup path")
	}
	if out.DockerfileBackupPath == "" {
		t.Fatal("expected dockerfile backup path")
	}

	composeBytes, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	mergedCompose := string(composeBytes)
	if !strings.Contains(mergedCompose, "custom:") {
		t.Fatalf("expected merged compose to preserve existing custom service")
	}
	if !strings.Contains(mergedCompose, "app:") {
		t.Fatalf("expected merged compose to add generated app service")
	}
	if !strings.Contains(mergedCompose, "app-net:") {
		t.Fatalf("expected merged compose to add generated network")
	}
	if !strings.Contains(mergedCompose, "x-local: keep") {
		t.Fatalf("expected merged compose to preserve unmanaged root keys")
	}

	dockerfileBytes, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read dockerfile: %v", err)
	}
	mergedDockerfile := string(dockerfileBytes)
	if !strings.Contains(mergedDockerfile, "CMD [\"./run.sh\"]") {
		t.Fatalf("expected merged Dockerfile to keep existing command")
	}
	if !strings.Contains(mergedDockerfile, "ENV APP_START_CMD=") {
		t.Fatalf("expected merged Dockerfile to include APP_START_CMD")
	}
	if strings.Contains(mergedDockerfile, "CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]") {
		t.Fatalf("expected merged Dockerfile to preserve existing CMD without adding another CMD")
	}

	composeBackupBytes, err := os.ReadFile(out.ComposeBackupPath)
	if err != nil {
		t.Fatalf("read compose backup: %v", err)
	}
	if string(composeBackupBytes) != existingCompose {
		t.Fatalf("expected compose backup to contain original content")
	}

	dockerfileBackupBytes, err := os.ReadFile(out.DockerfileBackupPath)
	if err != nil {
		t.Fatalf("read dockerfile backup: %v", err)
	}
	if string(dockerfileBackupBytes) != existingDockerfile {
		t.Fatalf("expected dockerfile backup to contain original content")
	}
}

func TestWriteFilesMarksUnchangedFiles(t *testing.T) {
	root := t.TempDir()
	compose := "services:\n"
	dockerfile := "FROM alpine\n"
	if err := os.WriteFile(filepath.Join(root, ComposeFileName), []byte(compose), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, DockerfileFileName), []byte(dockerfile), 0o644); err != nil {
		t.Fatalf("write dockerfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, DockerignoreFileName), []byte(DefaultDockerignore()), 0o644); err != nil {
		t.Fatalf("write dockerignore: %v", err)
	}

	out, err := WriteFiles(root, compose, dockerfile)
	if err != nil {
		t.Fatalf("write files: %v", err)
	}

	if out.ComposeStatus != WriteStatusUnchanged {
		t.Fatalf("expected compose unchanged, got %s", out.ComposeStatus)
	}
	if out.DockerfileStatus != WriteStatusUnchanged {
		t.Fatalf("expected dockerfile unchanged, got %s", out.DockerfileStatus)
	}
	if out.DockerignoreStatus != WriteStatusUnchanged {
		t.Fatalf("expected dockerignore unchanged, got %s", out.DockerignoreStatus)
	}
}

func TestMergeComposePreservesExistingSections(t *testing.T) {
	existing := "" +
		"version: \"3.9\"\n" +
		"services:\n" +
		"  app:\n" +
		"    ports:\n" +
		"      - \"9999:8080\"\n" +
		"  custom:\n" +
		"    image: busybox\n" +
		"x-local:\n" +
		"  note: keep\n"

	generated := "" +
		"version: \"3.9\"\n" +
		"services:\n" +
		"  app:\n" +
		"    ports:\n" +
		"      - \"8080:8080\"\n" +
		"networks:\n" +
		"  app-net:\n"

	merged, err := MergeCompose(existing, generated)
	if err != nil {
		t.Fatalf("merge compose: %v", err)
	}

	if !strings.Contains(merged, "custom:") {
		t.Fatalf("expected merged compose to preserve custom service")
	}
	if !strings.Contains(merged, "9999:8080") {
		t.Fatalf("expected merged compose to preserve existing app ports")
	}
	if !strings.Contains(merged, "8080:8080") {
		t.Fatalf("expected merged compose to add generated app ports")
	}
	if !strings.Contains(merged, "x-local:") {
		t.Fatalf("expected merged compose to preserve unmanaged keys")
	}
	if !strings.Contains(merged, "app-net:") {
		t.Fatalf("expected merged compose to include generated network")
	}
}
