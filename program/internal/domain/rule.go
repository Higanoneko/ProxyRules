package domain

type TestURLs struct {
	Internet string
	Proxy    string
}

type Ports struct {
	Mixed  int
	HTTP   int
	Socks5 int
}

type RemoteRuleSpec struct {
	RuleID           string
	PolicyName       string
	TagName          string
	MihomoPolicyName string
	SurgeOption      string
}

type RuleSourceRef struct {
	Category   string
	Behavior   string
	RemoteFile string
}

type RuleBinding struct {
	RuleID           string
	ProviderName     string
	PolicyName       string
	TagName          string
	MihomoPolicyName string
	SurgeOption      string
	Source           RuleSourceRef
}

type PolicyPlan struct {
	DNS      DNSPolicy
	Sniffer  map[string]any
	Proxy    ProxyPlan
	Rules    []RuleBinding
	TestURLs TestURLs
	Ports    Ports
}
