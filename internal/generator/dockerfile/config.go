package dockerfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type templateCatalog struct {
	Dockerfiles []templateSpec `json:"dockerfiles"`
}

type templateSpec struct {
	Language      string   `json:"language"`
	TemplateLines []string `json:"templateLines"`
}

func loadTemplates(root string) (map[Language][]string, error) {
	if root == "" {
		return nil, fmt.Errorf("root directory is required")
	}

	data, err := readTemplateData(root)
	if err != nil {
		return nil, err
	}

	var catalog templateCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parse dockerfile catalog: %w", err)
	}

	templates, err := normalizeTemplateCatalog(catalog)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

func readTemplateData(root string) ([]byte, error) {
	primary := filepath.Join(root, "config", "dockerfiles.json")
	data, err := os.ReadFile(primary)
	if err == nil {
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read dockerfile catalog: %w", err)
	}

	exe, exeErr := os.Executable()
	if exeErr != nil {
		return nil, fmt.Errorf("read dockerfile catalog: %w", exeErr)
	}
	secondary := filepath.Join(filepath.Dir(exe), "config", "dockerfiles.json")
	data, err = os.ReadFile(secondary)
	if err != nil {
		return nil, fmt.Errorf("read dockerfile catalog: %w", err)
	}

	return data, nil
}

func normalizeTemplateCatalog(catalog templateCatalog) (map[Language][]string, error) {
	if len(catalog.Dockerfiles) == 0 {
		return nil, fmt.Errorf("dockerfile catalog has no templates")
	}

	templates := make(map[Language][]string, len(catalog.Dockerfiles))
	for _, spec := range catalog.Dockerfiles {
		lang, ok := parseLanguage(spec.Language)
		if !ok {
			return nil, fmt.Errorf("invalid dockerfile language: %s", spec.Language)
		}
		if _, exists := templates[lang]; exists {
			return nil, fmt.Errorf("duplicate dockerfile language: %s", lang)
		}
		if len(spec.TemplateLines) == 0 {
			return nil, fmt.Errorf("dockerfile template for %s has no lines", lang)
		}

		hasContent := false
		for _, line := range spec.TemplateLines {
			if strings.TrimSpace(line) != "" {
				hasContent = true
				break
			}
		}
		if !hasContent {
			return nil, fmt.Errorf("dockerfile template for %s has no content", lang)
		}

		lines := make([]string, len(spec.TemplateLines))
		copy(lines, spec.TemplateLines)
		templates[lang] = lines
	}

	return templates, nil
}

func parseLanguage(value string) (Language, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(LanguageGo):
		return LanguageGo, true
	case string(LanguageNode):
		return LanguageNode, true
	case string(LanguagePython):
		return LanguagePython, true
	case string(LanguageRuby):
		return LanguageRuby, true
	case string(LanguagePHP):
		return LanguagePHP, true
	case string(LanguageJava):
		return LanguageJava, true
	case string(LanguageDotNet):
		return LanguageDotNet, true
	case string(LanguageUnknown):
		return LanguageUnknown, true
	default:
		return "", false
	}
}
