package render

import (
	"testing"

	"github.com/PianCat/ProxyRules/internal/domain"
)

func TestProjectGenericDNSUsesDefaultResolvers(t *testing.T) {
	policy := domain.DNSPolicy{
		BootstrapResolvers: []string{"119.29.29.29"},
		Upstreams: domain.DNSUpstreamPolicy{
			Default: []string{"https://doh.pub/dns-query"},
			Direct:  []string{"https://direct.example/dns-query"},
			Proxy:   []string{"https://proxy.example/dns-query"},
		},
		FakeIPFilter: []string{"+.lan"},
	}

	view := projectGenericDNS(policy)
	if len(view.Nameserver) != 1 || view.Nameserver[0] != "https://doh.pub/dns-query" {
		t.Fatalf("expected generic view to prefer default resolvers, got %#v", view.Nameserver)
	}
}

func TestProjectClashDNSAddsScopedResolvers(t *testing.T) {
	policy := domain.DNSPolicy{
		Enable:             true,
		IPv6:               true,
		EnhancedMode:       "fake-ip",
		BootstrapResolvers: []string{"119.29.29.29"},
		Upstreams: domain.DNSUpstreamPolicy{
			Default:        []string{"https://default.example/dns-query"},
			ProxyBootstrap: []string{"https://bootstrap.example/dns-query"},
			Direct:         []string{"https://direct.example/dns-query"},
			Proxy:          []string{"https://proxy.example/dns-query"},
		},
		FakeIPFilter: []string{"+.lan"},
	}

	view := projectClashDNS(policy)
	if !view.HasScopedResolvers {
		t.Fatal("expected clash view to enable scoped resolvers")
	}
	if got := view.ProxyServerNameserver[0]; got != "https://bootstrap.example/dns-query" {
		t.Fatalf("unexpected proxy bootstrap resolver: %s", got)
	}
	if got := view.DirectNameserver[0]; got != "https://direct.example/dns-query" {
		t.Fatalf("unexpected direct resolver: %s", got)
	}
	if got := view.Nameserver[0]; got != "https://proxy.example/dns-query" {
		t.Fatalf("unexpected proxy resolver: %s", got)
	}
}
