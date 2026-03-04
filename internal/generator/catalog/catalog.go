package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"docker-wizard/internal/utils"
)

type ServiceCatalog struct {
	Services []ServiceSpec `json:"services"`
}

func LoadCatalog(root string) (ServiceCatalog, error) {
	if root == "" {
		return ServiceCatalog{}, fmt.Errorf("root directory is required")
	}

	data, err := readCatalogData(root)
	if err != nil {
		return ServiceCatalog{}, err
	}

	var catalog ServiceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return ServiceCatalog{}, fmt.Errorf("parse service catalog: %w", err)
	}

	if err := normalizeCatalog(&catalog); err != nil {
		return ServiceCatalog{}, err
	}

	return catalog, nil
}

func readCatalogData(root string) ([]byte, error) {
	primary := filepath.Join(root, "config", "services.json")
	data, err := os.ReadFile(primary)
	if err == nil {
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read service catalog: %w", err)
	}

	exe, exeErr := os.Executable()
	if exeErr != nil {
		return nil, fmt.Errorf("read service catalog: %w", exeErr)
	}
	secondary := filepath.Join(filepath.Dir(exe), "config", "services.json")
	data, err = os.ReadFile(secondary)
	if err != nil {
		return nil, fmt.Errorf("read service catalog: %w", err)
	}
	return data, nil
}

func normalizeCatalog(catalog *ServiceCatalog) error {
	ids := make(map[string]bool, len(catalog.Services))
	for i := range catalog.Services {
		svc := &catalog.Services[i]
		if svc.ID == "" {
			return fmt.Errorf("service id is required")
		}
		if ids[svc.ID] {
			return fmt.Errorf("duplicate service id: %s", svc.ID)
		}
		ids[svc.ID] = true
		if svc.Name == "" {
			svc.Name = svc.ID
		}
		if svc.Label == "" {
			svc.Label = svc.ID
		}
		if svc.Selectable && svc.Category == "" {
			return fmt.Errorf("service %s missing category", svc.ID)
		}
		if svc.Category != "" && !validCategory(svc.Category) {
			return fmt.Errorf("service %s has invalid category: %s", svc.ID, svc.Category)
		}
	}

	for _, svc := range catalog.Services {
		for _, dep := range svc.Requires {
			if !ids[dep] {
				return fmt.Errorf("service %s requires missing %s", svc.ID, dep)
			}
		}
		for _, dep := range svc.DependsOn {
			if !ids[dep] {
				return fmt.Errorf("service %s depends on missing %s", svc.ID, dep)
			}
		}
	}

	return nil
}

func validCategory(category string) bool {
	switch category {
	case utils.CategoryDatabase, utils.CategoryMessageQueue, utils.CategoryCache, utils.CategoryAnalytics, utils.CategoryProxy:
		return true
	default:
		return false
	}
}

func SelectableServices(root string) ([]ServiceSpec, error) {
	catalog, err := LoadCatalog(root)
	if err != nil {
		return nil, err
	}

	services := make([]ServiceSpec, 0, len(catalog.Services))
	for _, svc := range catalog.Services {
		if svc.Selectable {
			services = append(services, svc)
		}
	}

	sortServices(services)
	return services, nil
}

func CatalogMap(root string) (map[string]ServiceSpec, []ServiceSpec, error) {
	catalog, err := LoadCatalog(root)
	if err != nil {
		return nil, nil, err
	}

	services := make([]ServiceSpec, len(catalog.Services))
	copy(services, catalog.Services)
	sortServices(services)

	serviceMap := make(map[string]ServiceSpec, len(services))
	for _, svc := range services {
		serviceMap[svc.ID] = svc
	}

	return serviceMap, services, nil
}

func sortServices(services []ServiceSpec) {
	sort.SliceStable(services, func(i, j int) bool {
		if services[i].Order == services[j].Order {
			return services[i].Label < services[j].Label
		}
		return services[i].Order < services[j].Order
	})
}

// AppendService loads the catalog at <root>/config/services.json (creating it if
// absent), appends svc with an auto-generated unique ID, and writes the file back.
func AppendService(root string, svc ServiceSpec) error {
	if root == "" {
		return fmt.Errorf("root directory is required")
	}
	if !validCategory(svc.Category) {
		return fmt.Errorf("invalid category: %s", svc.Category)
	}

	path := filepath.Join(root, "config", "services.json")

	var existing ServiceCatalog
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse service catalog: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read service catalog: %w", err)
	}

	id, err := generateUniqueID(svc.Name, existing.Services)
	if err != nil {
		return err
	}
	svc.ID = id
	svc.Selectable = true
	svc.Public = false
	if svc.Order == 0 {
		svc.Order = 100
	}

	existing.Services = append(existing.Services, svc)

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal service catalog: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	return os.WriteFile(path, out, 0o644)
}

func slugify(name string) string {
	s := strings.ToLower(name)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ', r == '_':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func generateUniqueID(name string, services []ServiceSpec) (string, error) {
	existing := make(map[string]bool, len(services))
	for _, s := range services {
		existing[s.ID] = true
	}

	base := slugify(name)
	if base == "" {
		base = "custom-service"
	}

	if !existing[base] {
		return base, nil
	}

	for i := 2; i <= 99; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !existing[candidate] {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not generate unique ID for service %q", name)
}
