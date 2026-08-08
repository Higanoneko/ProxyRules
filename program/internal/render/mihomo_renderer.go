package render

import (
	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

var proxyRulesPackPlaceholders = map[string]struct{}{
	"$ProxyRules_Pack": {},
}

var mihomoFullOnlyKeys = []string{
	"mixed-port",
	"allow-lan",
	"ipv6",
	"mode",
	"unified-delay",
	"tcp-concurrent",
	"find-process-mode",
	"global-client-fingerprint",
	"log-level",
	"geodata-loader",
	"external-controller",
	"external-ui",
	"external-ui-url",
	"disable-keep-alive",
	"profile",
	"geo-auto-update",
	"geo-update-interval",
}

var mihomoReplacePaths = []string{"proxy-groups", "rule-providers", "rules"}

type MihomoRenderer struct {
	base         repository.BaseData
	ruleResolver *RuleResolver
}

func NewMihomoRenderer(base repository.BaseData) *MihomoRenderer {
	return &MihomoRenderer{
		base:         base,
		ruleResolver: NewRuleResolver(base),
	}
}

func (r *MihomoRenderer) RenderStandard(plan domain.PolicyPlan, full bool) (string, error) {
	head, err := r.base.Head("mihomo")
	if err != nil {
		return "", err
	}

	rules, err := resolveMihomoRules(head, yamlHeadPlaceholders(plan), r.ruleResolver.MihomoRules(plan.Rules))
	if err != nil {
		return "", err
	}

	explicit, err := r.standardOverrides(plan, full, rules)
	if err != nil {
		return "", err
	}

	dropPaths := []string{}
	if !full {
		dropPaths = mihomoFullOnlyKeys
	}

	document, err := ComposeYAML(
		head,
		explicit,
		yamlHeadPlaceholders(plan),
		dropPaths,
		mihomoReplacePaths,
	)
	if err != nil {
		return "", err
	}

	content, err := renderYAMLDocument(document)
	if err != nil {
		return "", err
	}
	return addBlankLinesBeforeSections(content, []string{"dns:", "sniffer:", "geodata-mode:", "proxy-groups:", "rule-providers:", "rules:"}), nil
}

func (r *MihomoRenderer) FullDefaultsNode(plan domain.PolicyPlan) (*yaml.Node, error) {
	head, err := r.base.Head("mihomo")
	if err != nil {
		return nil, err
	}

	document, err := ComposeYAML(
		head,
		nil,
		yamlHeadPlaceholders(plan),
		[]string{"dns", "sniffer", "geodata-mode", "geox-url", "proxy-groups", "rule-providers", "rules"},
		nil,
	)
	if err != nil {
		return nil, err
	}

	root := document.Content[0]
	deletePath(root, []string{"mixed-port"})
	deletePath(root, []string{"ipv6"})
	return root, nil
}

func (r *MihomoRenderer) standardOverrides(plan domain.PolicyPlan, full bool, rules []string) (*yaml.Node, error) {
	providers, err := r.ruleResolver.MihomoRuleProviders(plan.Rules)
	if err != nil {
		return nil, err
	}

	root := newMappingNode()
	appendMappingValue(root, "dns", mihomoDNSNode(plan, true))
	appendMappingValue(root, "geodata-mode", newScalarNode(true))
	appendMappingValue(root, "geox-url", geoxNode())
	appendMappingValue(root, "proxy-groups", proxyGroupsNode(plan.Proxy.Groups))
	appendMappingValue(root, "rule-providers", providers)
	appendMappingValue(root, "rules", stringSequenceNode(rules))

	if full {
		appendMappingValue(root, "mixed-port", newScalarNode(plan.Ports.Mixed))
		appendMappingValue(root, "ipv6", newScalarNode(plan.DNS.IPv6))
	}

	return root, nil
}

func resolveMihomoRules(head string, placeholders map[string]any, generatedRules []string) ([]string, error) {
	preparedHead, err := applyYAMLPlaceholders(head, placeholders)
	if err != nil {
		return nil, err
	}

	document, err := loadYAMLDocument(preparedHead)
	if err != nil {
		return nil, err
	}

	root := document.Content[0]
	rulesNode := mappingValue(root, "rules")
	if rulesNode == nil || rulesNode.Kind != yaml.SequenceNode {
		return append([]string(nil), generatedRules...), nil
	}

	placeholderFound := false
	resolved := make([]string, 0, len(rulesNode.Content)+len(generatedRules))
	for _, entry := range rulesNode.Content {
		if _, ok := proxyRulesPackPlaceholders[entry.Value]; ok {
			resolved = append(resolved, generatedRules...)
			placeholderFound = true
			continue
		}
		resolved = append(resolved, entry.Value)
	}

	if !placeholderFound {
		return append([]string(nil), generatedRules...), nil
	}
	return resolved, nil
}
