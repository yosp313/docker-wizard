package compose

import (
	"fmt"
	"sort"
	"strings"

	"docker-wizard/internal/generator/catalog"
)

type ComposeSelection struct {
	Services []string
}

func Compose(root string, selection ComposeSelection) (string, error) {
	if selection.Services == nil {
		selection.Services = []string{}
	}

	serviceMap, ordered, err := catalog.CatalogMap(root)
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

	if err := ExpandRequiredServices(selected, serviceMap); err != nil {
		return "", err
	}

	services := []catalog.ServiceSpec{AppServiceSpec()}
	var volumes []string

	for _, spec := range ordered {
		if !selected[spec.ID] {
			continue
		}
		services = append(services, filterDepends(spec, selected))
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

	builder.WriteString("networks:\n")
	builder.WriteString("  app-net:\n")

	return builder.String(), nil
}

func ExpandRequiredServices(selected map[string]bool, services map[string]catalog.ServiceSpec) error {
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

func AppServiceSpec() catalog.ServiceSpec {
	return catalog.ServiceSpec{
		ID:     "app",
		Name:   "app",
		Ports:  []string{"8080:8080"},
		Public: true,
	}
}

func writeService(builder *strings.Builder, svc catalog.ServiceSpec) {
	builder.WriteString("  " + svc.Name + ":\n")
	if svc.ID == "app" {
		builder.WriteString("    build:\n")
		builder.WriteString("      context: .\n")
		builder.WriteString("      dockerfile: Dockerfile\n")
	}
	if svc.Image != "" {
		builder.WriteString("    image: " + svc.Image + "\n")
	}
	ports := append([]string(nil), svc.Ports...)
	if len(ports) > 0 {
		sort.Strings(ports)
		if svc.Public {
			builder.WriteString("    ports:\n")
			for _, port := range ports {
				builder.WriteString("      - \"" + port + "\"\n")
			}
		} else {
			expose := append([]string(nil), svc.Expose...)
			if len(expose) == 0 {
				expose = portsToExpose(ports)
			}
			if len(expose) > 0 {
				sort.Strings(expose)
				builder.WriteString("    expose:\n")
				for _, port := range expose {
					builder.WriteString("      - \"" + port + "\"\n")
				}
			}
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
		depends := append([]string(nil), svc.DependsOn...)
		sort.Strings(depends)
		builder.WriteString("    depends_on:\n")
		for _, dep := range depends {
			builder.WriteString("      - " + dep + "\n")
		}
	}
	builder.WriteString("    networks:\n")
	builder.WriteString("      - app-net\n")
}

func filterDepends(spec catalog.ServiceSpec, selected map[string]bool) catalog.ServiceSpec {
	if len(spec.DependsOn) == 0 {
		return spec
	}
	filtered := make([]string, 0, len(spec.DependsOn))
	for _, dep := range spec.DependsOn {
		if dep == "app" || selected[dep] {
			filtered = append(filtered, dep)
		}
	}
	updated := spec
	updated.DependsOn = filtered
	return updated
}

func portsToExpose(ports []string) []string {
	expose := make([]string, 0, len(ports))
	for _, port := range ports {
		container := containerPort(port)
		if container != "" {
			expose = append(expose, container)
		}
	}
	return expose
}

func containerPort(port string) string {
	parts := strings.Split(port, ":")
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[len(parts)-1]
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
