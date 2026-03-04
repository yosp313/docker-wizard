package wizard

import (
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

	categories := utils.CategoryOrder()
	category := categories[m.addServiceCategoryIdx]

	svc := catalog.ServiceSpec{
		Name:         name,
		Label:        name,
		Image:        image,
		Category:     category,
		Ports:        splitCommaValues(m.addServiceInputs[2].Value()),
		Env:          splitCommaValues(m.addServiceInputs[3].Value()),
		VolumeMounts: splitCommaValues(m.addServiceInputs[4].Value()),
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
