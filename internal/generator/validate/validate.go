package validate

import (
	"fmt"
	"sort"
	"strings"

	"docker-wizard/internal/generator/catalog"
	"docker-wizard/internal/generator/compose"
)

func SelectionWarnings(root string, selection compose.ComposeSelection) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("root directory is required")
	}

	serviceMap, _, err := catalog.CatalogMap(root)
	if err != nil {
		return nil, err
	}

	selected := make(map[string]bool, len(selection.Services))
	for _, id := range selection.Services {
		if id == "" {
			continue
		}
		if _, ok := serviceMap[id]; !ok {
			return nil, fmt.Errorf("unknown service: %s", id)
		}
		selected[id] = true
	}

	if err := compose.ExpandRequiredServices(selected, serviceMap); err != nil {
		return nil, err
	}

	warnings := []string{}
	warnings = append(warnings, dependencyWarnings(selected, serviceMap)...)
	warnings = append(warnings, portCollisionWarnings(selected, serviceMap)...)
	warnings = append(warnings, insecureDefaultWarnings(selected, serviceMap)...)
	sort.Strings(warnings)
	return warnings, nil
}

func dependencyWarnings(selected map[string]bool, services map[string]catalog.ServiceSpec) []string {
	warnings := []string{}
	for id := range selected {
		svc, ok := services[id]
		if !ok {
			continue
		}
		for _, dep := range svc.DependsOn {
			if dep == "app" {
				continue
			}
			if !selected[dep] {
				label := serviceDisplayName(svc)
				depLabel := dep
				if depSvc, ok := services[dep]; ok {
					depLabel = serviceDisplayName(depSvc)
				}
				warnings = append(warnings, fmt.Sprintf("%s depends on %s but it is not selected", label, depLabel))
			}
		}
	}
	return warnings
}

func portCollisionWarnings(selected map[string]bool, services map[string]catalog.ServiceSpec) []string {
	portOwners := map[string]map[string]bool{}
	selectedServices := make([]catalog.ServiceSpec, 0, len(selected)+1)
	selectedServices = append(selectedServices, compose.AppServiceSpec())
	for id := range selected {
		if svc, ok := services[id]; ok {
			selectedServices = append(selectedServices, svc)
		}
	}

	for _, svc := range selectedServices {
		if !svc.Public {
			continue
		}
		for _, port := range svc.Ports {
			host := hostPort(port)
			if host == "" {
				continue
			}
			owners, ok := portOwners[host]
			if !ok {
				owners = map[string]bool{}
				portOwners[host] = owners
			}
			owners[serviceDisplayName(svc)] = true
		}
	}

	ports := make([]string, 0, len(portOwners))
	for port := range portOwners {
		ports = append(ports, port)
	}
	sort.Strings(ports)

	warnings := []string{}
	for _, port := range ports {
		owners := portOwners[port]
		if len(owners) < 2 {
			continue
		}
		labels := make([]string, 0, len(owners))
		for owner := range owners {
			labels = append(labels, owner)
		}
		sort.Strings(labels)
		warnings = append(warnings, fmt.Sprintf("host port %s is published by %s", port, strings.Join(labels, ", ")))
	}

	return warnings
}

func serviceDisplayName(svc catalog.ServiceSpec) string {
	if svc.Label != "" {
		return svc.Label
	}
	if svc.Name != "" {
		return svc.Name
	}
	if svc.ID != "" {
		return svc.ID
	}
	return "service"
}

func hostPort(port string) string {
	parts := strings.Split(port, ":")
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return ""
	case 2:
		return parts[0]
	default:
		return parts[len(parts)-2]
	}
}

func insecureDefaultWarnings(selected map[string]bool, services map[string]catalog.ServiceSpec) []string {
	warnings := []string{}
	for id := range selected {
		svc, ok := services[id]
		if !ok {
			continue
		}

		label := serviceDisplayName(svc)
		for _, env := range svc.Env {
			if insecureEnvDefault(env) {
				warnings = append(warnings, fmt.Sprintf("%s includes placeholder environment defaults; update them before sharing or exposing this stack", label))
				break
			}
		}

		for _, arg := range svc.Command {
			trimmed := strings.TrimSpace(strings.ToLower(arg))
			if trimmed == "--api.insecure=true" {
				warnings = append(warnings, fmt.Sprintf("%s enables an insecure admin API flag (--api.insecure=true)", label))
				break
			}
		}
	}

	return dedupeStrings(warnings)
}

func insecureEnvDefault(entry string) bool {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return false
	}
	idx := strings.Index(entry, "=")
	if idx < 0 || idx == len(entry)-1 {
		return false
	}
	value := strings.ToLower(strings.TrimSpace(entry[idx+1:]))
	if value == "" {
		return false
	}

	return strings.Contains(value, "change-me") || strings.Contains(value, "example")
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
