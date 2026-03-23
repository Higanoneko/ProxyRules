package render

import (
	"gopkg.in/yaml.v3"

	"github.com/PianCat/ProxyRules/internal/domain"
)

func proxyGroupsNode(groups []domain.ProxyGroup) *yaml.Node {
	sequence := newSequenceNode()
	anchors := map[string]*yaml.Node{}

	for _, group := range groups {
		groupNode := newMappingNode()
		appendMappingValue(groupNode, "name", newScalarNode(group.Name))
		if group.Icon != "" {
			appendMappingValue(groupNode, "icon", newScalarNode(group.Icon))
		}
		if group.IncludeAll {
			appendMappingValue(groupNode, "include-all", newScalarNode(true))
		}
		if group.Filter != "" {
			appendMappingValue(groupNode, "filter", newScalarNode(group.Filter))
		}
		if group.ExcludeFilter != "" {
			appendMappingValue(groupNode, "exclude-filter", newScalarNode(group.ExcludeFilter))
		}
		appendMappingValue(groupNode, "type", newScalarNode(group.Type))
		if group.URL != "" {
			appendMappingValue(groupNode, "url", newScalarNode(group.URL))
		}
		if group.Interval > 0 {
			appendMappingValue(groupNode, "interval", newScalarNode(group.Interval))
		}
		if group.Tolerance > 0 {
			appendMappingValue(groupNode, "tolerance", newScalarNode(group.Tolerance))
		}
		if group.Lazy != nil {
			appendMappingValue(groupNode, "lazy", newScalarNode(*group.Lazy))
		}
		if len(group.Proxies) > 0 || group.ProxiesAlias != "" {
			appendMappingValue(groupNode, "proxies", buildProxySequenceNode(group, anchors))
		}
		sequence.Content = append(sequence.Content, groupNode)
	}

	return sequence
}

func buildProxySequenceNode(group domain.ProxyGroup, anchors map[string]*yaml.Node) *yaml.Node {
	if group.ProxiesAlias != "" {
		if target, ok := anchors[group.ProxiesAlias]; ok {
			return newAliasNode(target)
		}
	}

	sequence := stringSequenceNode(group.Proxies)
	if group.ProxiesAnchor != "" {
		sequence.Anchor = group.ProxiesAnchor
		anchors[group.ProxiesAnchor] = sequence
	}
	return sequence
}

func sequenceFromNodes(nodes ...*yaml.Node) *yaml.Node {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	sequence.Content = append(sequence.Content, nodes...)
	return sequence
}
