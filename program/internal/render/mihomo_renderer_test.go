package render_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Higanoneko/ProxyRules/internal/projectroot"
	"github.com/Higanoneko/ProxyRules/internal/render"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"github.com/Higanoneko/ProxyRules/internal/service"
)

func TestRenderMihomo4RootConfigPreservesTemplateAndCoreSections(t *testing.T) {
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

	if !strings.Contains(content, "Box4Root") {
		t.Fatal("expected Box4Root template comments")
	}
	for _, section := range []string{"proxy-groups:", "rule-providers:", "rules:"} {
		if !strings.Contains(content, section) {
			t.Fatalf("expected %s in tun config", section)
		}
	}
	for _, marker := range []string{"proxy-server-nameserver:", "direct-nameserver:"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in tun config", marker)
		}
	}
	for _, marker := range []string{"listen: 0.0.0.0:1053", "fake-ip-range: 198.18.0.1/16", "fake-ip-range6: fdfe:dcba:9876::1/64"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in mihomo4root dns config", marker)
		}
	}
	if strings.Contains(content, "{DNS_IP_List}") {
		t.Fatal("expected placeholders resolved")
	}
	if !strings.Contains(content, "tun:\n  enable: true") {
		t.Fatal("expected tun enabled in mihomo4root config")
	}
	if !strings.Contains(content, "\nrules:\n  - DST-PORT,53,DNS_Hijack\n") {
		t.Fatal("expected DNS_Hijack rule from head template")
	}
	if strings.Index(content, "\n  - DST-PORT,53,DNS_Hijack\n") > strings.Index(content, "\n  - RULE-SET,") {
		t.Fatal("expected DNS_Hijack rule to stay before generated rules")
	}
}

func TestRenderStandardConfigPreservesHeadDNSFields(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomoRenderer(base).RenderStandard(plan, false)
	if err != nil {
		t.Fatalf("render standard: %v", err)
	}

	for _, marker := range []string{"listen: 0.0.0.0:1053", "fake-ip-range: 198.18.0.1/16", "fake-ip-range6: fdfe:dcba:9876::1/64"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in standard dns config", marker)
		}
	}
}

func TestRenderMihomo4RootConfigSupportsTunDisabled(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomo4RootRenderer(base).RenderMihomo4Root(plan, false)
	if err != nil {
		t.Fatalf("render mihomo4root disabled tun: %v", err)
	}

	if !strings.Contains(content, "tun:\n  enable: false") {
		t.Fatal("expected tun disabled in mihomo4root config")
	}
}

func TestRenderArgsScriptContainsSharedPayload(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomoScriptRenderer(base).RenderArgs(plan)
	if err != nil {
		t.Fatalf("render args script: %v", err)
	}

	for _, marker := range []string{"const POLICY_TEMPLATES", "function buildPolicyGroup", "const RULE_PROVIDERS", "const DNS_TEMPLATE ="} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in args script", marker)
		}
	}
	if !strings.Contains(content, "const proxyGroups = [...policyGroups, ...countryGroups, ...globalGroups];") {
		t.Fatal("expected GLOBAL groups to be appended after country groups")
	}
	for _, marker := range []string{"proxy-server-nameserver", "direct-nameserver"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in args script", marker)
		}
	}
	for _, marker := range []string{`"listen":"0.0.0.0:1053"`, `"fake-ip-range":"198.18.0.1/16"`, `"fake-ip-range6":"fdfe:dcba:9876::1/64"`} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in args script dns template", marker)
		}
	}
}

func mihomoProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
