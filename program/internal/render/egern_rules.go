package render

import (
	"fmt"
	"path"
	"strings"

	"github.com/luluape/ProxyRules/internal/domain"
	"gopkg.in/yaml.v3"
)

const EgernRulesBaseURL = "https://raw.githubusercontent.com/luluape/ProxyRules/main/Rules/Egern"

type egernRuleSetBody struct {
	Match          string `yaml:"match"`
	Policy         string `yaml:"policy"`
	UpdateInterval int    `yaml:"update_interval"`
}

type egernRuleSetEntry struct {
	RuleSet *egernRuleSetBody `yaml:"rule_set,omitempty"`
}

type egernRuleSetDocument struct {
	Rules []egernRuleSetEntry `yaml:"rules"`
}

func egernRuleFilename(binding domain.RuleBinding) (string, error) {
	switch binding.Source.Category {
	case "Loyalsoldier":
		name := strings.TrimSuffix(
			path.Base(binding.Source.RemoteFile),
			path.Ext(binding.Source.RemoteFile),
		)
		if name == "" || name == "." {
			return "", fmt.Errorf("empty Loyalsoldier filename: %s", binding.RuleID)
		}
		return name + ".yaml", nil

	case "blackmatrix", "Higanoneko":
		return "external_" + binding.RuleID + ".yaml", nil

	case "Sukka", "dler", "ADBlock":
		return "external_" + binding.Source.Category +
			"_" + binding.RuleID + ".yaml", nil

	default:
		return "", fmt.Errorf(
			"unsupported Egern rule source %q for %s",
			binding.Source.Category, binding.RuleID,
		)
	}
}

// RenderRuleSets generates native Egern remote rule references.
// Ordering comes exclusively from PolicyPlan.Rules.
func (r *EgernRenderer) RenderRuleSets(
	plan domain.PolicyPlan,
) (string, error) {
	if len(plan.Rules) == 0 {
		return "", fmt.Errorf("empty Egern rule plan")
	}

	doc := egernRuleSetDocument{
		Rules: make([]egernRuleSetEntry, 0, len(plan.Rules)),
	}
	seen := make(map[string]bool)

	for _, binding := range plan.Rules {
		filename, err := egernRuleFilename(binding)
		if err != nil {
			return "", err
		}

		if seen[filename] {
			return "", fmt.Errorf(
				"duplicate Egern rule filename %q",
				filename,
			)
		}
		seen[filename] = true

		if strings.TrimSpace(binding.PolicyName) == "" {
			return "", fmt.Errorf(
				"empty Egern policy for %s",
				binding.RuleID,
			)
		}

		doc.Rules = append(doc.Rules, egernRuleSetEntry{
			RuleSet: &egernRuleSetBody{
				Match:          EgernRulesBaseURL + "/" + filename,
				Policy:         binding.PolicyName,
				UpdateInterval: 86400,
			},
		})
	}

	result, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshal Egern rules: %w", err)
	}
	return string(result), nil
}
