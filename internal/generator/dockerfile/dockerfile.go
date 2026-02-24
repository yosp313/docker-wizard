package dockerfile

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func Dockerfile(root string, details LanguageDetails) (string, error) {
	templates, err := loadTemplates(root)
	if err != nil {
		return "", err
	}

	language := details.Type
	if language == "" {
		language = LanguageUnknown
	}

	templateLines, ok := templates[language]
	if !ok {
		return "", fmt.Errorf("missing dockerfile template for language: %s", language)
	}

	content, err := renderTemplateLines(templateLines, templateDataFromDetails(details))
	if err != nil {
		return "", fmt.Errorf("render dockerfile template for %s: %w", language, err)
	}

	return content, nil
}

type templateData struct {
	HasGoSum           bool
	GoVersion          string
	HasPackageLock     bool
	HasYarnLock        bool
	HasPnpmLock        bool
	NodeVersion        string
	NodeInstallCommand string
	NodeStartCommand   string
	HasRequirements    bool
	PythonVersion      string
	HasGemfile         bool
	HasGemfileLock     bool
	RubyVersion        string
	HasComposerJSON    bool
	PHPVersion         string
	HasPomXML          bool
	HasGradle          bool
	HasGradleKts       bool
	JavaVersion        string
	DotNetVersion      string
	DotNetEntryDLL     string
}

func versionOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func templateDataFromDetails(details LanguageDetails) templateData {
	entryDLL := "app.dll"
	if details.DotNetProject != "" {
		entryDLL = details.DotNetProject + ".dll"
	}

	return templateData{
		HasGoSum:           details.HasGoSum,
		GoVersion:          versionOrDefault(details.GoVersion, "1.25"),
		HasPackageLock:     details.HasPackageLock,
		HasYarnLock:        details.HasYarnLock,
		HasPnpmLock:        details.HasPnpmLock,
		NodeVersion:        versionOrDefault(details.NodeVersion, "20"),
		NodeInstallCommand: nodeInstallCommand(details),
		NodeStartCommand:   nodeStartCommand(details),
		HasRequirements:    details.HasRequirements,
		PythonVersion:      versionOrDefault(details.PythonVersion, "3.12"),
		HasGemfile:         details.HasGemfile,
		HasGemfileLock:     details.HasGemfileLock,
		RubyVersion:        versionOrDefault(details.RubyVersion, "3.3"),
		HasComposerJSON:    details.HasComposerJSON,
		PHPVersion:         versionOrDefault(details.PHPVersion, "8.3"),
		HasPomXML:          details.HasPomXML,
		HasGradle:          details.HasGradle,
		HasGradleKts:       details.HasGradleKts,
		JavaVersion:        versionOrDefault(details.JavaVersion, "21"),
		DotNetVersion:      versionOrDefault(details.DotNetVersion, "8.0"),
		DotNetEntryDLL:     entryDLL,
	}
}

func nodeInstallCommand(details LanguageDetails) string {
	if details.HasYarnLock {
		return "yarn install --frozen-lockfile"
	}
	if details.HasPnpmLock {
		return "pnpm install --frozen-lockfile"
	}
	if details.HasPackageLock {
		return "npm ci"
	}
	return "npm install"
}

func nodeStartCommand(details LanguageDetails) string {
	if details.HasYarnLock {
		return "yarn start"
	}
	if details.HasPnpmLock {
		return "pnpm start"
	}
	return "npm start"
}

func renderTemplateLines(lines []string, data templateData) (string, error) {
	if len(lines) == 0 {
		return "", fmt.Errorf("template is empty")
	}

	renderedLines := make([]string, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			renderedLines = append(renderedLines, "")
			continue
		}

		tpl, err := template.New("dockerfile-line").Option("missingkey=error").Parse(line)
		if err != nil {
			return "", fmt.Errorf("line %d: %w", i+1, err)
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("line %d: %w", i+1, err)
		}

		rendered := strings.TrimSuffix(buf.String(), "\n")
		if rendered == "" {
			continue
		}

		renderedLines = append(renderedLines, strings.Split(rendered, "\n")...)
	}

	if len(renderedLines) == 0 {
		return "", fmt.Errorf("template rendered empty")
	}

	return strings.Join(renderedLines, "\n") + "\n", nil
}
