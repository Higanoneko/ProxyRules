package render

import (
	"strings"
	"testing"

	"github.com/luluape/ProxyRules/internal/domain"
	"gopkg.in/yaml.v3"
)

func TestEgernPolicyGroups(t *testing.T) {
	plan := domain.PolicyPlan{
		Proxy: domain.ProxyPlan{
			Groups: []domain.ProxyGroup{
				{
					Name: "选择代理",
					Type: "select",
					Icon: "https://example.com/proxy.png",
					Proxies: []string{
						"香港节点", "手动选择", "DIRECT",
					},
				},
				{
					Name:       "手动选择",
					Type:       "select",
					Icon:       "https://example.com/manual.png",
					IncludeAll: true,
				},
				{
					Name:       "香港节点",
					Type:       "url-test",
					Icon:       "https://example.com/hk.png",
					IncludeAll: true,
					Filter:     "(?i)香港|HK",
					URL:        "https://cp.cloudflare.com/generate_204",
					Interval:   60,
					Tolerance:  20,
				},
				{
					Name:          "其他节点",
					Type:          "select",
					Icon:          "https://example.com/other.png",
					IncludeAll:    true,
					ExcludeFilter: "(?i)香港|HK",
				},
			},
		},
	}

	renderer := NewEgernRenderer()

	// 没有真实订阅时必须拒绝生成。
	_, err := renderer.RenderPolicyGroups(plan, nil)
	if err == nil {
		t.Fatal("expected missing subscription error")
	}

	// 使用模拟订阅验证配置结构。
	content, err := renderer.RenderPolicyGroups(
		plan,
		[]string{"https://example.com/subscription"},
	)
	if err != nil {
		t.Fatal(err)
	}

	var result egernPolicyDocument
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		t.Fatal(err)
	}

	if len(result.PolicyGroups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(result.PolicyGroups))
	}

	var manual, hk, other egernPolicyGroup

	for _, item := range result.PolicyGroups {
		for _, group := range item {
			switch group.Name {
			case "手动选择":
				manual = group
			case "香港节点":
				hk = group
			case "其他节点":
				other = group
			}
		}
	}

	if len(manual.URLs) != 1 {
		t.Fatal("manual subscription missing")
	}

	if !hk.Flatten || hk.Filter == "" {
		t.Fatal("Hong Kong node filtering missing")
	}

	if !other.Flatten || !strings.Contains(other.Filter, "?!") {
		t.Fatal("other node exclusion missing")
	}

	if hk.Icon != "https://example.com/hk.png" {
		t.Fatal("Hong Kong icon was not preserved")
	}
}
