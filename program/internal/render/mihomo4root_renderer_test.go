package render_test

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/Higanoneko/ProxyRules/internal/render"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"github.com/Higanoneko/ProxyRules/internal/service"
)

type mihomo4RootTestGroup struct {
	Name          string `yaml:"name"`
	IncludeAll    bool   `yaml:"include-all"`
	ExcludeFilter string `yaml:"exclude-filter"`
}

func TestRenderMihomo4RootExcludesDNSHijackFromManualAndOtherGroups(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomo4RootRenderer(base).RenderMihomo4Root(plan, true)
	if err != nil {
		t.Fatalf("render mihomo4root: %v", err)
	}

	var document struct {
		ProxyGroups []mihomo4RootTestGroup `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("unmarshal rendered config: %v", err)
	}

	manual := findMihomo4RootGroup(t, document.ProxyGroups, "手动选择")
	if !manual.IncludeAll {
		t.Fatal("expected 手动选择 to keep include-all")
	}
	if !strings.Contains(manual.ExcludeFilter, "DNS_Hijack") {
		t.Fatalf("expected 手动选择 to exclude DNS_Hijack, got %q", manual.ExcludeFilter)
	}

	other := findMihomo4RootGroup(t, document.ProxyGroups, "其他节点")
	if !other.IncludeAll {
		t.Fatal("expected 其他节点 to keep include-all")
	}
	if !strings.Contains(other.ExcludeFilter, "香港") || !strings.Contains(other.ExcludeFilter, "DNS_Hijack") {
		t.Fatalf("expected 其他节点 to keep country excludes and add DNS_Hijack, got %q", other.ExcludeFilter)
	}
	if !strings.HasPrefix(other.ExcludeFilter, "(?i)") {
		t.Fatalf("expected 其他节点 exclude-filter to be case-insensitive, got %q", other.ExcludeFilter)
	}
}

func TestRenderMihomo4RootKeepsProxyProviderCommentBlockTight(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomo4RootRenderer(base).RenderMihomo4Root(plan, true)
	if err != nil {
		t.Fatalf("render mihomo4root: %v", err)
	}

	lines := strings.Split(content, "\n")
	start := -1
	for index, line := range lines {
		if strings.Contains(line, "# 多订阅按照下面照葫芦画瓢即可") {
			start = index
			break
		}
	}
	if start < 0 {
		t.Fatal("expected proxy provider comment block marker")
	}

	for index := start + 1; index < len(lines); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "" {
			if index+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[index+1]), "#") {
				t.Fatalf("unexpected blank line inside proxy provider comment block at line %d", index+1)
			}
			continue
		}
		break
	}
}

func findMihomo4RootGroup(t *testing.T, groups []mihomo4RootTestGroup, name string) mihomo4RootTestGroup {
	t.Helper()
	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}
	t.Fatalf("missing proxy group %s", name)
	return mihomo4RootTestGroup{}
}
