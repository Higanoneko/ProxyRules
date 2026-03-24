package render

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func ComposeYAML(
	templateContent string,
	explicitData *yaml.Node,
	placeholders map[string]any,
	dropPaths []string,
	replacePaths []string,
) (*yaml.Node, error) {
	preparedTemplate, err := applyYAMLPlaceholders(templateContent, placeholders)
	if err != nil {
		return nil, err
	}

	document, err := loadYAMLDocument(preparedTemplate)
	if err != nil {
		return nil, err
	}
	root := document.Content[0]

	for _, dropPath := range dropPaths {
		deletePath(root, strings.Split(dropPath, "."))
	}

	if explicitData == nil {
		return document, nil
	}

	explicitRoot := explicitData
	if explicitRoot.Kind == yaml.DocumentNode && len(explicitRoot.Content) > 0 {
		explicitRoot = explicitRoot.Content[0]
	}

	document.Content[0] = mergeYAMLNodes(root, explicitRoot, replacePathSet(replacePaths), "")
	return document, nil
}

func applyYAMLPlaceholders(templateContent string, placeholders map[string]any) (string, error) {
	preparedContent := templateContent
	for key, value := range placeholders {
		serialized, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		replacement := string(serialized)
		for _, token := range []string{
			`"$` + key + `"`,
			`'$` + key + `'`,
			"{" + key + "}",
			"$" + key,
		} {
			preparedContent = strings.ReplaceAll(preparedContent, token, replacement)
		}
	}
	return preparedContent, nil
}

func mergeYAMLNodes(baseNode *yaml.Node, explicitNode *yaml.Node, replacePaths map[string]struct{}, currentPath string) *yaml.Node {
	if baseNode == nil {
		return cloneNode(explicitNode)
	}
	if explicitNode == nil {
		return cloneNode(baseNode)
	}
	if baseNode.Kind != yaml.MappingNode || explicitNode.Kind != yaml.MappingNode {
		return cloneNode(explicitNode)
	}

	merged := cloneNode(baseNode)
	for index := 0; index < len(explicitNode.Content)-1; index += 2 {
		key := explicitNode.Content[index].Value
		nextPath := key
		if currentPath != "" {
			nextPath = currentPath + "." + key
		}

		explicitValue := explicitNode.Content[index+1]
		existingIndex := mappingKeyIndex(merged, key)
		if _, replace := replacePaths[nextPath]; replace || existingIndex < 0 {
			if existingIndex >= 0 {
				merged.Content[existingIndex+1] = cloneNode(explicitValue)
				continue
			}
			appendMappingValue(merged, key, cloneNode(explicitValue))
			continue
		}

		merged.Content[existingIndex+1] = mergeYAMLNodes(
			merged.Content[existingIndex+1],
			explicitValue,
			replacePaths,
			nextPath,
		)
	}
	return merged
}

func replacePathSet(paths []string) map[string]struct{} {
	result := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		result[path] = struct{}{}
	}
	return result
}
