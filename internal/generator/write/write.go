package write

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"docker-wizard/internal/utils"

	"gopkg.in/yaml.v3"
)

const (
	ComposeFileName      = "docker-compose.yml"
	DockerfileFileName   = "Dockerfile"
	DockerignoreFileName = ".dockerignore"
)

type Output struct {
	ComposePath          string
	ComposeStatus        WriteStatus
	ComposeBackupPath    string
	DockerfilePath       string
	DockerfileStatus     WriteStatus
	DockerfileBackupPath string
	DockerignorePath     string
	DockerignoreStatus   WriteStatus
}

type WriteStatus string

const (
	WriteStatusCreated   WriteStatus = "created"
	WriteStatusUpdated   WriteStatus = "updated"
	WriteStatusUnchanged WriteStatus = "unchanged"
)

func WriteFiles(root string, compose string, dockerfile string) (Output, error) {
	if root == "" {
		return Output{}, fmt.Errorf("root directory is required")
	}

	composePath := filepath.Join(root, ComposeFileName)
	dockerfilePath := filepath.Join(root, DockerfileFileName)
	dockerignorePath := filepath.Join(root, DockerignoreFileName)

	output := Output{
		ComposePath:      composePath,
		DockerfilePath:   dockerfilePath,
		DockerignorePath: dockerignorePath,
	}

	composeStatus, composeBackup, err := writeManagedFile(root, composePath, "docker-compose-*.tmp", compose, MergeCompose)
	if err != nil {
		return Output{}, err
	}
	output.ComposeStatus = composeStatus
	output.ComposeBackupPath = composeBackup

	dockerfileStatus, dockerfileBackup, err := writeManagedFile(root, dockerfilePath, "dockerfile-*.tmp", dockerfile, MergeDockerfile)
	if err != nil {
		return Output{}, err
	}
	output.DockerfileStatus = dockerfileStatus
	output.DockerfileBackupPath = dockerfileBackup

	if !utils.FileExists(dockerignorePath) {
		if err := os.WriteFile(dockerignorePath, []byte(DefaultDockerignore()), 0644); err != nil {
			return Output{}, fmt.Errorf("write dockerignore: %w", err)
		}
		output.DockerignoreStatus = WriteStatusCreated
	} else {
		output.DockerignoreStatus = WriteStatusUnchanged
	}

	return output, nil
}

type mergeFunc func(existing string, generated string) (string, error)

func writeManagedFile(root string, path string, pattern string, content string, merge mergeFunc) (WriteStatus, string, error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return "", "", fmt.Errorf("%s is a directory", filepath.Base(path))
		}
		existing, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", "", fmt.Errorf("read %s: %w", filepath.Base(path), readErr)
		}

		targetContent := content
		if merge != nil {
			mergedContent, mergeErr := merge(string(existing), content)
			if mergeErr != nil {
				return "", "", fmt.Errorf("merge %s: %w", filepath.Base(path), mergeErr)
			}
			targetContent = mergedContent
		}

		if string(existing) == targetContent {
			return WriteStatusUnchanged, "", nil
		}

		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, existing, info.Mode().Perm()); err != nil {
			return "", "", fmt.Errorf("write %s backup: %w", filepath.Base(path), err)
		}

		if err := writeByTemp(root, path, pattern, targetContent); err != nil {
			return "", "", err
		}

		return WriteStatusUpdated, backupPath, nil
	}
	if !os.IsNotExist(err) {
		return "", "", fmt.Errorf("stat %s: %w", filepath.Base(path), err)
	}

	if err := writeByTemp(root, path, pattern, content); err != nil {
		return "", "", err
	}

	return WriteStatusCreated, "", nil
}

func writeByTemp(root string, path string, pattern string, content string) error {
	tempPath, err := writeTempFile(root, pattern, content)
	if err != nil {
		return fmt.Errorf("write %s temp file: %w", filepath.Base(path), err)
	}
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("move %s into place: %w", filepath.Base(path), err)
	}

	return nil
}

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

func mapsEqual(left map[string]any, right map[string]any) bool {
	leftYAML, leftErr := marshalDeterministicYAML(left)
	rightYAML, rightErr := marshalDeterministicYAML(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftYAML == rightYAML
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
	if isServiceField(path, "depends_on") || isServiceField(path, "networks") || isServiceField(path, "volumes") {
		return mergeSetLikeList(existing, generated)
	}
	return mergeSetLikeList(existing, generated)
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

func marshalDeterministicYAML(value any) (string, error) {
	root := &yaml.Node{Kind: yaml.DocumentNode}
	root.Content = append(root.Content, toYAMLNode(value))

	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(root); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func toYAMLNode(value any) *yaml.Node {
	switch typed := value.(type) {
	case map[string]any:
		node := &yaml.Node{Kind: yaml.MappingNode}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key})
			node.Content = append(node.Content, toYAMLNode(typed[key]))
		}
		return node
	case []any:
		node := &yaml.Node{Kind: yaml.SequenceNode}
		for _, item := range typed {
			node.Content = append(node.Content, toYAMLNode(item))
		}
		return node
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: typed}
	case bool:
		if typed {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", typed)}
	case int64:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", typed)}
	case float64:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: fmt.Sprintf("%v", typed)}
	case nil:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprintf("%v", typed)}
	}
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

func DefaultDockerignore() string {
	return "" +
		".git\n" +
		".gitignore\n" +
		"node_modules\n" +
		"vendor\n" +
		"bin\n" +
		"dist\n" +
		"build\n" +
		"tmp\n"
}

func writeTempFile(root string, pattern string, content string) (string, error) {
	tempFile, err := os.CreateTemp(root, pattern)
	if err != nil {
		return "", err
	}

	path := tempFile.Name()
	if _, err := tempFile.WriteString(content); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := tempFile.Chmod(0644); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return path, nil
}
