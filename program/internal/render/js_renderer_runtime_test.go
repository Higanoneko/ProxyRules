package render_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/Higanoneko/ProxyRules/internal/render"
	"github.com/Higanoneko/ProxyRules/internal/repository"
	"github.com/Higanoneko/ProxyRules/internal/service"
)

func TestRenderMihomoScriptDNSDisabledPreservesInput(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node.js is required to execute the generated script")
	}

	base, err := repository.NewBaseRepository(mihomoProjectRoot()).Load()
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	plan, err := service.NewPolicyPlanBuilder(base).Build(true, nil)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	renderer := render.NewMihomoScriptRenderer(base)
	for _, test := range []struct {
		name   string
		render func() (string, error)
	}{
		{"fixed dns-0", func() (string, error) { return renderer.RenderFixed(plan, false, false) }},
		{"fixed full dns-0", func() (string, error) { return renderer.RenderFixed(plan, true, false) }},
		{"args dns-0", func() (string, error) { return renderer.RenderArgs(plan) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			script, err := test.render()
			if err != nil {
				t.Fatalf("render script: %v", err)
			}
			const harness = `const fs = require("node:fs");
const vm = require("node:vm");
const script = fs.readFileSync(0, "utf8");
const config = {
    proxies: [{name: "test"}],
    dns: {"marker": "keep-dns"},
    sniffer: {"marker": "keep-sniffer"},
};
const sandbox = {$arguments: {dns: false}, config};
vm.runInNewContext(script + "\nthis.output = main(config);", sandbox);
process.stdout.write(JSON.stringify(sandbox.output));`
			command := exec.Command("node", "-e", harness)
			command.Stdin = strings.NewReader(script)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("execute script: %v: %s", err, output)
			}
			var result map[string]json.RawMessage
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("decode script output: %v: %s", err, output)
			}
			for key, want := range map[string]string{"dns": "keep-dns", "sniffer": "keep-sniffer"} {
				var value struct {
					Marker string `json:"marker"`
				}
				if err := json.Unmarshal(result[key], &value); err != nil {
					t.Fatalf("decode %s: %v", key, err)
				}
				if value.Marker != want {
					t.Fatalf("%s was replaced or removed: got %q, want %q", key, value.Marker, want)
				}
			}
		})
	}
}
