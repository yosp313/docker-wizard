package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/utils"
)

type languageOption struct {
	Label    string
	Language generator.Language
	Auto     bool
}

func RunInteractive(root string) error {
	if root == "" {
		return fmt.Errorf("root directory is required")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Docker Wizard (CLI interactive mode)")
	fmt.Println()

	details, err := generator.DetectLanguage(root)
	if err != nil {
		return err
	}

	overrideType, overrideLang, err := promptLanguage(reader, details)
	if err != nil {
		return err
	}
	if overrideLang {
		details.Type = overrideType
	}

	services, err := generator.SelectableServices(root)
	if err != nil {
		return err
	}

	selected, err := promptServicesByCategory(reader, services)
	if err != nil {
		return err
	}
	selectedIDs := orderedSelectedIDs(services, selected)

	selection := generator.ComposeSelection{Services: selectedIDs}
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

	fmt.Println()
	fmt.Println("Review")
	fmt.Printf("- language: %s\n", languageLabelWithVersion(details))
	printSelectedSummary(services, selected)

	fmt.Println("- managed files:")
	fmt.Printf("  - docker-compose.yml (%s)\n", previewStatusLabel(preview.Compose.Status))
	fmt.Printf("  - Dockerfile (%s)\n", previewStatusLabel(preview.Dockerfile.Status))
	if preview.Dockerignore.Status == generator.FileStatusNew {
		fmt.Printf("  - .dockerignore (%s)\n", previewStatusLabel(preview.Dockerignore.Status))
	}

	if len(warnings) > 0 {
		sort.Strings(warnings)
		fmt.Println("- warnings:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	confirm, err := promptYesNo(reader, "\nGenerate files now? [y/N]: ")
	if err != nil {
		return err
	}
	if !confirm {
		fmt.Println("Cancelled. No files were written.")
		return nil
	}

	output, err := generator.WriteFiles(root, composeContent, dockerfileContent)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Done")
	fmt.Printf("- docker-compose.yml: %s\n", output.ComposeStatus)
	fmt.Printf("- Dockerfile: %s\n", output.DockerfileStatus)
	fmt.Printf("- .dockerignore: %s\n", output.DockerignoreStatus)
	if output.ComposeBackupPath != "" || output.DockerfileBackupPath != "" {
		fmt.Println("- backups:")
		if output.ComposeBackupPath != "" {
			fmt.Printf("  - %s\n", filepath.Base(output.ComposeBackupPath))
		}
		if output.DockerfileBackupPath != "" {
			fmt.Printf("  - %s\n", filepath.Base(output.DockerfileBackupPath))
		}
	}

	fmt.Println("- next: docker compose up")

	return nil
}

func promptLanguage(reader *bufio.Reader, details generator.LanguageDetails) (generator.Language, bool, error) {
	options := []languageOption{
		{Label: "Auto-detect", Auto: true},
		{Label: "Go", Language: generator.LanguageGo},
		{Label: "Node", Language: generator.LanguageNode},
		{Label: "Python", Language: generator.LanguagePython},
		{Label: "Ruby", Language: generator.LanguageRuby},
		{Label: "PHP", Language: generator.LanguagePHP},
		{Label: "Java", Language: generator.LanguageJava},
		{Label: ".NET", Language: generator.LanguageDotNet},
	}

	autoLabel := options[0].Label
	detected := languageLabelWithVersion(details)
	if detected != "" && details.Type != generator.LanguageUnknown {
		autoLabel = autoLabel + " (" + detected + ")"
	}
	options[0].Label = autoLabel

	fmt.Printf("Detected language: %s\n", languageLabelWithVersion(details))
	fmt.Println("Choose language:")
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i, option.Label)
	}

	for {
		input, err := promptLine(reader, "Language [0]: ")
		if err != nil {
			return "", false, err
		}
		if input == "" {
			return "", false, nil
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 0 || idx >= len(options) {
			fmt.Println("Enter a valid option number.")
			continue
		}
		chosen := options[idx]
		if chosen.Auto {
			return "", false, nil
		}
		return chosen.Language, true, nil
	}
}

func promptServicesByCategory(reader *bufio.Reader, services []generator.ServiceSpec) (map[string]bool, error) {
	selected := map[string]bool{}
	categories := utils.CategoryOrder()

	for _, category := range categories {
		group := filterServicesByCategory(services, category)
		if len(group) == 0 {
			continue
		}

		fmt.Println()
		fmt.Printf("%s\n", utils.CategoryLabel(category))
		for i, svc := range group {
			line := svc.Label
			if svc.Description != "" {
				line += " - " + svc.Description
			}
			fmt.Printf("  %d) %s\n", i+1, line)
		}

		for {
			input, err := promptLine(reader, "Select numbers (comma), 'all', or Enter for none: ")
			if err != nil {
				return nil, err
			}
			if input == "" {
				break
			}
			if strings.EqualFold(input, "all") {
				for _, svc := range group {
					selected[svc.ID] = true
				}
				break
			}

			indexes, parseErr := parseIndexSelection(input, len(group))
			if parseErr != nil {
				fmt.Printf("%v\n", parseErr)
				continue
			}
			for _, idx := range indexes {
				selected[group[idx-1].ID] = true
			}
			break
		}
	}

	return selected, nil
}

func promptLine(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptYesNo(reader *bufio.Reader, prompt string) (bool, error) {
	for {
		input, err := promptLine(reader, prompt)
		if err != nil {
			return false, err
		}
		if input == "" {
			return false, nil
		}
		normalized := strings.ToLower(strings.TrimSpace(input))
		if normalized == "y" || normalized == "yes" {
			return true, nil
		}
		if normalized == "n" || normalized == "no" {
			return false, nil
		}
		fmt.Println("Enter y or n.")
	}
}

func parseIndexSelection(input string, max int) ([]int, error) {
	tokens := strings.Split(input, ",")
	values := map[int]bool{}
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		index, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("%q is not a number", token)
		}
		if index < 1 || index > max {
			return nil, fmt.Errorf("%d is out of range 1-%d", index, max)
		}
		values[index] = true
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("enter at least one number")
	}

	indexes := make([]int, 0, len(values))
	for index := range values {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	return indexes, nil
}

func filterServicesByCategory(services []generator.ServiceSpec, category string) []generator.ServiceSpec {
	filtered := make([]generator.ServiceSpec, 0, len(services))
	for _, svc := range services {
		if svc.Category == category {
			filtered = append(filtered, svc)
		}
	}
	return filtered
}

func orderedSelectedIDs(services []generator.ServiceSpec, selected map[string]bool) []string {
	return utils.OrderedSelectedIDs(services, func(svc generator.ServiceSpec) string {
		return svc.ID
	}, selected)
}

func printSelectedSummary(services []generator.ServiceSpec, selected map[string]bool) {
	categoryOrder := utils.CategoryOrder()
	grouped := map[string][]string{}
	for _, svc := range services {
		if selected[svc.ID] {
			grouped[svc.Category] = append(grouped[svc.Category], svc.Label)
		}
	}

	fmt.Println("- selected services:")
	for _, category := range categoryOrder {
		labels := grouped[category]
		if len(labels) == 0 {
			fmt.Printf("  - %s: none\n", utils.CategoryLabel(category))
			continue
		}
		fmt.Printf("  - %s: %s\n", utils.CategoryLabel(category), strings.Join(labels, ", "))
	}
}

func previewStatusLabel(status generator.FileStatus) string {
	switch status {
	case generator.FileStatusNew:
		return "new"
	case generator.FileStatusSame:
		return "matches existing"
	case generator.FileStatusDifferent:
		return "differs from existing (will merge and create .bak)"
	case generator.FileStatusExists:
		return "exists"
	default:
		return string(status)
	}
}

func languageLabelWithVersion(details generator.LanguageDetails) string {
	return utils.LanguageLabelWithVersion(string(details.Type), utils.LanguageVersions{
		Go:     details.GoVersion,
		Node:   details.NodeVersion,
		Python: details.PythonVersion,
		Ruby:   details.RubyVersion,
		PHP:    details.PHPVersion,
		Java:   details.JavaVersion,
		DotNet: details.DotNetVersion,
	})
}
