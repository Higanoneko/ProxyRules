package repository

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReadPolicyConfigKeepsGroupOrderAndSharedReferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "PolicyConfig.yaml")
	content := `NodeExcludePattern: '(?i)home'
policyname:
  First:
    Icon: https://example.test/first.png
    Type: select
    Proxies: &common [DIRECT, "$CountryGroups"]
  Second:
    Icon: https://example.test/second.png
    Type: select
    Proxies: *common
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := readPolicyConfig(path)
	if err != nil {
		t.Fatalf("read policy config: %v", err)
	}
	if len(config.Groups) != 2 || config.Groups[0].Name != "First" || config.Groups[1].Name != "Second" {
		t.Fatalf("group order was not preserved: %+v", config.Groups)
	}
	if !slices.Equal(config.Groups[0].Proxies, config.Groups[1].Proxies) {
		t.Fatalf("YAML alias did not share proxy references: %+v", config.Groups)
	}
}

func TestReadPolicyConfigRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "PolicyConfig.yaml")
	content := `NodeExcludePattern: 'home'
policyname:
  First:
    Icon: https://example.test/first.png
    Type: select
    Proxise: [DIRECT]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := readPolicyConfig(path)
	if err == nil || !strings.Contains(err.Error(), `unknown field "Proxise"`) {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
