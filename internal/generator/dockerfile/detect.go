package dockerfile

import (
	"fmt"
	"path/filepath"

	"docker-wizard/internal/utils"
)

type Language string

const (
	LanguageGo      Language = "go"
	LanguageNode    Language = "node"
	LanguagePython  Language = "python"
	LanguageRuby    Language = "ruby"
	LanguagePHP     Language = "php"
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
	HasComposerJSON bool
	HasPHPVersion   bool
	HasPomXML       bool
	HasGradle       bool
	HasGradleKts    bool
	HasCSProj       bool
	GoVersion       string
	NodeVersion     string
	PythonVersion   string
	RubyVersion     string
	PHPVersion      string
	JavaVersion     string
	DotNetVersion   string
	DotNetProject   string
}

func DetectLanguage(root string) (LanguageDetails, error) {
	if root == "" {
		return LanguageDetails{}, fmt.Errorf("root directory is required")
	}

	details := LanguageDetails{}

	details.HasGoMod = utils.FileExists(filepath.Join(root, "go.mod"))
	details.HasGoSum = utils.FileExists(filepath.Join(root, "go.sum"))

	details.HasPackageJSON = utils.FileExists(filepath.Join(root, "package.json"))
	details.HasPackageLock = utils.FileExists(filepath.Join(root, "package-lock.json"))
	details.HasYarnLock = utils.FileExists(filepath.Join(root, "yarn.lock"))
	details.HasPnpmLock = utils.FileExists(filepath.Join(root, "pnpm-lock.yaml"))

	details.HasRequirements = utils.FileExists(filepath.Join(root, "requirements.txt"))
	details.HasPyProject = utils.FileExists(filepath.Join(root, "pyproject.toml"))
	details.HasPipfile = utils.FileExists(filepath.Join(root, "Pipfile"))

	details.HasGemfile = utils.FileExists(filepath.Join(root, "Gemfile"))
	details.HasGemfileLock = utils.FileExists(filepath.Join(root, "Gemfile.lock"))
	details.HasComposerJSON = utils.FileExists(filepath.Join(root, "composer.json"))
	details.HasPHPVersion = utils.FileExists(filepath.Join(root, ".php-version"))

	details.HasPomXML = utils.FileExists(filepath.Join(root, "pom.xml"))
	details.HasGradle = utils.FileExists(filepath.Join(root, "build.gradle"))
	details.HasGradleKts = utils.FileExists(filepath.Join(root, "build.gradle.kts"))

	csproj, err := filepath.Glob(filepath.Join(root, "*.csproj"))
	if err != nil {
		return LanguageDetails{}, fmt.Errorf("detect .csproj: %w", err)
	}
	details.HasCSProj = len(csproj) > 0
	if len(csproj) > 0 {
		details.DotNetProject = trimExt(filepath.Base(csproj[0]))
	}

	details.GoVersion = detectGoVersion(root)
	details.NodeVersion = detectNodeVersion(root)
	details.PythonVersion = detectPythonVersion(root)
	details.RubyVersion = detectRubyVersion(root)
	details.PHPVersion = detectPHPVersion(root)
	details.JavaVersion = detectJavaVersion(root)
	details.DotNetVersion = detectDotNetVersion(root)

	switch {
	case details.HasGoMod:
		details.Type = LanguageGo
	case details.HasPackageJSON:
		details.Type = LanguageNode
	case details.HasRequirements || details.HasPyProject || details.HasPipfile:
		details.Type = LanguagePython
	case details.HasGemfile:
		details.Type = LanguageRuby
	case details.HasComposerJSON || details.HasPHPVersion:
		details.Type = LanguagePHP
	case details.HasPomXML || details.HasGradle || details.HasGradleKts:
		details.Type = LanguageJava
	case details.HasCSProj:
		details.Type = LanguageDotNet
	default:
		details.Type = LanguageUnknown
	}

	return details, nil
}

func trimExt(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return name[:len(name)-len(ext)]
}
