package domain

type DNSUpstreamPolicy struct {
	Default        []string
	ProxyServer    []string
	Direct         []string
	Fallback       []string
}

type DNSSource struct {
	BootstrapResolvers []string
	Upstreams          DNSUpstreamPolicy
}

type DNSPolicy struct {
	Enable             bool
	IPv6               bool
	EnhancedMode       string
	BootstrapResolvers []string
	Upstreams          DNSUpstreamPolicy
	FakeIPFilter       []string
}
