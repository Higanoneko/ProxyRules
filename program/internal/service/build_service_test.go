package service

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/projectroot"
)

func TestBuildServiceGeneratesSectionTargets(t *testing.T) {
	root := buildServiceProjectRoot()
	buildService, err := NewBuildService(root)
	if err != nil {
		t.Fatalf("new build service: %v", err)
	}

	outputRoot := t.TempDir()
	if err := buildService.Generate(
		[]domain.Target{domain.TargetStash, domain.TargetLoon, domain.TargetSurge},
		outputRoot,
		CreateTestNodes(),
	); err != nil {
		t.Fatalf("generate: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(outputRoot, "Stash", "Stash_config_full.yaml"),
		filepath.Join(outputRoot, "Stash", "Stash_override.stoverride"),
		filepath.Join(outputRoot, "Loon", "Loon_config.lcf"),
		filepath.Join(outputRoot, "Surge", "Surge_config.conf"),
	}

	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); err != nil {
			t.Fatalf("expected generated file %s: %v", expectedFile, err)
		}
	}
}

func TestNormalizeTargetsIncludesEasytierForAll(t *testing.T) {
	selectedTargets := normalizeTargets([]domain.Target{domain.TargetAll})

	if !selectedTargets[domain.TargetEasytier] {
		t.Fatalf("expected all target to include easytier")
	}
}

func buildServiceProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	root, err := projectroot.Find(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	return root
}
