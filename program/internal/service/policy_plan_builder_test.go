package service

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/projectroot"
	"github.com/Higanoneko/ProxyRules/internal/render"
	"github.com/Higanoneko/ProxyRules/internal/repository"
)

func TestBuildPolicyPlanKeepsCoreSections(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	if len(plan.Proxy.Groups) == 0 {
		t.Fatal("expected proxy groups")
	}
	if len(plan.Rules) == 0 {
		t.Fatal("expected canonical rules")
	}
	if len(plan.DNS.Upstreams.Default) == 0 {
		t.Fatal("expected dns nameserver")
	}
	if len(plan.DNS.Upstreams.ProxyServer) == 0 || len(plan.DNS.Upstreams.Direct) == 0 || len(plan.DNS.Upstreams.Fallback) == 0 {
		t.Fatal("expected split dns resolver lists")
	}
	if len(plan.DNS.BootstrapResolvers) == 0 {
		t.Fatal("expected bootstrap dns resolvers")
	}
	if len(plan.Proxy.Countries) == 0 {
		t.Fatal("expected country fallback inventory")
	}
	if got := plan.Proxy.Groups[len(plan.Proxy.Groups)-1].Name; got != "GLOBAL" {
		t.Fatalf("expected GLOBAL to be the last proxy group, got %s", got)
	}
}

func TestPolicyGroupsReachGeneratedFormats(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	ai := policySpec(t, &base.PolicyConfig, "AI")
	ai.Icon = "https://example.test/ai-custom.png"
	ai.Proxies = []string{"手动选择", "DIRECT"}
	policySpec(t, &base.PolicyConfig, "香港节点").Filter = "(?i)MyHongKong"
	policySpec(t, &base.PolicyConfig, "Proxies").Icon = "https://example.test/proxies-custom.png"

	plan, err := NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	for _, group := range plan.Proxy.Groups {
		if group.Name == "AI" && !slices.Equal(group.Proxies, []string{"手动选择", "DIRECT"}) {
			t.Fatalf("AI proxies did not come from PolicyConfig: %v", group.Proxies)
		}
	}

	mihomo, err := render.NewMihomoRenderer(base).RenderStandard(plan, false, true)
	if err != nil {
		t.Fatalf("render mihomo: %v", err)
	}
	for _, value := range []string{ai.Icon, "(?i)MyHongKong"} {
		if !strings.Contains(mihomo, value) {
			t.Errorf("mihomo output missing configured value %q", value)
		}
	}

	script, err := render.NewMihomoScriptRenderer(base).RenderArgs(plan)
	if err != nil {
		t.Fatalf("render script: %v", err)
	}
	for _, value := range []string{ai.Icon, "(?i)MyHongKong", `"proxies":["手动选择","DIRECT"]`} {
		if !strings.Contains(script, value) {
			t.Errorf("script output missing configured value %q", value)
		}
	}
	if !strings.Contains(script, "const proxyGroups = POLICY_GROUPS") {
		t.Error("script does not build groups from PolicyConfig")
	}

	surge, err := render.NewSurgeRenderer(base).Render(plan)
	if err != nil {
		t.Fatalf("render surge: %v", err)
	}
	if !strings.Contains(surge, "icon-url="+ai.Icon) || !strings.Contains(surge, "icon-url="+policySpec(t, &base.PolicyConfig, "Proxies").Icon) || !strings.Contains(surge, "policy-regex-filter=(MyHongKong)") {
		t.Error("surge output missing configured group icons")
	}
	loon, err := render.NewLoonRenderer(base).Render(plan)
	if err != nil {
		t.Fatalf("render loon: %v", err)
	}
	if !strings.Contains(loon, `HK_Filter = NameRegex, FilterKey = "(?i)MyHongKong"`) {
		t.Error("loon output missing configured country regex")
	}
}

func TestPolicyIconsRequireEveryGeneratedGroup(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	policySpec(t, &base.PolicyConfig, "AI").Icon = ""

	_, err = NewPolicyPlanBuilder(base).Build(true, nil)
	if err == nil || !strings.Contains(err.Error(), `missing Icon for policyname "AI"`) {
		t.Fatalf("expected missing AI icon error, got %v", err)
	}
}

func TestNewPolicyGroupNeedsNoCatalogChange(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	base.PolicyConfig.Groups = append(base.PolicyConfig.Groups, domain.PolicyGroupSpec{
		Name: "Custom Group", Icon: "https://example.test/custom.png", Type: "select", Proxies: []string{"选择代理", "DIRECT"},
	})
	plan, err := NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if got := plan.Proxy.Groups[len(plan.Proxy.Groups)-1].Name; got != "Custom Group" {
		t.Fatalf("expected configured group, got %q", got)
	}
}

func TestConfiguredCountryPatternAndFallbackProxies(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	policySpec(t, &base.PolicyConfig, "香港节点").Filter = "(?i)MyHongKong"
	plan, err := NewPolicyPlanBuilder(base).Build(true, []string{"MyHongKong 01", "MyHongKong 02"})
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if !slices.Equal(plan.Proxy.CountryGroupNames, []string{"香港节点"}) {
		t.Fatalf("configured country regex was not used: %v", plan.Proxy.CountryGroupNames)
	}
	for _, group := range plan.Proxy.Groups {
		if group.Name == "US Media" && !slices.Equal(group.Proxies, []string{"选择代理", "香港节点", "手动选择", "直接连接"}) {
			t.Fatalf("missing country did not use configured fallback proxies: %v", group.Proxies)
		}
	}
}

func TestSurgeSubscriptionGroupNameComesFromPolicyConfig(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	policySpec(t, &base.PolicyConfig, "Proxies").Name = "My Subscriptions"
	plan, err := NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	surge, err := render.NewSurgeRenderer(base).Render(plan)
	if err != nil {
		t.Fatalf("render surge: %v", err)
	}
	if !strings.Contains(surge, "include-other-group=My Subscriptions") || !strings.Contains(surge, "My Subscriptions = select, policy-path=") {
		t.Fatal("surge subscription group name did not come from PolicyConfig")
	}
}

func policySpec(t *testing.T, config *domain.PolicyConfig, name string) *domain.PolicyGroupSpec {
	t.Helper()
	for i := range config.Groups {
		if config.Groups[i].Name == name {
			return &config.Groups[i]
		}
	}
	t.Fatalf("missing policy group %q", name)
	return nil
}

func TestRulePolicyNameMustMatchGeneratedGroup(t *testing.T) {
	base, err := repository.NewBaseRepository(serviceProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	base.RawRules.BaseRules.Entries["AI"]["policyname"] = "Unknown Group"

	_, err = NewPolicyPlanBuilder(base).Build(true, nil)
	if err == nil || !strings.Contains(err.Error(), `rule "AI": policyname "Unknown Group" is not a generated policy group`) {
		t.Fatalf("expected unmatched rule policyname error, got %v", err)
	}
}

func serviceProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
