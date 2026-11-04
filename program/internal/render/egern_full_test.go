package render_test

import (
	"testing"

	"github.com/luluape/ProxyRules/internal/render"
	"github.com/luluape/ProxyRules/internal/repository"
	"github.com/luluape/ProxyRules/internal/service"
	"gopkg.in/yaml.v3"
)

func TestEgernFullProfile(t *testing.T) {
	base, err := repository.NewBaseRepository("../../..").Load()
	if err != nil {
		t.Fatal(err)
	}

	planner := service.NewPolicyPlanBuilder(base)
	renderer := render.NewEgernRenderer()

	for _, ipv6 := range []bool{true, false} {
		plan, err := planner.Build(ipv6, nil)
		if err != nil {
			t.Fatal(err)
		}

		content, err := renderer.RenderFull(
			plan,
			[]string{"https://example.com/test-subscription"},
		)
		if err != nil {
			t.Fatal(err)
		}

		var doc struct {
			IPv6         bool                     `yaml:"ipv6"`
			PolicyGroups []map[string]interface{} `yaml:"policy_groups"`
			Rules        []map[string]interface{} `yaml:"rules"`
		}

		if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
			t.Fatal(err)
		}

		if doc.IPv6 != ipv6 {
			t.Fatalf("IPv6 mismatch: expected %v", ipv6)
		}
		if len(doc.PolicyGroups) != 29 {
			t.Fatalf("expected 29 groups, got %d", len(doc.PolicyGroups))
		}
		if len(doc.Rules) != 31 {
			t.Fatalf("expected 30 rules and default, got %d", len(doc.Rules))
		}
		if _, ok := doc.Rules[len(doc.Rules)-1]["default"]; !ok {
			t.Fatal("missing final default rule")
		}
	}

	t.Log("PASS: full IPv4/IPv6 Egern profiles")
}
