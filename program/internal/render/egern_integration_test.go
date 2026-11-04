package render_test

import (
	"strings"
	"testing"

	"github.com/luluape/ProxyRules/internal/render"
	"github.com/luluape/ProxyRules/internal/repository"
	"github.com/luluape/ProxyRules/internal/service"
	"gopkg.in/yaml.v3"
)

type nativeGroup struct {
	Name     string   `yaml:"name"`
	Icon     string   `yaml:"icon"`
	Policies []string `yaml:"policies"`
	URLs     []string `yaml:"urls"`
	Flatten  bool     `yaml:"flatten"`
	Filter   string   `yaml:"filter"`
}

type nativeConfig struct {
	PolicyGroups []map[string]nativeGroup `yaml:"policy_groups"`
}

func TestEgernRealPolicyConfig(t *testing.T) {
	// Locate repository root from program/internal/render.
	root := "../../.."

	base, err := repository.NewBaseRepository(root).Load()
	if err != nil {
		t.Fatal(err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatal(err)
	}

	subscription := "https://example.com/test-subscription"

	result, err := render.NewEgernRenderer().RenderPolicyGroups(
		plan, []string{subscription},
	)
	if err != nil {
		t.Fatal(err)
	}

	var config nativeConfig
	if err := yaml.Unmarshal([]byte(result), &config); err != nil {
		t.Fatal(err)
	}

	expected := make([]string, 0)
	expectedIcons := make(map[string]string)
	for _, spec := range base.PolicyConfig.Groups {
		if spec.MihomoOnly || spec.SurgeOnly {
			continue
		}
		expected = append(expected, spec.Name)
		expectedIcons[spec.Name] = spec.Icon
	}

	if len(config.PolicyGroups) != len(expected) {
		t.Fatalf(
			"policy group count: got %d, want %d",
			len(config.PolicyGroups), len(expected),
		)
	}

	if len(expected) != 29 {
		t.Fatalf("expected 29 common policy groups, got %d", len(expected))
	}

	found := make(map[string]nativeGroup)

	for i, item := range config.PolicyGroups {
		if len(item) != 1 {
			t.Fatalf("invalid group at index %d", i)
		}
		for kind, group := range item {
			if group.Name != expected[i] {
				t.Fatalf(
					"group order mismatch at %d: got %q want %q",
					i, group.Name, expected[i],
				)
			}
			if _, exists := found[group.Name]; exists {
				t.Fatalf("duplicate group: %s", group.Name)
			}
			if group.Icon != expectedIcons[group.Name] {
				t.Fatalf("icon mismatch: %s", group.Name)
			}
			if kind != "select" && kind != "auto_test" {
				t.Fatalf("unsupported group kind %q", kind)
			}
			found[group.Name] = group
		}
	}

	manual := found["手动选择"]
	if len(manual.URLs) != 1 || manual.URLs[0] != subscription {
		t.Fatal("manual subscription source not preserved")
	}

	for _, name := range []string{
		"香港节点", "台湾节点", "新加坡节点",
		"日本节点", "美国节点",
	} {
		group, ok := found[name]
		if !ok {
			t.Fatalf("missing region group: %s", name)
		}
		if !group.Flatten || group.Filter == "" {
			t.Fatalf("missing region filtering: %s", name)
		}
		if len(group.Policies) != 1 ||
			group.Policies[0] != "手动选择" {
			t.Fatalf("invalid region source: %s", name)
		}
	}

	other := found["其他节点"]
	if !other.Flatten || !strings.Contains(other.Filter, "?!") {
		t.Fatal("missing other-region exclusion")
	}

	// Preserve all service categories defined by PolicyConfig.
	for _, name := range []string{
		"AI", "Telegram", "YouTube", "Netflix",
		"Spotify", "TikTok", "Steam", "Game",
		"E-Hentai", "PornSite", "US Media",
		"Taiwan Media", "Japan Media", "Global Media",
		"Apple", "Microsoft", "Google", "Google FCM",
		"Sogou Privacy", "ADBlock", "直接连接",
	} {
		if _, ok := found[name]; !ok {
			t.Fatalf("missing service policy: %s", name)
		}
	}

	t.Logf(
		"PASS: %d original policy groups, icons and regions preserved",
		len(found),
	)
}
