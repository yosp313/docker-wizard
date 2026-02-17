package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
		return nil, fmt.Errorf("read service catalog: %w", err)
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
