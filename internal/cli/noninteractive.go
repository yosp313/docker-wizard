package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"docker-wizard/internal/generator"
)

type NonInteractiveOptions struct {
	Services []string
	Language string
	Write    bool
	DryRun   bool
}

func RunNonInteractive(root string, options NonInteractiveOptions) error {
	if root == "" {
		return fmt.Errorf("root directory is required")
	}
	if options.Write && options.DryRun {
		return fmt.Errorf("--write and --dry-run cannot be used together")
	}

	dryRun := options.DryRun || !options.Write

	overrideType, overrideLang, err := parseLanguageOption(options.Language)
	if err != nil {
		return err
	}

	selectedServices, err := resolveServices(root, options.Services)
	if err != nil {
		return err
	}

	details, err := generator.DetectLanguage(root)
	if err != nil {
		return err
	}
	if overrideLang {
		details.Type = overrideType
	}

	selection := generator.ComposeSelection{Services: selectedServices}
	warnings, err := generator.SelectionWarnings(root, selection)
	if err != nil {
		return err
	}

	dockerfileContent, err := generator.Dockerfile(root, details)
	if err != nil {
		return err
	}
	composeContent, err := generator.Compose(root, selection)
	if err != nil {
		return err
	}

	preview, err := generator.PreviewFiles(root, composeContent, dockerfileContent)
	if err != nil {
		return err
	}

	fmt.Println("Docker Wizard (batch mode)")
	fmt.Printf("- language: %s\n", languageLabelWithVersion(details))
	fmt.Printf("- selected services: %s\n", serviceSelectionLabel(selectedServices))

	if len(warnings) > 0 {
		sort.Strings(warnings)
		fmt.Println("- warnings:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if dryRun {
		fmt.Println("- managed files:")
		fmt.Printf("  - docker-compose.yml (%s)\n", previewStatusLabel(preview.Compose.Status))
		fmt.Printf("  - Dockerfile (%s)\n", previewStatusLabel(preview.Dockerfile.Status))
		fmt.Printf("  - .dockerignore (%s)\n", previewStatusLabel(preview.Dockerignore.Status))
		fmt.Println("- result: dry-run (no files were written)")
		return nil
	}

	output, err := generator.WriteFiles(root, composeContent, dockerfileContent)
	if err != nil {
		return err
	}

	fmt.Println("- result:")
	fmt.Printf("  - docker-compose.yml: %s\n", output.ComposeStatus)
	fmt.Printf("  - Dockerfile: %s\n", output.DockerfileStatus)
	fmt.Printf("  - .dockerignore: %s\n", output.DockerignoreStatus)
	if output.ComposeBackupPath != "" || output.DockerfileBackupPath != "" {
		fmt.Println("  - backups:")
		if output.ComposeBackupPath != "" {
			fmt.Printf("    - %s\n", filepath.Base(output.ComposeBackupPath))
		}
		if output.DockerfileBackupPath != "" {
			fmt.Printf("    - %s\n", filepath.Base(output.DockerfileBackupPath))
		}
	}
	fmt.Println("- next: docker compose up")

	return nil
}

func parseLanguageOption(value string) (generator.Language, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return "", false, nil
	case "go":
		return generator.LanguageGo, true, nil
	case "node":
		return generator.LanguageNode, true, nil
	case "python":
		return generator.LanguagePython, true, nil
	case "ruby":
		return generator.LanguageRuby, true, nil
	case "php":
		return generator.LanguagePHP, true, nil
	case "java":
		return generator.LanguageJava, true, nil
	case "dotnet", ".net":
		return generator.LanguageDotNet, true, nil
	case "unknown":
		return generator.LanguageUnknown, true, nil
	default:
		return "", false, fmt.Errorf("invalid language %q (expected auto, go, node, python, ruby, php, java, dotnet)", value)
	}
}

func resolveServices(root string, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return []string{}, nil
	}
	if len(requested) == 1 && strings.EqualFold(strings.TrimSpace(requested[0]), "all") {
		selectable, err := generator.SelectableServices(root)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(selectable))
		for _, svc := range selectable {
			ids = append(ids, svc.ID)
		}
		return ids, nil
	}

	serviceMap, ordered, err := generator.CatalogMap(root)
	if err != nil {
		return nil, err
	}

	selected := map[string]bool{}
	unknown := []string{}
	for _, id := range requested {
		normalized := strings.ToLower(strings.TrimSpace(id))
		if normalized == "" {
			continue
		}
		if _, ok := serviceMap[normalized]; !ok {
			unknown = append(unknown, normalized)
			continue
		}
		selected[normalized] = true
	}

	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("unknown services: %s", strings.Join(unknown, ", "))
	}

	orderedIDs := make([]string, 0, len(selected))
	for _, svc := range ordered {
		if selected[svc.ID] {
			orderedIDs = append(orderedIDs, svc.ID)
		}
	}

	return orderedIDs, nil
}

func serviceSelectionLabel(services []string) string {
	if len(services) == 0 {
		return "none"
	}
	return strings.Join(services, ", ")
}
