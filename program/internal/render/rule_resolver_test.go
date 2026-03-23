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
	"gopkg.in/yaml.v3"
)

func TestRuleResolverKeepsCanonicalTargets(t *testing.T) {
	base, err := repository.NewBaseRepository(renderProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	resolver := render.NewRuleResolver(base)
	providers, err := resolver.MihomoRuleProviders(plan.Rules)
	if err != nil {
		t.Fatalf("providers: %v", err)
	}
	providersYAML, err := yaml.Marshal(providers)
	if err != nil {
		t.Fatalf("marshal providers: %v", err)
	}
	if !strings.Contains(string(providersYAML), "YouTubeMusic:") {
		t.Fatal("expected YouTubeMusic provider")
	}

	mihomoRules := resolver.MihomoRules(plan.Rules)
	if !containsString(mihomoRules, "RULE-SET,LocalNetwork,DIRECT") {
		t.Fatal("expected LocalNetwork direct rule")
	}
	if !containsString(mihomoRules, "RULE-SET,LocalNetworkIP,DIRECT") {
		t.Fatal("expected LocalNetworkIP direct rule")
	}

	surgeRules, err := resolver.SurgeRemoteRules(plan.Rules)
	if err != nil {
		t.Fatalf("surge rules: %v", err)
	}
	joined := strings.Join(surgeRules, "\n")
	if !strings.Contains(joined, "rule/Surge/YouTube/YouTube.list,YouTube") {
		t.Fatal("expected YouTube surge rule URL")
	}
	if !strings.Contains(joined, "ruleset.skk.moe/List/ip/lan.conf,DIRECT") {
		t.Fatal("expected local network ip surge rule URL")
	}
}

func renderProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
