package render

import (
	"bytes"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

func loadYAMLDocument(content string) (*yaml.Node, error) {
	document := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(content), document); err != nil {
		return nil, err
	}
	if len(document.Content) == 0 {
		document.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
	}
	return document, nil
}

func renderYAMLDocument(document *yaml.Node) (string, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func cloneNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	cloned := *node
	if len(node.Content) == 0 {
		cloned.Content = nil
		return &cloned
	}
	cloned.Content = make([]*yaml.Node, len(node.Content))
	for index, child := range node.Content {
		cloned.Content[index] = cloneNode(child)
	}
	return &cloned
}

func newMappingNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

func newSequenceNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}

func newAliasNode(target *yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.AliasNode, Value: target.Anchor, Alias: target}
}

func newScalarNode(value any) *yaml.Node {
	switch typed := value.(type) {
	case nil:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: typed}
	case bool:
		if typed {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(typed)}
	case int64:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(typed, 10)}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprint(typed)}
	}
}

func nodeFromValue(value any) (*yaml.Node, error) {
	content, err := yaml.Marshal(value)
	if err != nil {
		return nil, err
	}
	document := &yaml.Node{}
	if err := yaml.Unmarshal(content, document); err != nil {
		return nil, err
	}
	if len(document.Content) == 0 {
		return newMappingNode(), nil
	}
	return document.Content[0], nil
}

func nodeToValue(node *yaml.Node) (any, error) {
	content, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}
	var value any
	if err := yaml.Unmarshal(content, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func appendMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	node.Content = append(node.Content, newScalarNode(key), value)
}

func prependMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	content := make([]*yaml.Node, 0, len(node.Content)+2)
	content = append(content, newScalarNode(key), value)
	content = append(content, node.Content...)
	node.Content = content
}

func setMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	if index := mappingKeyIndex(node, key); index >= 0 {
		node.Content[index+1] = value
		return
	}
	appendMappingValue(node, key, value)
}

func mappingKeyIndex(node *yaml.Node, key string) int {
	if node == nil || node.Kind != yaml.MappingNode {
		return -1
	}
	for index := 0; index < len(node.Content)-1; index += 2 {
		if node.Content[index].Value == key {
			return index
		}
	}
	return -1
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	index := mappingKeyIndex(node, key)
	if index < 0 {
		return nil
	}
	return node.Content[index+1]
}

func deletePath(node *yaml.Node, segments []string) {
	if node == nil || node.Kind != yaml.MappingNode || len(segments) == 0 {
		return
	}

	index := mappingKeyIndex(node, segments[0])
	if index < 0 {
		return
	}

	if len(segments) == 1 {
		node.Content = append(node.Content[:index], node.Content[index+2:]...)
		return
	}

	deletePath(node.Content[index+1], segments[1:])
}
