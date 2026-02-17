package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

type Language string

const (
	LanguageGo      Language = "go"
	LanguageNode    Language = "node"
	LanguagePython  Language = "python"
	LanguageRuby    Language = "ruby"
	LanguageJava    Language = "java"
	LanguageDotNet  Language = "dotnet"
	LanguageUnknown Language = "unknown"
)

type LanguageDetails struct {
	Type            Language
	HasGoMod        bool
	HasGoSum        bool
	HasPackageJSON  bool
	HasPackageLock  bool
	HasYarnLock     bool
	HasPnpmLock     bool
	HasRequirements bool
	HasPyProject    bool
	HasPipfile      bool
	HasGemfile      bool
	HasGemfileLock  bool
	HasPomXML       bool
	HasGradle       bool
	HasGradleKts    bool
	HasCSProj       bool
}

func DetectLanguage(root string) (LanguageDetails, error) {
	if root == "" {
		return LanguageDetails{}, fmt.Errorf("root directory is required")
	}

	details := LanguageDetails{}

	details.HasGoMod = fileExists(filepath.Join(root, "go.mod"))
	details.HasGoSum = fileExists(filepath.Join(root, "go.sum"))

	details.HasPackageJSON = fileExists(filepath.Join(root, "package.json"))
	details.HasPackageLock = fileExists(filepath.Join(root, "package-lock.json"))
	details.HasYarnLock = fileExists(filepath.Join(root, "yarn.lock"))
	details.HasPnpmLock = fileExists(filepath.Join(root, "pnpm-lock.yaml"))

	details.HasRequirements = fileExists(filepath.Join(root, "requirements.txt"))
	details.HasPyProject = fileExists(filepath.Join(root, "pyproject.toml"))
	details.HasPipfile = fileExists(filepath.Join(root, "Pipfile"))

	details.HasGemfile = fileExists(filepath.Join(root, "Gemfile"))
	details.HasGemfileLock = fileExists(filepath.Join(root, "Gemfile.lock"))

	details.HasPomXML = fileExists(filepath.Join(root, "pom.xml"))
	details.HasGradle = fileExists(filepath.Join(root, "build.gradle"))
	details.HasGradleKts = fileExists(filepath.Join(root, "build.gradle.kts"))

	csproj, err := filepath.Glob(filepath.Join(root, "*.csproj"))
	if err != nil {
		return LanguageDetails{}, fmt.Errorf("detect .csproj: %w", err)
	}
	details.HasCSProj = len(csproj) > 0

	switch {
	case details.HasGoMod:
		details.Type = LanguageGo
	case details.HasPackageJSON:
		details.Type = LanguageNode
	case details.HasRequirements || details.HasPyProject || details.HasPipfile:
		details.Type = LanguagePython
	case details.HasGemfile:
		details.Type = LanguageRuby
	case details.HasPomXML || details.HasGradle || details.HasGradleKts:
		details.Type = LanguageJava
	case details.HasCSProj:
		details.Type = LanguageDotNet
	default:
		details.Type = LanguageUnknown
	}

	return details, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
