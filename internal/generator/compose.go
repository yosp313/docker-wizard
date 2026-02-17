package generator

import (
	"fmt"
	"sort"
	"strings"
)

type ComposeSelection struct {
	Services []string
}

func Compose(root string, selection ComposeSelection) (string, error) {
	if selection.Services == nil {
		selection.Services = []string{}
	}

	serviceMap, ordered, err := CatalogMap(root)
	if err != nil {
		return "", err
	}

	selected := make(map[string]bool, len(selection.Services))
	for _, id := range selection.Services {
		if _, ok := serviceMap[id]; !ok {
			return "", fmt.Errorf("unknown service: %s", id)
		}
		selected[id] = true
	}

	if err := expandRequiredServices(selected, serviceMap); err != nil {
		return "", err
	}

	services := []ServiceSpec{appServiceSpec()}
	var volumes []string

	for _, spec := range ordered {
		if !selected[spec.ID] {
			continue
		}
		services = append(services, spec)
		volumes = append(volumes, spec.NamedVolumes...)
	}

	volumes = uniqueStrings(volumes)

	builder := &strings.Builder{}
	builder.WriteString("version: \"3.9\"\n")
	builder.WriteString("services:\n")

	for _, svc := range services {
		writeService(builder, svc)
	}

	if len(volumes) > 0 {
		sort.Strings(volumes)
		builder.WriteString("volumes:\n")
		for _, name := range volumes {
			builder.WriteString("  " + name + ":\n")
		}
	}

	return builder.String(), nil
}

func expandRequiredServices(selected map[string]bool, services map[string]ServiceSpec) error {
	changed := true
	for changed {
		changed = false
		for id := range selected {
			svc, ok := services[id]
			if !ok {
				return fmt.Errorf("unknown service: %s", id)
			}
			for _, req := range svc.Requires {
				if !selected[req] {
					selected[req] = true
					changed = true
				}
			}
		}
	}

	return nil
}

func appServiceSpec() ServiceSpec {
	return ServiceSpec{
		ID:    "app",
		Name:  "app",
		Ports: []string{"8080:8080"},
	}
}

func writeService(builder *strings.Builder, svc ServiceSpec) {
	builder.WriteString("  " + svc.Name + ":\n")
	if svc.ID == "app" {
		builder.WriteString("    build:\n")
		builder.WriteString("      context: .\n")
		builder.WriteString("      dockerfile: Dockerfile\n")
	}
	if svc.Image != "" {
		builder.WriteString("    image: " + svc.Image + "\n")
	}
	if len(svc.Ports) > 0 {
		builder.WriteString("    ports:\n")
		for _, port := range svc.Ports {
			builder.WriteString("      - \"" + port + "\"\n")
		}
	}
	if len(svc.Env) > 0 {
		builder.WriteString("    environment:\n")
		for _, env := range svc.Env {
			builder.WriteString("      - " + env + "\n")
		}
	}
	if len(svc.Command) > 0 {
		builder.WriteString("    command:\n")
		for _, arg := range svc.Command {
			builder.WriteString("      - " + arg + "\n")
		}
	}
	if len(svc.VolumeMounts) > 0 {
		builder.WriteString("    volumes:\n")
		for _, mount := range svc.VolumeMounts {
			builder.WriteString("      - " + mount + "\n")
		}
	}
	if len(svc.DependsOn) > 0 {
		builder.WriteString("    depends_on:\n")
		for _, dep := range svc.DependsOn {
			builder.WriteString("      - " + dep + "\n")
		}
	}
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	unique := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		unique[value] = true
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	return result
}
