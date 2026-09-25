package domain

type PolicyGroupSpec struct {
	Name             string   `json:"name"`
	Icon             string   `json:"icon" yaml:"Icon"`
	Type             string   `json:"type" yaml:"Type"`
	Proxies          []string `json:"proxies,omitempty" yaml:"Proxies"`
	FallbackProxies  []string `json:"fallback_proxies,omitempty" yaml:"FallbackProxies"`
	IncludeAll       bool     `json:"include_all,omitempty" yaml:"IncludeAll"`
	Filter           string   `json:"filter,omitempty" yaml:"Filter"`
	ExcludeFilter    string   `json:"exclude_filter,omitempty" yaml:"ExcludeFilter"`
	URL              string   `json:"url,omitempty" yaml:"URL"`
	Interval         int      `json:"interval,omitempty" yaml:"Interval"`
	Tolerance        int      `json:"tolerance,omitempty" yaml:"Tolerance"`
	Lazy             *bool    `json:"lazy,omitempty" yaml:"Lazy"`
	Country          string   `json:"country,omitempty" yaml:"Country"`
	LoonFilter       string   `json:"loon_filter,omitempty" yaml:"LoonFilter"`
	MihomoOnly       bool     `json:"mihomo_only,omitempty" yaml:"MihomoOnly"`
	ExcludeDNSHijack bool     `json:"exclude_dns_hijack,omitempty" yaml:"ExcludeDNSHijack"`
	SurgeOnly        bool     `json:"surge_only,omitempty" yaml:"SurgeOnly"`
	PolicyPath       string   `json:"policy_path,omitempty" yaml:"PolicyPath"`
}

type PolicyConfig struct {
	NodeExcludePattern string
	Groups             []PolicyGroupSpec
}

type ProxyGroup struct {
	Name             string
	Type             string
	Icon             string
	Proxies          []string
	IncludeAll       bool
	Filter           string
	ExcludeFilter    string
	URL              string
	Interval         int
	Tolerance        int
	Lazy             *bool
	LoonFilter       string
	MihomoOnly       bool
	ExcludeDNSHijack bool
	ProxiesAnchor    string
	ProxiesAlias     string
}

type ProxyPlan struct {
	Countries         []CountryInfo
	CountryGroupNames []string
	Groups            []ProxyGroup
}
