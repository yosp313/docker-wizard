package dockerfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var versionPattern = regexp.MustCompile(`\d+(?:\.\d+){0,2}`)

func detectGoVersion(root string) string {
	content := readFile(filepath.Join(root, "go.mod"))
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return normalizeMajorMinor(strings.TrimSpace(strings.TrimPrefix(line, "go ")))
		}
	}
	return ""
}

func detectNodeVersion(root string) string {
	if version := firstToken(readFile(filepath.Join(root, ".nvmrc"))); version != "" {
		return normalizeMajor(version)
	}
	if version := firstToken(readFile(filepath.Join(root, ".node-version"))); version != "" {
		return normalizeMajor(version)
	}

	content := readFile(filepath.Join(root, "package.json"))
	if content == "" {
		return ""
	}
	var parsed struct {
		Engines map[string]string `json:"engines"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ""
	}
	if parsed.Engines == nil {
		return ""
	}
	return normalizeMajor(parsed.Engines["node"])
}

func detectPythonVersion(root string) string {
	if version := firstToken(readFile(filepath.Join(root, ".python-version"))); version != "" {
		return normalizeMajorMinor(version)
	}
	if version := pythonFromRuntime(readFile(filepath.Join(root, "runtime.txt"))); version != "" {
		return normalizeMajorMinor(version)
	}

	content := readFile(filepath.Join(root, "pyproject.toml"))
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "requires-python") {
			return normalizeMajorMinor(extractVersion(line))
		}
	}

	return ""
}

func detectRubyVersion(root string) string {
	return normalizeMajorMinor(firstToken(readFile(filepath.Join(root, ".ruby-version"))))
}

func detectJavaVersion(root string) string {
	if version := javaFromPom(readFile(filepath.Join(root, "pom.xml"))); version != "" {
		return normalizeMajor(version)
	}
	if version := javaFromGradle(readFile(filepath.Join(root, "build.gradle"))); version != "" {
		return normalizeMajor(version)
	}
	if version := javaFromGradle(readFile(filepath.Join(root, "build.gradle.kts"))); version != "" {
		return normalizeMajor(version)
	}
	return ""
}

func detectDotNetVersion(root string) string {
	content := readFile(filepath.Join(root, "global.json"))
	if content == "" {
		return ""
	}
	var parsed struct {
		SDK struct {
			Version string `json:"version"`
		} `json:"sdk"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ""
	}
	return normalizeMajorMinor(parsed.SDK.Version)
}

func detectPHPVersion(root string) string {
	if version := firstToken(readFile(filepath.Join(root, ".php-version"))); version != "" {
		return normalizeMajorMinor(version)
	}
	content := readFile(filepath.Join(root, "composer.json"))
	if content == "" {
		return ""
	}
	var parsed struct {
		Config struct {
			Platform map[string]string `json:"platform"`
		} `json:"config"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ""
	}
	if parsed.Config.Platform == nil {
		return ""
	}
	return normalizeMajorMinor(parsed.Config.Platform["php"])
}

func javaFromPom(content string) string {
	if content == "" {
		return ""
	}
	for _, key := range []string{"maven.compiler.release", "maven.compiler.source", "java.version"} {
		if version := betweenTags(content, key); version != "" {
			return version
		}
	}
	return ""
}

func javaFromGradle(content string) string {
	if content == "" {
		return ""
	}
	if version := extractVersionFromPattern(content, `JavaLanguageVersion\.of\((\d+)\)`); version != "" {
		return version
	}
	if version := extractVersionFromPattern(content, `sourceCompatibility\s*=\s*["']?(\d+)`); version != "" {
		return version
	}
	if version := extractVersionFromPattern(content, `targetCompatibility\s*=\s*["']?(\d+)`); version != "" {
		return version
	}
	if version := extractVersionFromPattern(content, `JavaVersion\.VERSION_(\d+)`); version != "" {
		return version
	}
	return ""
}

func pythonFromRuntime(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	content = strings.TrimPrefix(content, "python-")
	return extractVersion(content)
}

func betweenTags(content string, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(content, open)
	if start == -1 {
		return ""
	}
	start += len(open)
	end := strings.Index(content[start:], close)
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(content[start : start+end])
}

func extractVersion(value string) string {
	match := versionPattern.FindString(value)
	return match
}

func extractVersionFromPattern(content string, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func normalizeMajor(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(value, "v"))
	version := extractVersion(value)
	if version == "" {
		return ""
	}
	parts := strings.Split(version, ".")
	return parts[0]
}

func normalizeMajorMinor(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(value, "v"))
	version := extractVersion(value)
	if version == "" {
		return ""
	}
	parts := strings.Split(version, ".")
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + "." + parts[1]
}

func firstToken(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
