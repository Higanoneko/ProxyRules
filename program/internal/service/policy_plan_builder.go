package service

import (
	"github.com/PianCat/ProxyRules/internal/catalog"
	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
)

type PolicyPlanBuilder struct {
	base           repository.BaseData
	classifier     *NodeClassifier
	ruleBuilder    *RuleCatalogBuilder
	policyTemplate []domain.PolicyTemplate
}

func NewPolicyPlanBuilder(base repository.BaseData) *PolicyPlanBuilder {
	return &PolicyPlanBuilder{
		base:           base,
		classifier:     NewNodeClassifier(),
		ruleBuilder:    NewRuleCatalogBuilder(),
		policyTemplate: catalog.PolicyTemplates(),
	}
}

func (b *PolicyPlanBuilder) Build(ipv6 bool, nodeNames []string) (domain.PolicyPlan, error) {
	rules, err := b.ruleBuilder.Build(b.base.RawRules)
	if err != nil {
		return domain.PolicyPlan{}, err
	}

	return domain.PolicyPlan{
		DNS:      b.buildDNS(ipv6),
		Sniffer:  buildSniffer(),
		Proxy:    b.buildProxyPlan(nodeNames),
		Rules:    rules,
		TestURLs: b.base.TestURLs,
		Ports:    b.base.Ports,
	}, nil
}

func (b *PolicyPlanBuilder) buildDNS(ipv6 bool) domain.DNSPolicy {
	bootstrapResolvers := make([]string, 0, len(b.base.DNS.BootstrapResolvers))
	for _, dnsIP := range b.base.DNS.BootstrapResolvers {
		if !ipv6 && containsColon(dnsIP) {
			continue
		}
		bootstrapResolvers = append(bootstrapResolvers, dnsIP)
	}

	return domain.DNSPolicy{
		Enable:             true,
		IPv6:               ipv6,
		EnhancedMode:       "fake-ip",
		BootstrapResolvers: bootstrapResolvers,
		Upstreams:          cloneDNSUpstreamPolicy(b.base.DNS.Upstreams),
		FakeIPFilter:       append([]string(nil), b.base.FakeIPFilter...),
	}
}

func buildSniffer() map[string]any {
	return map[string]any{
		"sniff": map[string]any{
			"HTTP": map[string]any{
				"ports":                []any{80, "8080-8880"},
				"override-destination": true,
			},
			"TLS": map[string]any{
				"ports": []int{443, 8443},
			},
			"QUIC": map[string]any{
				"ports": []int{443, 8443},
			},
		},
		"skip-domain": []string{
			"Mijia Cloud",
			"dlg.io.mi.com",
			"+.push.apple.com",
		},
	}
}

func (b *PolicyPlanBuilder) buildProxyPlan(nodeNames []string) domain.ProxyPlan {
	countries := b.classifier.ParseCountryInfos(nodeNames, 2)
	if len(nodeNames) == 0 || len(countries) == 0 {
		countries = b.classifier.DefaultCountryInfos()
	}

	countryGroupNames := make([]string, 0, len(countries))
	for _, country := range countries {
		countryGroupNames = append(countryGroupNames, country.Name+"节点")
	}

	groups := make([]domain.ProxyGroup, 0, len(b.policyTemplate)+len(countries))
	globalGroups := make([]domain.ProxyGroup, 0, 1)
	defaultProxies := append([]string{"选择代理"}, countryGroupNames...)
	defaultProxies = append(defaultProxies, "手动选择", "直接连接")
	directFirst := append([]string{"直接连接", "选择代理"}, countryGroupNames...)
	directFirst = append(directFirst, "手动选择")
	selector := append([]string{}, countryGroupNames...)
	selector = append(selector, "手动选择", "DIRECT")

	defaultAnchorAssigned := false
	for _, template := range b.policyTemplate {
		group := domain.ProxyGroup{
			Name: template.Name,
			Icon: template.IconURL,
			Type: "select",
		}

		switch template.Strategy {
		case domain.StrategySelector:
			group.Proxies = selector
		case domain.StrategyManual:
			group.IncludeAll = true
		case domain.StrategyDefault:
			group.Proxies = defaultProxies
			defaultAnchorAssigned = assignDefaultReference(&group, defaultAnchorAssigned)
		case domain.StrategyMediaPreferred:
			if contains(countryGroupNames, template.PreferredCountryGroup) {
				group.Proxies = []string{template.PreferredCountryGroup, "选择代理", "手动选择", "直接连接"}
			} else {
				group.Proxies = defaultProxies
				defaultAnchorAssigned = assignDefaultReference(&group, defaultAnchorAssigned)
			}
		case domain.StrategyDirectFirst:
			group.Proxies = directFirst
		case domain.StrategyFixed:
			group.Proxies = append([]string(nil), template.FixedProxies...)
		case domain.StrategyGlobal:
			group.IncludeAll = true
			group.Proxies = defaultProxies
			defaultAnchorAssigned = assignDefaultReference(&group, defaultAnchorAssigned)
			globalGroups = append(globalGroups, group)
			continue
		}

		groups = append(groups, group)
	}

	for _, country := range countries {
		if country.Name == "其他" {
			groups = append(groups, domain.ProxyGroup{
				Name:          "其他节点",
				Icon:          "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png",
				Type:          "select",
				IncludeAll:    true,
				ExcludeFilter: b.classifier.CountryExcludePattern(),
			})
			continue
		}

		lazy := false
		groups = append(groups, domain.ProxyGroup{
			Name:       country.Name + "节点",
			Icon:       country.IconURL,
			Type:       "url-test",
			IncludeAll: true,
			Filter:     country.Pattern,
			URL:        "https://cp.cloudflare.com/generate_204",
			Interval:   60,
			Tolerance:  20,
			Lazy:       &lazy,
		})
	}

	groups = append(groups, globalGroups...)

	return domain.ProxyPlan{
		Countries:         countries,
		CountryGroupNames: countryGroupNames,
		Groups:            groups,
	}
}

func assignDefaultReference(group *domain.ProxyGroup, anchorAssigned bool) bool {
	if !anchorAssigned {
		group.ProxiesAnchor = "a1"
		return true
	}
	group.ProxiesAlias = "a1"
	return true
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsColon(value string) bool {
	for _, runeValue := range value {
		if runeValue == ':' {
			return true
		}
	}
	return false
}

func cloneDNSUpstreamPolicy(policy domain.DNSUpstreamPolicy) domain.DNSUpstreamPolicy {
	return domain.DNSUpstreamPolicy{
		Default:        append([]string(nil), policy.Default...),
		ProxyBootstrap: append([]string(nil), policy.ProxyBootstrap...),
		Direct:         append([]string(nil), policy.Direct...),
		Proxy:          append([]string(nil), policy.Proxy...),
	}
}
