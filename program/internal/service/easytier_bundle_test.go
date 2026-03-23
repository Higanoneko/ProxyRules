package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const easytierFixture = `// ============================================
// Wireguard_Easytier 用户配置区域
const EASYTIER_CONFIG = {
    proxy: {},
    rules: ["IP-CIDR,10.0.0.0/8,Easytier,no-resolve"]
};
// Wireguard_Easytier 用户配置区域结束
// ============================================

function main(config) {
    return config;
}
`

func TestGenerateEasytierBundleWritesMergedScripts(t *testing.T) {
	root := t.TempDir()
	mihomoOutputDir := filepath.Join(root, "Config", "Mihomo")
	outputDir := filepath.Join(root, "Wireguard_Easytier", "Mihomo")
	easytierSourcePath := filepath.Join(root, "Base", "Wireguard", "Easytier", "Easytier.js")

	if err := os.MkdirAll(mihomoOutputDir, 0o755); err != nil {
		t.Fatalf("mkdir mihomo output: %v", err)
	}
	if err := writeFile(easytierSourcePath, easytierFixture); err != nil {
		t.Fatalf("write easytier source: %v", err)
	}
	for _, name := range []string{"mihomo_convert_ipv6-1_full-0.js", "mihomo_convert_args.js"} {
		if err := writeFile(filepath.Join(mihomoOutputDir, name), "/* header */\nfunction main(config) {\n    return config;\n}\n"); err != nil {
			t.Fatalf("write source script %s: %v", name, err)
		}
	}

	generatedAt := time.Date(2026, 3, 23, 15, 4, 5, 0, time.UTC)
	if err := generateEasytierBundle(mihomoOutputDir, easytierSourcePath, outputDir, generatedAt); err != nil {
		t.Fatalf("generate easytier bundle: %v", err)
	}

	for _, name := range []string{"mihomo_convert_ipv6-1_full-0.js", "mihomo_convert_args.js"} {
		content, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			t.Fatalf("read generated file %s: %v", name, err)
		}
		if !strings.HasPrefix(string(content), "// Generated at (UTC): 2026-03-23T15:04:05Z") {
			t.Fatalf("expected generated header in %s", name)
		}
		for _, marker := range []string{"EASYTIER_CONFIG", "function _originalMain", "function _easytierEnhance"} {
			if !strings.Contains(string(content), marker) {
				t.Fatalf("expected marker %s in %s", marker, name)
			}
		}
	}
}

func TestResolveEasytierOutputDir(t *testing.T) {
	if got := resolveEasytierOutputDir(filepath.Join("D:\\", "Projects", "ProxyRules", "Config")); got != filepath.Join("D:\\", "Projects", "ProxyRules", "Wireguard_Easytier", "Mihomo") {
		t.Fatalf("unexpected sibling output dir: %s", got)
	}

	if got := resolveEasytierOutputDir(filepath.Join("D:\\", "Temp", "ProxyRules")); got != filepath.Join("D:\\", "Temp", "ProxyRules", "Wireguard_Easytier", "Mihomo") {
		t.Fatalf("unexpected nested output dir: %s", got)
	}
}

func TestCopyEasytierSurgeModule(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "Base", "Wireguard", "Easytier", "Easytier.sgmodule")

	if err := writeFile(sourcePath, "surge module"); err != nil {
		t.Fatalf("write source module: %v", err)
	}
	generatedAt := time.Date(2026, 3, 23, 15, 4, 5, 0, time.UTC)
	if err := copyEasytierSurgeModule(sourcePath, filepath.Join(root, "Config"), generatedAt); err != nil {
		t.Fatalf("copy surge module: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "Wireguard_Easytier", "Surge", "Easytier.sgmodule"))
	if err != nil {
		t.Fatalf("read copied module: %v", err)
	}
	if !strings.HasPrefix(string(content), "# Generated at (UTC): 2026-03-23T15:04:05Z") {
		t.Fatalf("expected generated header, got %s", string(content))
	}
	if !strings.Contains(string(content), "surge module") {
		t.Fatalf("unexpected copied content: %s", string(content))
	}
}
