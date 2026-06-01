package render

import (
	"testing"

	"github.com/Higanoneko/ProxyRules/internal/domain"
)

func TestProjectGenericDNSUsesDefaultResolvers(t *testing.T) {
	policy := domain.DNSPolicy{
		BootstrapResolvers: []string{"119.29.29.29"},
		Upstreams: domain.DNSUpstreamPolicy{
			Default:     []string{"https://doh.pub/dns-query"},
			ProxyServer: []string{"https://proxy.example/dns-query"},
			Direct:      []string{"https://direct.example/dns-query"},
			Fallback:    []string{"https://fallback.example/dns-query"},
		},
		FakeIPFilter: []string{"+.lan"},
	}

	view := projectGenericDNS(policy)
	if len(view.Nameserver) != 1 || view.Nameserver[0] != "https://doh.pub/dns-query" {
		t.Fatalf("expected generic view to prefer default resolvers, got %#v", view.Nameserver)
	}
	if len(view.ProxyServerDoH) != 1 || view.ProxyServerDoH[0] != "https://proxy.example/dns-query" {
		t.Fatalf("expected generic view to expose proxy-server resolvers, got %#v", view.ProxyServerDoH)
	}
}

func TestProjectClashDNSAddsScopedResolvers(t *testing.T) {
	policy := domain.DNSPolicy{
		Enable:             true,
		IPv6:               true,
		EnhancedMode:       "fake-ip",
		BootstrapResolvers: []string{"119.29.29.29"},
		Upstreams: domain.DNSUpstreamPolicy{
			Default:     []string{"https://default.example/dns-query"},
			ProxyServer: []string{"https://proxy.example/dns-query"},
			Direct:      []string{"https://direct.example/dns-query"},
			Fallback:    []string{"https://fallback.example/dns-query"},
		},
		FakeIPFilter: []string{"+.lan"},
	}

	view := projectClashDNS(policy)
	if !view.HasScopedResolvers {
		t.Fatal("expected clash view to enable scoped resolvers")
	}
	if got := view.ProxyServerNameserver[0]; got != "https://proxy.example/dns-query" {
		t.Fatalf("unexpected proxy-server resolver: %s", got)
	}
	if got := view.DirectNameserver[0]; got != "https://direct.example/dns-query" {
		t.Fatalf("unexpected direct resolver: %s", got)
	}
	if got := view.Nameserver[0]; got != "https://default.example/dns-query" {
		t.Fatalf("unexpected default resolver: %s", got)
	}
}
