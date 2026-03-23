package domain

type GroupStrategy string

const (
	StrategySelector       GroupStrategy = "selector"
	StrategyManual         GroupStrategy = "manual"
	StrategyDefault        GroupStrategy = "default"
	StrategyMediaPreferred GroupStrategy = "media_preferred"
	StrategyDirectFirst    GroupStrategy = "direct_first"
	StrategyFixed          GroupStrategy = "fixed"
	StrategyGlobal         GroupStrategy = "global"
)

type PolicyTemplate struct {
	Name                  string        `json:"name"`
	IconURL               string        `json:"icon_url"`
	Strategy              GroupStrategy `json:"strategy"`
	PreferredCountryGroup string        `json:"preferred_country_group,omitempty"`
	FixedProxies          []string      `json:"fixed_proxies,omitempty"`
}

type ProxyGroup struct {
	Name          string
	Type          string
	Icon          string
	Proxies       []string
	IncludeAll    bool
	Filter        string
	ExcludeFilter string
	URL           string
	Interval      int
	Tolerance     int
	Lazy          *bool
	ProxiesAnchor string
	ProxiesAlias  string
}

type ProxyPlan struct {
	Countries         []CountryInfo
	CountryGroupNames []string
	Groups            []ProxyGroup
}
