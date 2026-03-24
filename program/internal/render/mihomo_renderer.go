package render

import (
	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

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

var mihomoReplacePaths = []string{"dns", "proxy-groups", "rule-providers", "rules"}

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

	explicit, err := r.standardOverrides(plan, full)
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

func (r *MihomoRenderer) RenderBox4Root(plan domain.PolicyPlan, tunEnabled bool) (string, error) {
	head, err := r.base.Head("mihomo_tun")
	if err != nil {
		return "", err
	}

	explicit, err := r.box4RootOverrides(plan, tunEnabled)
	if err != nil {
		return "", err
	}

	document, err := ComposeYAML(head, explicit, yamlHeadPlaceholders(plan), nil, mihomoReplacePaths)
	if err != nil {
		return "", err
	}
	return renderYAMLDocument(document)
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

func (r *MihomoRenderer) standardOverrides(plan domain.PolicyPlan, full bool) (*yaml.Node, error) {
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
	appendMappingValue(root, "rules", stringSequenceNode(r.ruleResolver.MihomoRules(plan.Rules)))

	if full {
		appendMappingValue(root, "mixed-port", newScalarNode(plan.Ports.Mixed))
		appendMappingValue(root, "ipv6", newScalarNode(plan.DNS.IPv6))
	}

	return root, nil
}

func (r *MihomoRenderer) box4RootOverrides(plan domain.PolicyPlan, tunEnabled bool) (*yaml.Node, error) {
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
	appendMappingValue(root, "proxy-groups", proxyGroupsNode(plan.Proxy.Groups))
	appendMappingValue(root, "rule-providers", providers)
	appendMappingValue(root, "rules", stringSequenceNode(r.ruleResolver.MihomoRules(plan.Rules)))
	return root, nil
}
