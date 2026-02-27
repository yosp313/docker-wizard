package write

import "strings"

func MergeDockerfile(existing string, generated string) (string, error) {
	merged := strings.TrimRight(existing, "\n")
	if merged == "" {
		merged = existing
	}

	envLine := firstLineWithPrefix(generated, "ENV APP_START_CMD=")
	if envLine != "" && !containsLine(existing, envLine) {
		if merged != "" {
			merged += "\n"
		}
		merged += envLine
	}

	cmdLine := `CMD ["sh", "-lc", "$APP_START_CMD"]`
	hasAnyCMD := hasDirective(existing, "CMD ")
	if !hasAnyCMD && containsLine(generated, cmdLine) && !containsLine(merged, cmdLine) {
		if merged != "" {
			merged += "\n"
		}
		merged += cmdLine
	}

	if !strings.HasSuffix(merged, "\n") {
		merged += "\n"
	}
	return merged, nil
}

func firstLineWithPrefix(content string, prefix string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return trimmed
		}
	}
	return ""
}

func containsLine(content string, line string) bool {
	for _, existing := range strings.Split(content, "\n") {
		if strings.TrimSpace(existing) == line {
			return true
		}
	}
	return false
}

func hasDirective(content string, directive string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), directive) {
			return true
		}
	}
	return false
}
