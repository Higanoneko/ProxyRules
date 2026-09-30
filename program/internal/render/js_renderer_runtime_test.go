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

func TestRenderMihomoScriptOverwritesOnlyEnabledSections(t *testing.T) {
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
		name      string
		render    func() (string, error)
		full      bool
		dns       bool
		arguments string
	}{
		{"fixed full-0 dns-0", func() (string, error) { return renderer.RenderFixed(plan, false, false) }, false, false, "{}"},
		{"fixed full-1 dns-0", func() (string, error) { return renderer.RenderFixed(plan, true, false) }, true, false, "{}"},
		{"fixed full-0 dns-1", func() (string, error) { return renderer.RenderFixed(plan, false, true) }, false, true, "{}"},
		{"fixed full-1 dns-1", func() (string, error) { return renderer.RenderFixed(plan, true, true) }, true, true, "{}"},
		{"args full-0 dns-0", func() (string, error) { return renderer.RenderArgs(plan) }, false, false, `{"full":false,"dns":false}`},
		{"args full-1 dns-1", func() (string, error) { return renderer.RenderArgs(plan) }, true, true, `{"full":true,"dns":true}`},
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
    port: 7890,
    hosts: {"example.test": "127.0.0.1"},
    "mixed-port": 7893,
    ipv6: false,
    profile: {tracing: true},
    dns: {"marker": "keep-dns"},
    sniffer: {"marker": "keep-sniffer"},
};
const original = JSON.stringify(config);
const sandbox = {$arguments: JSON.parse(process.argv[1]), config};
vm.runInNewContext(script + "\nthis.output = main(config);", sandbox);
if (JSON.stringify(config) !== original) throw new Error("input config was mutated");
process.stdout.write(JSON.stringify(sandbox.output));`
			command := exec.Command("node", "-e", harness, test.arguments)
			command.Stdin = strings.NewReader(script)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("execute script: %v: %s", err, output)
			}
			var result map[string]json.RawMessage
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("decode script output: %v: %s", err, output)
			}
			if string(result["port"]) != "7890" || string(result["hosts"]) != `{"example.test":"127.0.0.1"}` {
				t.Fatalf("unrelated input fields were removed: port=%s, hosts=%s", result["port"], result["hosts"])
			}
			wantMixedPort := 7893
			wantIPv6 := false
			if test.full {
				wantMixedPort = plan.Ports.Mixed
				wantIPv6 = plan.DNS.IPv6
			}
			var mixedPort int
			if err := json.Unmarshal(result["mixed-port"], &mixedPort); err != nil || mixedPort != wantMixedPort {
				t.Fatalf("mixed-port: got %d (%v), want %d", mixedPort, err, wantMixedPort)
			}
			var ipv6 bool
			if err := json.Unmarshal(result["ipv6"], &ipv6); err != nil || ipv6 != wantIPv6 {
				t.Fatalf("ipv6: got %t (%v), want %t", ipv6, err, wantIPv6)
			}
			for key, want := range map[string]string{"dns": "keep-dns", "sniffer": "keep-sniffer"} {
				var value struct {
					Marker string `json:"marker"`
				}
				if err := json.Unmarshal(result[key], &value); err != nil {
					t.Fatalf("decode %s: %v", key, err)
				}
				if test.dns && value.Marker != "" {
					t.Fatalf("%s was not overwritten: %q", key, value.Marker)
				}
				if !test.dns && value.Marker != want {
					t.Fatalf("%s was replaced or removed: got %q, want %q", key, value.Marker, want)
				}
			}
		})
	}
}
