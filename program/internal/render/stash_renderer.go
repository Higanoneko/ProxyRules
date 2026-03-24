package render

import (
	"strings"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

type StashRenderer struct {
	base         repository.BaseData
	ruleResolver *RuleResolver
}

func NewStashRenderer(base repository.BaseData) *StashRenderer {
	return &StashRenderer{
		base:         base,
		ruleResolver: NewRuleResolver(base),
	}
}

func (r *StashRenderer) RenderFull(plan domain.PolicyPlan) (string, error) {
	document, err := r.compose(plan)
	if err != nil {
		return "", err
	}

	content, err := renderYAMLDocument(document)
	if err != nil {
		return "", err
	}
	content = addBlankLinesBeforeSections(content, []string{"dns:", "proxy-providers:", "proxies:", "proxy-groups:", "rule-providers:", "rules:"})
	return injectStashSubscriptionPlaceholders(content), nil
}

func (r *StashRenderer) RenderOverride(plan domain.PolicyPlan) (string, error) {
	document, err := r.compose(plan)
	if err != nil {
		return "", err
	}

	root := document.Content[0]
	metadata := []struct {
		key   string
		value string
	}{
		{key: "category", value: "Override"},
		{key: "icon", value: "https://fastly.jsdelivr.net/gh/shindgewongxj/WHATSINStash@master/icon/substore.png"},
		{key: "author", value: "PianCat"},
		{key: "desc", value: "PianCat's Config Override"},
		{key: "name", value: "PianCat Stash Override"},
	}
	for _, entry := range metadata {
		prependMappingValue(root, entry.key, newScalarNode(entry.value))
	}

	content, err := renderYAMLDocument(document)
	if err != nil {
		return "", err
	}
	content = addBlankLinesBeforeSections(content, []string{"dns:", "proxy-providers:", "proxies:", "proxy-groups:", "rule-providers:", "rules:"})
	content = strings.Replace(content, "dns:", "dns: #!replace", 1)
	content = strings.Replace(content, "proxy-groups:", "proxy-groups: #!replace", 1)
	content = strings.Replace(content, "rule-providers:", "rule-providers: #!replace", 1)
	content = strings.Replace(content, "rules:", "rules: #!replace", 1)
	return content, nil
}

func (r *StashRenderer) compose(plan domain.PolicyPlan) (*yaml.Node, error) {
	head, err := r.base.Head("stash")
	if err != nil {
		return nil, err
	}

	providers, err := r.ruleResolver.MihomoRuleProviders(plan.Rules)
	if err != nil {
		return nil, err
	}

	root := newMappingNode()
	appendMappingValue(root, "dns", mihomoDNSNode(plan, false))
	appendMappingValue(root, "proxy-groups", proxyGroupsNode(filterGroups(plan.Proxy.Groups, "GLOBAL")))
	appendMappingValue(root, "rule-providers", providers)
	appendMappingValue(root, "rules", stringSequenceNode(r.ruleResolver.MihomoRules(plan.Rules)))

	return ComposeYAML(head, root, yamlHeadPlaceholders(plan), nil, []string{"proxy-groups", "rule-providers", "rules"})
}

func injectStashSubscriptionPlaceholders(content string) string {
	if strings.Contains(content, "\nproxy-providers:\n") || strings.Contains(content, "\nproxies:\n") {
		return content
	}

	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)+8)
	inserted := false
	for _, line := range lines {
		if !inserted && strings.HasPrefix(strings.TrimSpace(line), "proxy-groups:") {
			result = append(result,
				"proxy-providers:",
				"# Your Subscription Proxy Providers Here",
				"# Example:",
				"#   url: https://example.com/proxy.yaml",
				"#   interval: 600",
				"",
				"proxies:",
				"",
			)
			inserted = true
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func filterGroups(groups []domain.ProxyGroup, excludedName string) []domain.ProxyGroup {
	filtered := make([]domain.ProxyGroup, 0, len(groups))
	for _, group := range groups {
		if group.Name == excludedName {
			continue
		}
		filtered = append(filtered, group)
	}
	return filtered
}
