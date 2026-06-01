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

func TestRenderStashConfigPreservesHeadDNSFields(t *testing.T) {
	base, err := repository.NewBaseRepository(stashProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}

	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	content, err := render.NewStashRenderer(base).RenderFull(plan)
	if err != nil {
		t.Fatalf("render stash: %v", err)
	}

	for _, marker := range []string{"listen: 0.0.0.0:1053", "fake-ip-range: 198.18.0.1/16", "fake-ip-range6: fdfe:dcba:9876::1/64"} {
		if !strings.Contains(content, marker) {
			t.Fatalf("expected %s in stash dns config", marker)
		}
	}
}

func stashProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
