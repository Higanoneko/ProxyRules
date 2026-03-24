package render

import "github.com/PianCat/ProxyRules/internal/domain"

type genericDNSView struct {
	BootstrapResolvers []string
	Nameserver         []string
	DefaultDoH         []string
	ProxyServerDoH     []string
	DirectDoH          []string
	FallbackDoH        []string
	FakeIPFilter       []string
}

type clashDNSView struct {
	Enable                bool
	IPv6                  bool
	EnhancedMode          string
	DefaultNameserver     []string
	Nameserver            []string
	ProxyServerNameserver []string
	DirectNameserver      []string
	FakeIPFilter          []string
	HasScopedResolvers    bool
}

type clashScriptDNSView struct {
	Enable                bool     `json:"enable"`
	EnhancedMode          string   `json:"enhanced-mode"`
	Nameserver            []string `json:"nameserver"`
	ProxyServerNameserver []string `json:"proxy-server-nameserver,omitempty"`
	DirectNameserver      []string `json:"direct-nameserver,omitempty"`
	FakeIPFilter          []string `json:"fake-ip-filter"`
}

func projectGenericDNS(policy domain.DNSPolicy) genericDNSView {
	defaultUpstreams := cloneStringsOrFallback(
		policy.Upstreams.Default,
		mergeUniqueStringSlices(
			policy.Upstreams.ProxyServer,
			policy.Upstreams.Direct,
			policy.Upstreams.Fallback,
		),
	)
	directUpstreams := cloneStringsOrFallback(policy.Upstreams.Direct, defaultUpstreams)
	proxyServerUpstreams := cloneStringsOrFallback(policy.Upstreams.ProxyServer, directUpstreams)
	fallbackUpstreams := cloneStringsOrFallback(policy.Upstreams.Fallback, defaultUpstreams)

	return genericDNSView{
		BootstrapResolvers: append([]string(nil), policy.BootstrapResolvers...),
		Nameserver:         defaultUpstreams,
		DefaultDoH:         defaultUpstreams,
		ProxyServerDoH:     proxyServerUpstreams,
		DirectDoH:          directUpstreams,
		FallbackDoH:        fallbackUpstreams,
		FakeIPFilter:       append([]string(nil), policy.FakeIPFilter...),
	}
}

func projectClashDNS(policy domain.DNSPolicy) clashDNSView {
	generic := projectGenericDNS(policy)
	view := clashDNSView{
		Enable:             policy.Enable,
		IPv6:               policy.IPv6,
		EnhancedMode:       policy.EnhancedMode,
		DefaultNameserver:  generic.BootstrapResolvers,
		Nameserver:         generic.DefaultDoH,
		FakeIPFilter:       generic.FakeIPFilter,
		HasScopedResolvers: hasScopedUpstreams(policy.Upstreams),
	}

	if !view.HasScopedResolvers {
		return view
	}

	view.ProxyServerNameserver = generic.ProxyServerDoH
	view.DirectNameserver = generic.DirectDoH
	view.Nameserver = generic.DefaultDoH
	return view
}

func projectClashScriptDNS(policy domain.DNSPolicy) clashScriptDNSView {
	view := projectClashDNS(policy)
	return clashScriptDNSView{
		Enable:                view.Enable,
		EnhancedMode:          view.EnhancedMode,
		Nameserver:            append([]string(nil), view.Nameserver...),
		ProxyServerNameserver: append([]string(nil), view.ProxyServerNameserver...),
		DirectNameserver:      append([]string(nil), view.DirectNameserver...),
		FakeIPFilter:          append([]string(nil), view.FakeIPFilter...),
	}
}

func hasScopedUpstreams(policy domain.DNSUpstreamPolicy) bool {
	return len(policy.ProxyServer) > 0 || len(policy.Direct) > 0 || len(policy.Fallback) > 0
}

func cloneStringsOrFallback(values []string, fallback []string) []string {
	if len(values) > 0 {
		return append([]string(nil), values...)
	}
	return append([]string(nil), fallback...)
}

func mergeUniqueStringSlices(groups ...[]string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, group := range groups {
		for _, value := range group {
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
