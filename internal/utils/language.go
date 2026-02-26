package utils

import "strings"

type LanguageVersions struct {
	Go     string
	Node   string
	Python string
	Ruby   string
	PHP    string
	Java   string
	DotNet string
}

func LanguageLabel(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "go":
		return "Go"
	case "node":
		return "Node"
	case "python":
		return "Python"
	case "ruby":
		return "Ruby"
	case "php":
		return "PHP"
	case "java":
		return "Java"
	case "dotnet", ".net":
		return ".NET"
	default:
		return "Unknown"
	}
}

func LanguageVersion(language string, versions LanguageVersions) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "go":
		return versions.Go
	case "node":
		return versions.Node
	case "python":
		return versions.Python
	case "ruby":
		return versions.Ruby
	case "php":
		return versions.PHP
	case "java":
		return versions.Java
	case "dotnet", ".net":
		return versions.DotNet
	default:
		return ""
	}
}

func LanguageLabelWithVersion(language string, versions LanguageVersions) string {
	label := LanguageLabel(language)
	version := LanguageVersion(language, versions)
	if version == "" {
		return label
	}
	return label + " " + version
}
