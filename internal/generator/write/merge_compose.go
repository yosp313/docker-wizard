package write

import "gopkg.in/yaml.v3"

func MergeCompose(existing string, generated string) (string, error) {
	existingDoc := map[string]any{}
	if err := yaml.Unmarshal([]byte(existing), &existingDoc); err != nil {
		return "", err
	}
	generatedDoc := map[string]any{}
	if err := yaml.Unmarshal([]byte(generated), &generatedDoc); err != nil {
		return "", err
	}

	merged := deepMergeComposeValue(nil, existingDoc, generatedDoc)
	mergedMap, ok := merged.(map[string]any)
	if !ok {
		return generated, nil
	}

	if mapsEqual(existingDoc, mergedMap) {
		return existing, nil
	}

	output, err := marshalDeterministicYAML(mergedMap)
	if err != nil {
		return "", err
	}
	return output, nil
}

func deepMergeComposeValue(path []string, existing any, generated any) any {
	switch existingTyped := existing.(type) {
	case map[string]any:
		generatedMap, ok := generated.(map[string]any)
		if !ok {
			return deepCopy(existing)
		}
		out := make(map[string]any, len(existingTyped))
		for key, value := range existingTyped {
			out[key] = deepCopy(value)
		}
		for key, generatedValue := range generatedMap {
			existingValue, found := out[key]
			if !found {
				out[key] = deepCopy(generatedValue)
				continue
			}
			out[key] = deepMergeComposeValue(appendComposePath(path, key), existingValue, generatedValue)
		}
		return out
	case []any:
		generatedSlice, ok := generated.([]any)
		if !ok {
			return deepCopy(existing)
		}
		return mergeComposeSlice(path, existingTyped, generatedSlice)
	default:
		if existing == nil {
			return deepCopy(generated)
		}
		return deepCopy(existing)
	}
}

func appendComposePath(path []string, key string) []string {
	next := make([]string, 0, len(path)+1)
	next = append(next, path...)
	next = append(next, key)
	return next
}

func mergeComposeSlice(path []string, existing []any, generated []any) []any {
	if isServiceField(path, "environment") {
		return mergeEnvironmentList(existing, generated)
	}
	if isServiceField(path, "ports") {
		return mergePortsList(existing, generated)
	}
	if isServiceField(path, "command") || isServiceField(path, "entrypoint") {
		return mergeUserPriorityList(existing, generated)
	}
	if isServiceField(path, "depends_on") || isServiceField(path, "networks") || isServiceField(path, "volumes") {
		return mergeSetLikeList(existing, generated)
	}
	return mergeSetLikeList(existing, generated)
}

func mergeUserPriorityList(existing []any, generated []any) []any {
	if len(existing) > 0 {
		out := make([]any, 0, len(existing))
		for _, value := range existing {
			out = append(out, deepCopy(value))
		}
		return out
	}

	out := make([]any, 0, len(generated))
	for _, value := range generated {
		out = append(out, deepCopy(value))
	}
	return out
}

func isServiceField(path []string, field string) bool {
	return len(path) == 3 && path[0] == "services" && path[2] == field
}

func mergeEnvironmentList(existing []any, generated []any) []any {
	out := make([]any, 0, len(existing)+len(generated))
	seenKeys := map[string]bool{}

	for _, value := range existing {
		out = append(out, deepCopy(value))
		if key, ok := environmentKey(value); ok {
			seenKeys[key] = true
		}
	}

	for _, generatedValue := range generated {
		if key, ok := environmentKey(generatedValue); ok {
			if seenKeys[key] {
				continue
			}
			seenKeys[key] = true
			out = append(out, deepCopy(generatedValue))
			continue
		}
		if containsDeepValue(out, generatedValue) {
			continue
		}
		out = append(out, deepCopy(generatedValue))
	}

	return out
}

func mergePortsList(existing []any, generated []any) []any {
	out := make([]any, 0, len(existing)+len(generated))
	hostPorts := map[string]bool{}

	for _, value := range existing {
		out = append(out, deepCopy(value))
		if host, ok := portHostKey(value); ok {
			hostPorts[host] = true
		}
	}

	for _, generatedValue := range generated {
		if host, ok := portHostKey(generatedValue); ok {
			if hostPorts[host] {
				continue
			}
			hostPorts[host] = true
			out = append(out, deepCopy(generatedValue))
			continue
		}
		if containsDeepValue(out, generatedValue) {
			continue
		}
		out = append(out, deepCopy(generatedValue))
	}

	return out
}

func mergeSetLikeList(existing []any, generated []any) []any {
	out := make([]any, 0, len(existing)+len(generated))
	for _, value := range existing {
		out = append(out, deepCopy(value))
	}
	for _, generatedValue := range generated {
		if containsDeepValue(out, generatedValue) {
			continue
		}
		out = append(out, deepCopy(generatedValue))
	}
	return out
}
