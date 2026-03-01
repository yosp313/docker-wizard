package cli

import (
	"fmt"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/utils"
)

func RunList(root string) error {
	if root == "" {
		return fmt.Errorf("root directory is required")
	}

	services, err := generator.SelectableServices(root)
	if err != nil {
		return err
	}

	// group services by category
	grouped := make(map[string][]generator.ServiceSpec)
	for _, svc := range services {
		grouped[svc.Category] = append(grouped[svc.Category], svc)
	}

	fmt.Println("Available services:")
	for _, cat := range utils.CategoryOrder() {
		svcs, ok := grouped[cat]
		if !ok {
			continue
		}
		fmt.Printf("\n  %s:\n", utils.CategoryLabel(cat))
		for _, svc := range svcs {
			fmt.Printf("    %-20s %s\n", svc.ID, svc.Label)
		}
	}

	return nil
}
