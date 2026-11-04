package render

import (
	"fmt"

	"github.com/luluape/ProxyRules/internal/domain"
	"gopkg.in/yaml.v3"
)

// RenderFull composes the shared policy groups and ordered
// remote rules into one native Egern profile.
func (r *EgernRenderer) RenderFull(
	plan domain.PolicyPlan,
	subscriptionURLs []string,
) (string, error) {
	groupYAML, err := r.RenderPolicyGroups(plan, subscriptionURLs)
	if err != nil {
		return "", fmt.Errorf("Egern policy groups: %w", err)
	}

	ruleYAML, err := r.RenderRuleSets(plan)
	if err != nil {
		return "", fmt.Errorf("Egern rule sets: %w", err)
	}

	var groups struct {
		PolicyGroups []map[string]egernPolicyGroup `yaml:"policy_groups"`
	}
	if err := yaml.Unmarshal([]byte(groupYAML), &groups); err != nil {
		return "", err
	}

	var rules egernRuleSetDocument
	if err := yaml.Unmarshal([]byte(ruleYAML), &rules); err != nil {
		return "", err
	}

	if len(groups.PolicyGroups) == 0 || len(rules.Rules) == 0 {
		return "", fmt.Errorf("Egern configuration is incomplete")
	}

	type finalRule struct {
		Default *struct {
			Policy string `yaml:"policy"`
		} `yaml:"default,omitempty"`
		RuleSet *egernRuleSetBody `yaml:"rule_set,omitempty"`
	}

	finalRules := make([]finalRule, 0, len(rules.Rules)+1)

	for _, entry := range rules.Rules {
		finalRules = append(finalRules, finalRule{
			RuleSet: entry.RuleSet,
		})
	}

	finalRules = append(finalRules, finalRule{
		Default: &struct {
			Policy string `yaml:"policy"`
		}{
			Policy: "选择代理",
		},
	})

	document := struct {
		DNS struct {
			Bootstrap        []string `yaml:"bootstrap"`
			ProxyNameservers []string `yaml:"proxy_nameservers"`
		} `yaml:"dns"`
		IPv6         bool                          `yaml:"ipv6"`
		PolicyGroups []map[string]egernPolicyGroup `yaml:"policy_groups"`
		Rules        []finalRule                   `yaml:"rules"`
	}{
		IPv6:         plan.DNS.IPv6,
		PolicyGroups: groups.PolicyGroups,
		Rules:        finalRules,
	}

	document.DNS.Bootstrap = []string{"system"}
	document.DNS.ProxyNameservers = []string{"system"}

	content, err := yaml.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("marshal Egern profile: %w", err)
	}

	return string(content), nil
}
