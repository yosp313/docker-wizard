package wizard

import "docker-wizard/internal/generator"

import "docker-wizard/internal/utils"

func (m *model) toggleCurrentSelection() {
	services := m.filteredServices(m.step)
	if len(services) == 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(services) {
		return
	}
	id := services[m.cursor].ID
	if m.selected[id] {
		delete(m.selected, id)
		return
	}
	m.selected[id] = true
}

func (m model) filteredServices(current step) []serviceChoice {
	category := stepCategory(current)
	if category == "" {
		return nil
	}
	filtered := make([]serviceChoice, 0, len(m.services))
	for _, svc := range m.services {
		if svc.Category == category {
			filtered = append(filtered, svc)
		}
	}
	return filtered
}

func (m model) selectedByCategory() map[string][]string {
	grouped := map[string][]string{}
	for _, svc := range m.services {
		if !m.selected[svc.ID] {
			continue
		}
		grouped[svc.Category] = append(grouped[svc.Category], svc.Label)
	}
	return grouped
}

func stepCategory(current step) string {
	switch current {
	case stepDatabase:
		return utils.CategoryDatabase
	case stepMessageQueue:
		return utils.CategoryMessageQueue
	case stepCache:
		return utils.CategoryCache
	case stepAnalytics:
		return utils.CategoryAnalytics
	case stepProxy:
		return utils.CategoryProxy
	default:
		return ""
	}
}

func stepTitle(current step) string {
	switch current {
	case stepDatabase:
		return "Databases"
	case stepMessageQueue:
		return "Message Queues"
	case stepCache:
		return "Caching"
	case stepAnalytics:
		return "Analytics"
	case stepProxy:
		return "Webservers / Proxies"
	case stepAddService:
		return "Add Service"
	default:
		return "Services"
	}
}

func defaultLanguageOptions() []languageChoice {
	return []languageChoice{
		{ID: "auto", Label: "Auto-detect", Description: "use detected language"},
		{ID: "go", Label: "Go", Language: generator.LanguageGo},
		{ID: "node", Label: "Node", Language: generator.LanguageNode},
		{ID: "python", Label: "Python", Language: generator.LanguagePython},
		{ID: "ruby", Label: "Ruby", Language: generator.LanguageRuby},
		{ID: "php", Label: "PHP", Language: generator.LanguagePHP},
		{ID: "java", Label: "Java", Language: generator.LanguageJava},
		{ID: "dotnet", Label: ".NET", Language: generator.LanguageDotNet},
	}
}

func languageOptionsForDetected(details generator.LanguageDetails) []languageChoice {
	options := defaultLanguageOptions()
	if details.Type == generator.LanguageUnknown {
		return options
	}
	label := "Auto-detect"
	versioned := languageLabelWithVersion(details)
	if versioned != "" {
		label = "Auto-detect (" + versioned + ")"
	}
	options[0].Label = label
	return options
}

func clampCursor(cursor int, length int) int {
	if length <= 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func (m *model) nextStep() step {
	switch m.step {
	case stepDatabase:
		return stepMessageQueue
	case stepMessageQueue:
		return stepCache
	case stepCache:
		return stepAnalytics
	case stepAnalytics:
		return stepProxy
	case stepProxy:
		return stepReview
	default:
		return m.step
	}
}

func (m *model) prevStep() step {
	switch m.step {
	case stepDatabase:
		if m.langVisited {
			return stepLanguage
		}
		return stepDetect
	case stepMessageQueue:
		return stepDatabase
	case stepCache:
		return stepMessageQueue
	case stepAnalytics:
		return stepCache
	case stepProxy:
		return stepAnalytics
	default:
		return m.step
	}
}

func (m model) stepIndex() int {
	switch m.step {
	case stepWelcome:
		return 1
	case stepDetect:
		return 2
	case stepLanguage:
		return 3
	case stepDatabase:
		return 4
	case stepMessageQueue:
		return 5
	case stepCache:
		return 6
	case stepAnalytics:
		return 7
	case stepProxy:
		return 8
	case stepReview:
		return 9
	case stepPreview:
		return 10
	case stepGenerate:
		return 11
	case stepResult:
		return 11
	case stepError:
		return 11
	case stepAddService:
		return 0
	default:
		return 1
	}
}

func serviceChoicesFromCatalog(services []generator.ServiceSpec) []serviceChoice {
	choices := make([]serviceChoice, 0, len(services))
	for _, svc := range services {
		choices = append(choices, serviceChoice{
			ID:          svc.ID,
			Label:       svc.Label,
			Description: svc.Description,
			Category:    svc.Category,
		})
	}
	return choices
}

func languageLabelWithVersion(details generator.LanguageDetails) string {
	return utils.LanguageLabelWithVersion(string(details.Type), utils.LanguageVersions{
		Go:     details.GoVersion,
		Node:   details.NodeVersion,
		Python: details.PythonVersion,
		Ruby:   details.RubyVersion,
		PHP:    details.PHPVersion,
		Java:   details.JavaVersion,
		DotNet: details.DotNetVersion,
	})
}

func selectedServiceIDs(services []serviceChoice, selected map[string]bool) []string {
	return utils.OrderedSelectedIDs(services, func(svc serviceChoice) string {
		return svc.ID
	}, selected)
}
