package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

func jsonFromNode(node *yaml.Node) (string, error) {
	var buffer bytes.Buffer
	if err := writeJSONNode(&buffer, node); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func writeJSONNode(buffer *bytes.Buffer, node *yaml.Node) error {
	if node == nil {
		buffer.WriteString("null")
		return nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			buffer.WriteString("null")
			return nil
		}
		return writeJSONNode(buffer, node.Content[0])
	case yaml.MappingNode:
		buffer.WriteByte('{')
		for index := 0; index < len(node.Content)-1; index += 2 {
			if index > 0 {
				buffer.WriteByte(',')
			}
			keyJSON, err := json.Marshal(node.Content[index].Value)
			if err != nil {
				return err
			}
			buffer.Write(keyJSON)
			buffer.WriteByte(':')
			if err := writeJSONNode(buffer, node.Content[index+1]); err != nil {
				return err
			}
		}
		buffer.WriteByte('}')
		return nil
	case yaml.SequenceNode:
		buffer.WriteByte('[')
		for index, child := range node.Content {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeJSONNode(buffer, child); err != nil {
				return err
			}
		}
		buffer.WriteByte(']')
		return nil
	case yaml.AliasNode:
		return writeJSONNode(buffer, node.Alias)
	case yaml.ScalarNode:
		return writeJSONScalar(buffer, node)
	default:
		return fmt.Errorf("unsupported yaml kind: %d", node.Kind)
	}
}

func writeJSONScalar(buffer *bytes.Buffer, node *yaml.Node) error {
	switch node.Tag {
	case "!!bool":
		if node.Value == "true" || node.Value == "false" {
			buffer.WriteString(node.Value)
			return nil
		}
	case "!!int":
		if _, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
			buffer.WriteString(node.Value)
			return nil
		}
	case "!!null":
		buffer.WriteString("null")
		return nil
	}

	valueJSON, err := json.Marshal(node.Value)
	if err != nil {
		return err
	}
	buffer.Write(valueJSON)
	return nil
}
