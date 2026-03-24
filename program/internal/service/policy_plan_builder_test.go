package service

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/PianCat/ProxyRules/internal/projectroot"
	"github.com/PianCat/ProxyRules/internal/repository"
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
	if len(plan.DNS.Upstreams.ProxyBootstrap) == 0 || len(plan.DNS.Upstreams.Direct) == 0 || len(plan.DNS.Upstreams.Proxy) == 0 {
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

func serviceProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
