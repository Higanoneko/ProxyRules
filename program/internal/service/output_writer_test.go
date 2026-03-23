package service

import (
	"strings"
	"testing"
	"time"
)

func TestPrependGeneratedHeaderUsesExtensionCommentStyle(t *testing.T) {
	generatedAt := time.Date(2026, 3, 23, 15, 4, 5, 0, time.UTC)

	yamlContent := prependGeneratedHeader("sample.yaml", "foo: bar\n", generatedAt)
	if !strings.HasPrefix(yamlContent, "# Generated at (UTC): 2026-03-23T15:04:05Z") {
		t.Fatalf("expected yaml header, got %q", yamlContent)
	}

	jsContent := prependGeneratedHeader("sample.js", "const x = 1;\n", generatedAt)
	if !strings.HasPrefix(jsContent, "// Generated at (UTC): 2026-03-23T15:04:05Z") {
		t.Fatalf("expected js header, got %q", jsContent)
	}
}

func TestStripGeneratedHeaderRemovesGeneratedJSBanner(t *testing.T) {
	content := "// Generated at (UTC): 2026-03-23T15:04:05Z\n\n/* header */\nconst x = 1;\n"

	stripped := stripGeneratedHeader("sample.js", content)

	if strings.Contains(stripped, "Generated at (UTC)") {
		t.Fatalf("expected generated header removed, got %q", stripped)
	}
	if !strings.HasPrefix(stripped, "/* header */") {
		t.Fatalf("expected original js header preserved, got %q", stripped)
	}
}

func TestPrependGeneratedHeaderKeepsSGModuleMetadataFirst(t *testing.T) {
	generatedAt := time.Date(2026, 3, 23, 15, 4, 5, 0, time.UTC)
	content := "#!name=Easytier\n#!desc=desc\n\n[Proxy]\n"

	withHeader := prependGeneratedHeader("Easytier.sgmodule", content, generatedAt)

	if !strings.HasPrefix(withHeader, "#!name=Easytier\n#!desc=desc\n") {
		t.Fatalf("expected sgmodule metadata first, got %q", withHeader)
	}
	if !strings.Contains(withHeader, "# Generated at (UTC): 2026-03-23T15:04:05Z") {
		t.Fatalf("expected sgmodule header, got %q", withHeader)
	}
}
