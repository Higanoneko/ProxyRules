package postprocess

import (
	"strings"
	"testing"
)

func TestMergeEasytierWrapsOriginalMain(t *testing.T) {
	convertScript := "/* header */\nfunction main(config) {\n    return config;\n}\n"
	easytierScript := configStart + "\nconst EASYTIER_CONFIG = {};\n" + configEnd + "\n\nfunction main(config) {\n    return config;\n}\n"

	merged := MergeEasytier(convertScript, easytierScript)

	for _, marker := range []string{"EASYTIER_CONFIG", "function _originalMain", "function _easytierEnhance", "return _easytierEnhance(_originalMain(config));"} {
		if !contains(merged, marker) {
			t.Fatalf("expected %s in merged script", marker)
		}
	}
}

func TestRenameMainFunctionRenamesOnlyFirstMatch(t *testing.T) {
	content := "function main(config) {\n    return config;\n}\nfunction main(other) {\n    return other;\n}\n"

	renamed := renameMainFunction(content, "_originalMain")

	if !contains(renamed, "function _originalMain(config)") {
		t.Fatalf("expected first main function to be renamed")
	}
	if !contains(renamed, "function main(other)") {
		t.Fatalf("expected second main function to stay unchanged")
	}
}

func TestMergeEasytierSupportsCRLFSource(t *testing.T) {
	convertScript := "/* header */\nfunction main(config) {\n    return config;\n}\n"
	easytierScript := strings.ReplaceAll(configStart+"\nconst EASYTIER_CONFIG = {};\n"+configEnd+"\n\nfunction main(config) {\n    return config;\n}\n", "\n", "\r\n")

	merged := MergeEasytier(convertScript, easytierScript)

	if !contains(merged, "const EASYTIER_CONFIG = {};") {
		t.Fatalf("expected config block to be preserved")
	}
	if !contains(merged, "function _easytierEnhance(config)") {
		t.Fatalf("expected easytier logic to be renamed from CRLF source")
	}
}

func contains(content string, expected string) bool {
	return len(content) >= len(expected) && (content == expected || containsAt(content, expected))
}

func containsAt(content string, expected string) bool {
	for index := 0; index+len(expected) <= len(content); index++ {
		if content[index:index+len(expected)] == expected {
			return true
		}
	}
	return false
}
