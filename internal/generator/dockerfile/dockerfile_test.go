package dockerfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerfileNodeUsesNpmCIWithPackageLock(t *testing.T) {
	root := t.TempDir()
	writeDockerfileCatalog(t, root)

	details := LanguageDetails{Type: LanguageNode, HasPackageLock: true}
	content, err := Dockerfile(root, details)
	if err != nil {
		t.Fatalf("dockerfile: %v", err)
	}

	if !strings.Contains(content, "RUN npm ci") {
		t.Fatalf("expected npm ci install command")
	}
	if !strings.Contains(content, "ENV APP_START_CMD=\"npm start\"") {
		t.Fatalf("expected APP_START_CMD for node")
	}
	if !strings.Contains(content, "CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]") {
		t.Fatalf("expected shell-based command fallback")
	}
}

func TestDockerfileNodeUsesNpmInstallWithoutLockfile(t *testing.T) {
	root := t.TempDir()
	writeDockerfileCatalog(t, root)

	details := LanguageDetails{Type: LanguageNode}
	content, err := Dockerfile(root, details)
	if err != nil {
		t.Fatalf("dockerfile: %v", err)
	}

	if !strings.Contains(content, "RUN npm install") {
		t.Fatalf("expected npm install command")
	}
}

func TestDockerfileJavaUsesMavenMultiStage(t *testing.T) {
	root := t.TempDir()
	writeDockerfileCatalog(t, root)

	details := LanguageDetails{Type: LanguageJava, HasPomXML: true, JavaVersion: "21"}
	content, err := Dockerfile(root, details)
	if err != nil {
		t.Fatalf("dockerfile: %v", err)
	}

	if !strings.Contains(content, "FROM maven:3.9-eclipse-temurin-21 AS build") {
		t.Fatalf("expected maven build stage")
	}
	if !strings.Contains(content, "COPY --from=build /out/app.jar /app/app.jar") {
		t.Fatalf("expected runtime copy from build stage")
	}
	if !strings.Contains(content, "ENV APP_START_CMD=\"java -jar /app/app.jar\"") {
		t.Fatalf("expected APP_START_CMD for java")
	}
}

func TestDockerfileDotNetUsesMultiStageAndProjectDLL(t *testing.T) {
	root := t.TempDir()
	writeDockerfileCatalog(t, root)

	details := LanguageDetails{Type: LanguageDotNet, DotNetVersion: "8.0", DotNetProject: "Service.Api"}
	content, err := Dockerfile(root, details)
	if err != nil {
		t.Fatalf("dockerfile: %v", err)
	}

	if !strings.Contains(content, "FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build") {
		t.Fatalf("expected sdk build stage")
	}
	if !strings.Contains(content, "FROM mcr.microsoft.com/dotnet/aspnet:8.0") {
		t.Fatalf("expected aspnet runtime stage")
	}
	if !strings.Contains(content, "ENV APP_START_CMD=\"dotnet /app/Service.Api.dll\"") {
		t.Fatalf("expected project dll start command")
	}
}

func TestDockerfileReturnsErrorWhenTemplateMissing(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	content := `{"dockerfiles":[{"language":"node","templateLines":["FROM node:20-alpine"]}]}`
	if err := os.WriteFile(filepath.Join(configDir, "dockerfiles.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write dockerfile catalog: %v", err)
	}

	_, err := Dockerfile(root, LanguageDetails{Type: LanguagePython})
	if err == nil {
		t.Fatal("expected missing template error")
	}
	if !strings.Contains(err.Error(), "missing dockerfile template") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeDockerfileCatalog(t *testing.T, root string) {
	t.Helper()
	configDir := filepath.Join(root, "config")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	catalogPath := filepath.Join("..", "..", "..", "config", "dockerfiles.json")
	catalogData, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("read default dockerfile catalog: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "dockerfiles.json"), catalogData, 0o644); err != nil {
		t.Fatalf("write dockerfile catalog: %v", err)
	}
}
