package render_test

import (
	"strings"
	"testing"

	"github.com/luluape/ProxyRules/internal/render"
	"github.com/luluape/ProxyRules/internal/repository"
	"github.com/luluape/ProxyRules/internal/service"
	"gopkg.in/yaml.v3"
)

type actualEgernRule struct {
	Match          string `yaml:"match"`
	Policy         string `yaml:"policy"`
	UpdateInterval int    `yaml:"update_interval"`
}

type actualEgernRuleDocument struct {
	Rules []struct {
		RuleSet actualEgernRule `yaml:"rule_set"`
	} `yaml:"rules"`
}

func TestEgernAllThirtyRuleSets(t *testing.T) {
	base, err := repository.NewBaseRepository("../../..").Load()
	if err != nil {
		t.Fatal(err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := render.NewEgernRenderer().RenderRuleSets(plan)
	if err != nil {
		t.Fatal(err)
	}

	var doc actualEgernRuleDocument
	if err := yaml.Unmarshal([]byte(result), &doc); err != nil {
		t.Fatal(err)
	}

	if len(plan.Rules) != 30 {
		t.Fatalf("source bindings: got %d, want 30", len(plan.Rules))
	}
	if len(doc.Rules) != len(plan.Rules) {
		t.Fatalf(
			"generated references: got %d, want %d",
			len(doc.Rules), len(plan.Rules),
		)
	}

	seen := make(map[string]bool)
	for i, binding := range plan.Rules {
		entry := doc.Rules[i].RuleSet

		if entry.Policy != binding.PolicyName {
			t.Fatalf(
				"rule %d %s: policy %q != %q",
				i, binding.RuleID, entry.Policy, binding.PolicyName,
			)
		}
		if !strings.HasPrefix(
			entry.Match, render.EgernRulesBaseURL+"/",
		) {
			t.Fatalf("invalid URL: %s", entry.Match)
		}
		if entry.UpdateInterval != 86400 {
			t.Fatalf("invalid update interval: %s", binding.RuleID)
		}
		if seen[entry.Match] {
			t.Fatalf("duplicate rule URL: %s", entry.Match)
		}
		seen[entry.Match] = true
	}

	t.Logf("PASS: all %d rule bindings preserved in source order", len(doc.Rules))
}
