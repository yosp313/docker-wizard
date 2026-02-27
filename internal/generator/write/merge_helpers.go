package write

import (
	"reflect"
	"strings"
)

func environmentKey(value any) (string, bool) {
	entry, ok := value.(string)
	if !ok {
		return "", false
	}
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return "", false
	}
	if idx := strings.Index(entry, "="); idx >= 0 {
		key := strings.TrimSpace(entry[:idx])
		if key == "" {
			return "", false
		}
		return key, true
	}
	return entry, true
}

func portHostKey(value any) (string, bool) {
	entry, ok := value.(string)
	if !ok {
		return "", false
	}
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return "", false
	}

	withoutProtocol := strings.SplitN(entry, "/", 2)[0]
	parts := strings.Split(withoutProtocol, ":")
	if len(parts) < 2 {
		return "", false
	}
	host := strings.TrimSpace(parts[len(parts)-2])
	if host == "" {
		return "", false
	}
	return host, true
}

func containsDeepValue(values []any, target any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, target) {
			return true
		}
	}
	return false
}

func deepCopy(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, nested := range typed {
			out[key] = deepCopy(nested)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = deepCopy(typed[i])
		}
		return out
	default:
		return typed
	}
}
