package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

var nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]`)

func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumHyphen.ReplaceAllString(s, "")
	return s
}

// AppendService appends a new ServiceSpec to <root>/config/services.json.
// It auto-generates a unique ID from svc.Name (or svc.Label as fallback),
// applies sensible defaults, validates the category, and writes the file.
func AppendService(root string, svc ServiceSpec) error {
	if svc.Category == "" || !validCategory(svc.Category) {
		return fmt.Errorf("invalid category: %q", svc.Category)
	}

	// Load existing catalog (treat missing file as empty).
	primaryPath := filepath.Join(root, "config", "services.json")
	var existing ServiceCatalog
	data, err := os.ReadFile(primaryPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing catalog: %w", err)
	}
	if err == nil {
		if jsonErr := json.Unmarshal(data, &existing); jsonErr != nil {
			return fmt.Errorf("parse existing catalog: %w", jsonErr)
		}
	}

	// Build set of existing IDs for uniqueness check.
	existingIDs := make(map[string]bool, len(existing.Services))
	for _, s := range existing.Services {
		existingIDs[s.ID] = true
	}

	// Determine the base name for slug generation.
	baseName := svc.Name
	if baseName == "" {
		baseName = svc.Label
	}

	// Generate unique ID.
	baseSlug := slugify(baseName)
	if baseSlug == "" {
		baseSlug = "service"
	}

	candidate := baseSlug
	if existingIDs[candidate] {
		found := false
		for i := 2; i <= 99; i++ {
			candidate = fmt.Sprintf("%s-%d", baseSlug, i)
			if !existingIDs[candidate] {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("cannot generate unique id for %q: all candidates up to -99 are taken", baseName)
		}
	}
	svc.ID = candidate

	// Apply defaults.
	if !svc.Selectable {
		svc.Selectable = true
	}
	// Public defaults to false — zero value is already false, nothing to do.
	if svc.Order == 0 {
		svc.Order = 100
	}
	if svc.Name == "" {
		if svc.Label != "" {
			svc.Name = svc.Label
		} else {
			svc.Name = svc.ID
		}
	}
	if svc.Label == "" {
		svc.Label = svc.Name
	}

	// Append and marshal.
	existing.Services = append(existing.Services, svc)
	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	// Ensure directory exists.
	if mkErr := os.MkdirAll(filepath.Join(root, "config"), 0o755); mkErr != nil {
		return fmt.Errorf("create config directory: %w", mkErr)
	}

	if writeErr := os.WriteFile(primaryPath, out, 0o644); writeErr != nil {
		return fmt.Errorf("write catalog: %w", writeErr)
	}

	return nil
}
