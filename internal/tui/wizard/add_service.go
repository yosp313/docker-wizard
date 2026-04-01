package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"docker-wizard/internal/generator"
	"docker-wizard/internal/generator/catalog"
	"docker-wizard/internal/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// addServiceFieldCount is the total number of fields in the form (including category).
const addServiceFieldCount = 6

// addServiceTextInputIndex maps the focused field index (0-5) to the textinput
// slice index (0-4). Returns -1 for the category field (index 2).
func addServiceTextInputIndex(field int) int {
	if field < 2 {
		return field
	}
	if field > 2 {
		return field - 1
	}
	return -1
}

func initAddServiceInputs() [5]textinput.Model {
	placeholders := [5]string{
		"e.g. My Redis",
		"e.g. redis:7",
		"e.g. 6379:6379",
		"e.g. FOO=bar,BAR=baz",
		"e.g. /data:/data",
	}
	var inputs [5]textinput.Model
	for i := 0; i < 5; i++ {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		inputs[i] = ti
	}
	inputs[0].Focus()
	return inputs
}

func (m *model) syncAddServiceFocus() {
	for i := range m.addServiceInputs {
		m.addServiceInputs[i].Blur()
	}
	idx := addServiceTextInputIndex(m.addServiceFocusedField)
	if idx >= 0 {
		m.addServiceInputs[idx].Focus()
	}
}

func (m *model) resetAddServiceForm() {
	m.addServiceFocusedField = 0
	m.addServiceCategoryIdx = 0
	m.addServiceFormError = ""
	for i := range m.addServiceInputs {
		m.addServiceInputs[i].Reset()
		m.addServiceInputs[i].Blur()
	}
	m.addServiceInputs[0].Focus()
}

func (m *model) validateAddServiceForm() bool {
	m.addServiceFormError = ""

	ports := splitCommaValues(m.addServiceInputs[2].Value())
	if err := validatePorts(ports); err != nil {
		m.addServiceFormError = err.Error()
		return false
	}

	env := splitCommaValues(m.addServiceInputs[3].Value())
	if err := validateEnvVars(env); err != nil {
		m.addServiceFormError = err.Error()
		return false
	}

	volumes := splitCommaValues(m.addServiceInputs[4].Value())
	if err := validateVolumes(volumes); err != nil {
		m.addServiceFormError = err.Error()
		return false
	}

	return true
}

func (m *model) confirmAddService() tea.Cmd {
	name := strings.TrimSpace(m.addServiceInputs[0].Value())
	image := strings.TrimSpace(m.addServiceInputs[1].Value())

	if name == "" {
		m.addServiceFormError = "Name is required"
		return nil
	}
	if image == "" {
		m.addServiceFormError = "Docker Image is required"
		return nil
	}

	if !m.validateAddServiceForm() {
		return nil
	}

	ports := splitCommaValues(m.addServiceInputs[2].Value())
	env := splitCommaValues(m.addServiceInputs[3].Value())
	volumes := splitCommaValues(m.addServiceInputs[4].Value())

	categories := utils.CategoryOrder()
	category := categories[m.addServiceCategoryIdx]

	svc := catalog.ServiceSpec{
		Name:         name,
		Label:        name,
		Image:        image,
		Category:     category,
		Ports:        ports,
		Env:          env,
		VolumeMounts: volumes,
	}

	if err := catalog.AppendService(m.root, svc); err != nil {
		m.addServiceFormError = err.Error()
		return nil
	}

	services, err := generator.SelectableServices(m.root)
	if err != nil {
		m.addServiceFormError = err.Error()
		return nil
	}
	m.services = serviceChoicesFromCatalog(services)
	m.step = m.previousStep
	m.animateHeader()
	return nil
}

func splitCommaValues(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// validation functions for form inputs
func validatePorts(ports []string) error {
	for _, p := range ports {
		if !strings.Contains(p, ":") {
			return fmt.Errorf("invalid port format %q: expected host:container (e.g., 8080:80)", p)
		}
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid port format %q", p)
		}
		for _, part := range parts {
			if _, err := strconv.Atoi(part); err != nil {
				return fmt.Errorf("invalid port number in %q: %v", p, err)
			}
		}
	}
	return nil
}

func validateEnvVars(env []string) error {
	for _, e := range env {
		if !strings.Contains(e, "=") {
			return fmt.Errorf("invalid env var format %q: expected KEY=VALUE", e)
		}
	}
	return nil
}

func validateVolumes(volumes []string) error {
	for _, v := range volumes {
		if !strings.Contains(v, ":") {
			return fmt.Errorf("invalid volume format %q: expected host:container (e.g., ./data:/data)", v)
		}
	}
	return nil
}
