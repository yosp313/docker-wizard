package write

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

func mapsEqual(left map[string]any, right map[string]any) bool {
	leftYAML, leftErr := marshalDeterministicYAML(left)
	rightYAML, rightErr := marshalDeterministicYAML(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftYAML == rightYAML
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
