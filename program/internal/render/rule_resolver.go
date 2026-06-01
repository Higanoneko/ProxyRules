package render

import (
	"fmt"
	"path"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

type RuleResolver struct {
	base repository.BaseData
}

func NewRuleResolver(base repository.BaseData) *RuleResolver {
	return &RuleResolver{base: base}
}

func (r *RuleResolver) MihomoRuleProviders(bindings []domain.RuleBinding) (*yaml.Node, error) {
	providers := newMappingNode()
	for _, binding := range bindings {
		url, err := r.ResolveURL(domain.TargetMihomo, binding)
		if err != nil {
			return nil, err
		}
		formatType := detectFormatType(url)
		provider := newMappingNode()
		appendMappingValue(provider, "type", newScalarNode("http"))
		appendMappingValue(provider, "behavior", newScalarNode(binding.Source.Behavior))
		appendMappingValue(provider, "format", newScalarNode(formatType))
		appendMappingValue(provider, "interval", newScalarNode(86400))
		appendMappingValue(provider, "url", newScalarNode(url))
		appendMappingValue(provider, "path", newScalarNode("./ruleset/"+binding.ProviderName+"."+providerPathExt(formatType)))
		appendMappingValue(providers, binding.RuleID, provider)
	}
	return providers, nil
}

func (r *RuleResolver) MihomoRules(bindings []domain.RuleBinding) []string {
	rules := make([]string, 0, len(bindings)+2)
	for _, binding := range bindings {
		rules = append(rules, "RULE-SET,"+binding.RuleID+","+binding.PolicyName)
	}
	rules = append(rules, "GEOIP,CN,直接连接", "MATCH,选择代理")
	return rules
}

func (r *RuleResolver) LoonRemoteRules(bindings []domain.RuleBinding) ([]string, error) {
	lines := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		url, err := r.ResolveURL(domain.TargetLoon, binding)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("%s, policy = %s, tag = %s, enabled = true", url, binding.PolicyName, binding.TagName))
	}
	return lines, nil
}

func (r *RuleResolver) SurgeRemoteRules(bindings []domain.RuleBinding) ([]string, error) {
	lines := make([]string, 0, len(bindings)*3)
	for index, binding := range bindings {
		url, err := r.ResolveURL(domain.TargetSurge, binding)
		if err != nil {
			return nil, err
		}

		if index > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "# "+binding.TagName)
		if binding.SurgeOption != "" {
			lines = append(lines, fmt.Sprintf("RULE-SET,%s,%s,%s", url, binding.PolicyName, binding.SurgeOption))
			continue
		}
		lines = append(lines, fmt.Sprintf("RULE-SET,%s,%s", url, binding.PolicyName))
	}
	return lines, nil
}

func (r *RuleResolver) ResolveURL(target domain.Target, binding domain.RuleBinding) (string, error) {
	categoryConfig, ok := r.base.Categories[binding.Source.Category]
	if !ok {
		return "", fmt.Errorf("unknown category: %s", binding.Source.Category)
	}

	urlTemplate := categoryConfig.URL
	toolKey := toolKeyForTarget(target)
	mappedTool := mappedValue(r.base.ToolMappings[binding.Source.Category], toolKey, toolKey)
	remoteFile := strings.TrimPrefix(binding.Source.RemoteFile, "./")
	fileType := ""

	if strings.Contains(urlTemplate, "{filetype}") {
		if extension := path.Ext(remoteFile); extension != "" {
			remoteFile = strings.TrimSuffix(remoteFile, extension)
		}
		fileType = mappedValue(r.base.FileTypeMappings[binding.Source.Category], toolKey, "")
	}

	result := strings.ReplaceAll(urlTemplate, "{proxytools}", mappedTool)
	result = strings.ReplaceAll(result, "{remotefile}", remoteFile)
	result = strings.ReplaceAll(result, "{filetype}", fileType)
	return result, nil
}

func toolKeyForTarget(target domain.Target) string {
	switch target {
	case domain.TargetMihomo, domain.TargetStash:
		return "Mihomo"
	case domain.TargetLoon:
		return "Loon"
	default:
		return "Surge"
	}
}

func mappedValue(mapping map[string]string, key string, fallback string) string {
	if mapping == nil {
		return fallback
	}
	if value, ok := mapping[key]; ok {
		return value
	}
	if value, ok := mapping["fallback"]; ok {
		return value
	}
	return fallback
}

func detectFormatType(url string) string {
	switch {
	case strings.HasSuffix(url, ".mrs"):
		return "mrs"
	case strings.HasSuffix(url, ".yaml"), strings.HasSuffix(url, ".yml"):
		return "yaml"
	default:
		return "text"
	}
}

func providerPathExt(formatType string) string {
	switch formatType {
	case "yaml":
		return "yaml"
	case "mrs":
		return "mrs"
	default:
		return "list"
	}
}
