package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"docker-wizard/internal/generator"
)

type AddOptions struct {
	Services []string
	Write    bool
}

func RunAdd(root string, options AddOptions) error {
	if root == "" {
		return fmt.Errorf("root directory is required")
	}
	if len(options.Services) == 0 {
		return fmt.Errorf("at least one service ID is required")
	}

	// validate all service IDs against catalog
	serviceMap, _, err := generator.CatalogMap(root)
	if err != nil {
		return err
	}

	var unknown []string
	for _, id := range options.Services {
		if _, ok := serviceMap[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return fmt.Errorf("unknown services: %s", strings.Join(unknown, ", "))
	}

	// parse existing compose file to detect already-present services
	composePath := filepath.Join(root, generator.ComposeFileName)
	var existing map[string]bool
	if data, err := os.ReadFile(composePath); err == nil {
		existing, err = generator.ExistingComposeServices(string(data))
		if err != nil {
			return fmt.Errorf("parse existing compose: %w", err)
		}
	}

	// filter out already-present services
	var toAdd []string
	var skipped []string
	for _, id := range options.Services {
		name := serviceMap[id].Name
		if existing[name] {
			skipped = append(skipped, id)
		} else {
			toAdd = append(toAdd, id)
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("skipping (already present): %s\n", strings.Join(skipped, ", "))
	}
	if len(toAdd) == 0 {
		fmt.Println("nothing to add — all requested services already exist")
		return nil
	}

	// generate compose fragment (handles dependency expansion internally)
	composeContent, expanded, err := generator.ComposeFragment(root, toAdd)
	if err != nil {
		return err
	}

	if len(expanded) > 0 {
		fmt.Printf("auto-adding dependencies: %s\n", strings.Join(expanded, ", "))
	}

	fmt.Printf("adding services: %s\n", strings.Join(toAdd, ", "))

	if !options.Write {
		// dry-run: preview only
		preview, err := generator.PreviewComposeFile(root, composeContent)
		if err != nil {
			return err
		}
		fmt.Printf("docker-compose.yml: %s (dry-run)\n", previewStatusLabel(preview.Status))
		if preview.Content != "" {
			fmt.Println("---")
			fmt.Print(preview.Content)
			fmt.Println("---")
		}
		fmt.Println("pass --write to apply changes")
		return nil
	}

	// write mode
	status, backup, err := generator.WriteComposeFile(root, composeContent)
	if err != nil {
		return err
	}

	fmt.Printf("docker-compose.yml: %s\n", status)
	if backup != "" {
		fmt.Printf("backup: %s\n", filepath.Base(backup))
	}

	return nil
}
