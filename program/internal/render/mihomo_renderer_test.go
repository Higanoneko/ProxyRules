package render_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/PianCat/ProxyRules/internal/projectroot"
	"github.com/PianCat/ProxyRules/internal/render"
	"github.com/PianCat/ProxyRules/internal/repository"
	"github.com/PianCat/ProxyRules/internal/service"
)

func TestRenderTunConfigPreservesTemplateAndCoreSections(t *testing.T) {
	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewMihomoRenderer(base).RenderTun(plan)
	if err != nil {
		t.Fatalf("render tun: %v", err)
	}

	if !strings.Contains(content, "# 此为精简配置") {
		t.Fatal("expected tun template comments")
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
	if strings.Contains(content, "{DNS_IP_List}") {
		t.Fatal("expected placeholders resolved")
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
}

func mihomoProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
