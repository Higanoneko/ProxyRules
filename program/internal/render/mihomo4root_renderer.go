package render

import (
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

const mihomo4RootDNSHijackProxy = "DNS_Hijack"

type Mihomo4RootRenderer struct {
	base         repository.BaseData
	ruleResolver *RuleResolver
}

func NewMihomo4RootRenderer(base repository.BaseData) *Mihomo4RootRenderer {
	return &Mihomo4RootRenderer{
		base:         base,
		ruleResolver: NewRuleResolver(base),
	}
}

func (r *Mihomo4RootRenderer) RenderMihomo4Root(plan domain.PolicyPlan, tunEnabled bool) (string, error) {
	head, err := r.base.Head("box4root")
	if err != nil {
		return "", err
	}

	rules, err := resolveMihomoRules(head, yamlHeadPlaceholders(plan), r.ruleResolver.MihomoRules(plan.Rules))
	if err != nil {
		return "", err
	}

	groups := excludeDNSHijackFromGroups(plan.Proxy.Groups, "手动选择", "其他节点")
	explicit, err := r.mihomo4RootOverrides(plan, tunEnabled, rules, groups)
	if err != nil {
		return "", err
	}

	document, err := ComposeYAML(head, explicit, yamlHeadPlaceholders(plan), nil, mihomoReplacePaths)
	if err != nil {
		return "", err
	}
	content, err := renderYAMLDocument(document)
	if err != nil {
		return "", err
	}
	return compactProxyProviderCommentBlock(content), nil
}

func (r *Mihomo4RootRenderer) mihomo4RootOverrides(
	plan domain.PolicyPlan,
	tunEnabled bool,
	rules []string,
	groups []domain.ProxyGroup,
) (*yaml.Node, error) {
	providers, err := r.ruleResolver.MihomoRuleProviders(plan.Rules)
	if err != nil {
		return nil, err
	}

	root := newMappingNode()
	appendMappingValue(root, "ipv6", newScalarNode(plan.DNS.IPv6))
	tun := newMappingNode()
	appendMappingValue(tun, "enable", newScalarNode(tunEnabled))
	appendMappingValue(root, "tun", tun)
	appendMappingValue(root, "dns", mihomoDNSNode(plan, true))
	appendMappingValue(root, "proxy-groups", proxyGroupsNode(groups))
	appendMappingValue(root, "rule-providers", providers)
	appendMappingValue(root, "rules", stringSequenceNode(rules))
	return root, nil
}

func excludeDNSHijackFromGroups(groups []domain.ProxyGroup, names ...string) []domain.ProxyGroup {
	result := make([]domain.ProxyGroup, 0, len(groups))
	for _, group := range groups {
		if containsGroupName(names, group.Name) {
			group.ExcludeFilter = appendExcludeFilter(group.ExcludeFilter, mihomo4RootDNSHijackProxy)
		}
		result = append(result, group)
	}
	return result
}

func containsGroupName(names []string, expected string) bool {
	for _, name := range names {
		if name == expected {
			return true
		}
	}
	return false
}

func appendExcludeFilter(existing string, extra string) string {
	if existing == "" {
		return "(?i)" + extra
	}
	base := strings.TrimSuffix(strings.TrimPrefix(existing, "(?i)"), "|")
	return "(?i)" + base + "|" + extra
}

func compactProxyProviderCommentBlock(content string) string {
	lines := strings.Split(content, "\n")
	start := -1
	for index, line := range lines {
		if strings.Contains(line, "# 多订阅按照下面照葫芦画瓢即可") {
			start = index
			break
		}
	}
	if start < 0 {
		return content
	}

	rewritten := make([]string, 0, len(lines))
	for index, line := range lines {
		if index > start && strings.TrimSpace(line) == "" {
			previous := ""
			if len(rewritten) > 0 {
				previous = rewritten[len(rewritten)-1]
			}
			next := ""
			if index+1 < len(lines) {
				next = lines[index+1]
			}
			if isCommentLine(previous) && isCommentLine(next) {
				continue
			}
		}
		rewritten = append(rewritten, line)
	}
	return strings.Join(rewritten, "\n")
}

func isCommentLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}
