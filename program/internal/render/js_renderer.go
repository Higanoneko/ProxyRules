package render

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

//go:embed templates/mihomo_runtime.js.tpl
var mihomoRuntimeTemplate string

type MihomoScriptRenderer struct {
	base           repository.BaseData
	ruleResolver   *RuleResolver
	mihomoRenderer *MihomoRenderer
}

func NewMihomoScriptRenderer(base repository.BaseData) *MihomoScriptRenderer {
	return &MihomoScriptRenderer{
		base:           base,
		ruleResolver:   NewRuleResolver(base),
		mihomoRenderer: NewMihomoRenderer(base),
	}
}

func (r *MihomoScriptRenderer) RenderArgs(plan domain.PolicyPlan) (string, error) {
	return r.render(plan, argsParameterBlock())
}

func (r *MihomoScriptRenderer) RenderFixed(plan domain.PolicyPlan, full bool, dnsEnabled bool) (string, error) {
	return r.render(plan, fixedParameterBlock(plan.DNS.IPv6, full, dnsEnabled))
}

func (r *MihomoScriptRenderer) render(plan domain.PolicyPlan, parameterBlock string) (string, error) {
	providersNode, err := r.ruleResolver.MihomoRuleProviders(plan.Rules)
	if err != nil {
		return "", err
	}
	ruleProvidersJSON, err := jsonFromNode(providersNode)
	if err != nil {
		return "", err
	}

	fullDefaultsNode, err := r.mihomoRenderer.FullDefaultsNode(plan)
	if err != nil {
		return "", err
	}
	fullDefaultsJSON, err := jsonFromNode(fullDefaultsNode)
	if err != nil {
		return "", err
	}

	snifferNode, err := nodeFromValue(plan.Sniffer)
	if err != nil {
		return "", err
	}
	snifferJSON, err := jsonFromNode(snifferNode)
	if err != nil {
		return "", err
	}

	geoXJSON, err := jsonFromNode(geoxNode())
	if err != nil {
		return "", err
	}

	dns := projectGenericDNS(plan.DNS)

	dnsIPJSON, err := json.Marshal(dns.BootstrapResolvers)
	if err != nil {
		return "", err
	}
	dnsTemplateNode, err := r.dnsTemplateNode(plan)
	if err != nil {
		return "", err
	}
	dnsTemplateJSON, err := jsonFromNode(dnsTemplateNode)
	if err != nil {
		return "", err
	}
	rulesJSON, err := json.Marshal(r.ruleResolver.MihomoRules(plan.Rules))
	if err != nil {
		return "", err
	}
	policyGroupsJSON, err := json.Marshal(r.base.PolicyConfig.Groups)
	if err != nil {
		return "", err
	}
	nodeExcludeJSON, err := json.Marshal(r.base.PolicyConfig.NodeExcludePattern)
	if err != nil {
		return "", err
	}

	replacer := strings.NewReplacer(
		"__PARAMETER_BLOCK__", parameterBlock,
		"__DNS_BOOTSTRAP_LIST__", string(dnsIPJSON),
		"__DNS_TEMPLATE__", dnsTemplateJSON,
		"__MIXED_PORT__", fmt.Sprintf("%d", plan.Ports.Mixed),
		"__FULL_DEFAULTS__", fullDefaultsJSON,
		"__RULE_PROVIDERS__", ruleProvidersJSON,
		"__RULES__", string(rulesJSON),
		"__SNIFFER__", snifferJSON,
		"__GEOX_URL__", geoXJSON,
		"__POLICY_GROUPS__", string(policyGroupsJSON),
		"__NODE_EXCLUDE_PATTERN__", string(nodeExcludeJSON),
	)
	return replacer.Replace(mihomoRuntimeTemplate), nil
}

func (r *MihomoScriptRenderer) dnsTemplateNode(plan domain.PolicyPlan) (*yaml.Node, error) {
	head, err := r.base.Head("mihomo")
	if err != nil {
		return nil, err
	}

	explicitRoot := newMappingNode()
	appendMappingValue(explicitRoot, "dns", mihomoDNSNode(plan, true))

	document, err := ComposeYAML(head, explicitRoot, yamlHeadPlaceholders(plan), nil, nil)
	if err != nil {
		return nil, err
	}

	dnsNode := mappingValue(document.Content[0], "dns")
	if dnsNode == nil {
		return newMappingNode(), nil
	}
	return dnsNode, nil
}

func argsParameterBlock() string {
	return `function buildFeatureFlags(args) {
    const flags = {
        ipv6Enabled: true,
        fullConfig: false,
        dnsEnabled: true,
        countryThreshold: 0,
    };

    if (args && Object.prototype.hasOwnProperty.call(args, "ipv6")) {
        flags.ipv6Enabled = parseBool(args.ipv6);
    }
    if (args && Object.prototype.hasOwnProperty.call(args, "full")) {
        flags.fullConfig = parseBool(args.full);
    }
    if (args && Object.prototype.hasOwnProperty.call(args, "dns")) {
        flags.dnsEnabled = parseBool(args.dns);
    }
    if (args && Object.prototype.hasOwnProperty.call(args, "threshold")) {
        flags.countryThreshold = parseNumber(args.threshold, 0);
    }

    return flags;
}

const rawArgs = typeof $arguments !== "undefined" ? $arguments : {};
const {
    ipv6Enabled,
    fullConfig,
    dnsEnabled,
    countryThreshold,
} = buildFeatureFlags(rawArgs);`
}

func fixedParameterBlock(ipv6Enabled bool, fullConfig bool, dnsEnabled bool) string {
	return fmt.Sprintf(`// ============================================
// 参数定义区域（可根据需要修改）
// ============================================
const ipv6Enabled = %t;
const fullConfig = %t;
const dnsEnabled = %t;
const countryThreshold = 0;
// ============================================
// 参数定义区域结束
// ============================================`, ipv6Enabled, fullConfig, dnsEnabled)
}
