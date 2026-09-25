package service

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
)

type PolicyPlanBuilder struct {
	base        repository.BaseData
	ruleBuilder *RuleCatalogBuilder
}

func NewPolicyPlanBuilder(base repository.BaseData) *PolicyPlanBuilder {
	base.PolicyConfig = clonePolicyConfig(base.PolicyConfig)
	return &PolicyPlanBuilder{
		base:        base,
		ruleBuilder: NewRuleCatalogBuilder(),
	}
}

func (b *PolicyPlanBuilder) Build(ipv6 bool, nodeNames []string) (domain.PolicyPlan, error) {
	rules, err := b.ruleBuilder.Build(b.base.RawRules)
	if err != nil {
		return domain.PolicyPlan{}, err
	}
	if err := validatePolicyConfig(b.base.PolicyConfig, rules); err != nil {
		return domain.PolicyPlan{}, err
	}
	classifier, err := NewNodeClassifier(b.base.PolicyConfig)
	if err != nil {
		return domain.PolicyPlan{}, err
	}

	return domain.PolicyPlan{
		DNS:      b.buildDNS(ipv6),
		Sniffer:  buildSniffer(),
		Proxy:    b.buildProxyPlan(classifier, nodeNames),
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

func (b *PolicyPlanBuilder) buildProxyPlan(classifier *NodeClassifier, nodeNames []string) domain.ProxyPlan {
	countries := classifier.ParseCountryInfos(nodeNames, 2)
	if len(nodeNames) == 0 || len(countries) == 0 {
		countries = classifier.DefaultCountryInfos()
	}
	availableCountries := make(map[string]bool, len(countries))
	for _, country := range countries {
		availableCountries[country.Name] = true
	}
	countryGroupNames := make([]string, 0, len(countries))
	allCountryGroups := make(map[string]bool)
	for _, spec := range b.base.PolicyConfig.Groups {
		if spec.Country == "" {
			continue
		}
		allCountryGroups[spec.Name] = true
		if availableCountries[spec.Country] {
			countryGroupNames = append(countryGroupNames, spec.Name)
		}
	}
	availableGroups := make(map[string]bool, len(countryGroupNames))
	for _, name := range countryGroupNames {
		availableGroups[name] = true
	}
	groups := make([]domain.ProxyGroup, 0, len(b.base.PolicyConfig.Groups))
	for _, spec := range b.base.PolicyConfig.Groups {
		if spec.SurgeOnly || (spec.Country != "" && !availableCountries[spec.Country]) {
			continue
		}
		proxies := spec.Proxies
		if len(spec.FallbackProxies) > 0 && hasMissingCountryReference(proxies, allCountryGroups, availableGroups) {
			proxies = spec.FallbackProxies
		}
		group := domain.ProxyGroup{
			Name:             spec.Name,
			Type:             spec.Type,
			Icon:             spec.Icon,
			Proxies:          expandProxyReferences(proxies, countryGroupNames, allCountryGroups, availableGroups),
			IncludeAll:       spec.IncludeAll,
			Filter:           spec.Filter,
			ExcludeFilter:    expandGroupFilter(spec.ExcludeFilter, classifier),
			URL:              spec.URL,
			Interval:         spec.Interval,
			Tolerance:        spec.Tolerance,
			Lazy:             cloneBool(spec.Lazy),
			LoonFilter:       spec.LoonFilter,
			MihomoOnly:       spec.MihomoOnly,
			ExcludeDNSHijack: spec.ExcludeDNSHijack,
		}
		groups = append(groups, group)
	}
	return domain.ProxyPlan{
		Countries:         countries,
		CountryGroupNames: countryGroupNames,
		Groups:            anchorRepeatedProxies(groups),
	}
}

func hasMissingCountryReference(proxies []string, allCountries, available map[string]bool) bool {
	for _, name := range proxies {
		if allCountries[name] && !available[name] {
			return true
		}
	}
	return false
}

func expandProxyReferences(proxies, countryGroups []string, allCountries, available map[string]bool) []string {
	result := make([]string, 0, len(proxies)+len(countryGroups))
	for _, name := range proxies {
		switch {
		case name == "$CountryGroups":
			result = append(result, countryGroups...)
		case allCountries[name] && !available[name]:
			continue
		default:
			result = append(result, name)
		}
	}
	return result
}

func expandGroupFilter(filter string, classifier *NodeClassifier) string {
	if filter == "$CountryPatterns" {
		return classifier.CountryExcludePattern()
	}
	return filter
}

func anchorRepeatedProxies(groups []domain.ProxyGroup) []domain.ProxyGroup {
	result := append([]domain.ProxyGroup(nil), groups...)
	anchorNumber := 0
	for i := range result {
		if len(result[i].Proxies) == 0 || result[i].ProxiesAlias != "" {
			continue
		}
		anchor := ""
		for j := i + 1; j < len(result); j++ {
			if result[j].ProxiesAlias == "" && slices.Equal(result[i].Proxies, result[j].Proxies) {
				if anchor == "" {
					anchorNumber++
					anchor = fmt.Sprintf("a%d", anchorNumber)
					result[i].ProxiesAnchor = anchor
				}
				result[j].ProxiesAlias = anchor
			}
		}
	}
	return result
}

func clonePolicyConfig(config domain.PolicyConfig) domain.PolicyConfig {
	groups := make([]domain.PolicyGroupSpec, len(config.Groups))
	for i, group := range config.Groups {
		groups[i] = group
		groups[i].Proxies = append([]string(nil), group.Proxies...)
		groups[i].FallbackProxies = append([]string(nil), group.FallbackProxies...)
		groups[i].Lazy = cloneBool(group.Lazy)
	}
	return domain.PolicyConfig{NodeExcludePattern: config.NodeExcludePattern, Groups: groups}
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func validatePolicyConfig(config domain.PolicyConfig, rules []domain.RuleBinding) error {
	groupNames := make(map[string]bool, len(config.Groups))
	generatedNames := make(map[string]bool, len(config.Groups))
	groupsByName := make(map[string]domain.PolicyGroupSpec, len(config.Groups))
	countryNames := make(map[string]bool)
	loonFilters := make(map[string]bool)
	surgeProviders := 0
	for _, group := range config.Groups {
		if groupNames[group.Name] {
			return fmt.Errorf("PolicyConfig.yaml: duplicate policyname %q", group.Name)
		}
		groupNames[group.Name] = true
		groupsByName[group.Name] = group
		if isBuiltinPolicy(group.Name) {
			return fmt.Errorf("PolicyConfig.yaml: policyname %q conflicts with a built-in policy", group.Name)
		}
		if !group.SurgeOnly {
			generatedNames[group.Name] = true
		}
		if strings.TrimSpace(group.Icon) == "" {
			return fmt.Errorf("PolicyConfig.yaml: missing Icon for policyname %q", group.Name)
		}
		switch group.Type {
		case "select", "url-test":
		default:
			return fmt.Errorf("PolicyConfig.yaml: policyname %q has unsupported Type %q", group.Name, group.Type)
		}
		if group.Type != "select" && strings.TrimSpace(group.URL) == "" {
			return fmt.Errorf("PolicyConfig.yaml: policyname %q requires URL", group.Name)
		}
		if group.Filter != "" {
			if _, err := regexp.Compile(group.Filter); err != nil {
				return fmt.Errorf("PolicyConfig.yaml: policyname %q Filter: %w", group.Name, err)
			}
		}
		if group.ExcludeFilter != "" && group.ExcludeFilter != "$CountryPatterns" {
			if _, err := regexp.Compile(group.ExcludeFilter); err != nil {
				return fmt.Errorf("PolicyConfig.yaml: policyname %q ExcludeFilter: %w", group.Name, err)
			}
		}
		if group.Country != "" {
			if group.SurgeOnly {
				return fmt.Errorf("PolicyConfig.yaml: country policyname %q cannot be SurgeOnly", group.Name)
			}
			if countryNames[group.Country] {
				return fmt.Errorf("PolicyConfig.yaml: duplicate Country %q", group.Country)
			}
			countryNames[group.Country] = true
			if group.Country != "其他" && group.Filter == "" {
				return fmt.Errorf("PolicyConfig.yaml: country policyname %q requires Filter", group.Name)
			}
		}
		if group.SurgeOnly && group.PolicyPath == "" {
			return fmt.Errorf("PolicyConfig.yaml: SurgeOnly policyname %q requires PolicyPath", group.Name)
		}
		if group.SurgeOnly {
			if group.Type != "select" {
				return fmt.Errorf("PolicyConfig.yaml: SurgeOnly policyname %q requires Type select", group.Name)
			}
			surgeProviders++
		}
		if !group.SurgeOnly && !group.MihomoOnly && (group.IncludeAll || group.Filter != "") && group.LoonFilter == "" {
			return fmt.Errorf("PolicyConfig.yaml: policyname %q requires LoonFilter", group.Name)
		}
		if group.LoonFilter != "" {
			if loonFilters[group.LoonFilter] {
				return fmt.Errorf("PolicyConfig.yaml: duplicate LoonFilter %q", group.LoonFilter)
			}
			loonFilters[group.LoonFilter] = true
		}
	}
	if surgeProviders != 1 {
		return fmt.Errorf("PolicyConfig.yaml: exactly one SurgeOnly subscription group is required")
	}
	for _, group := range config.Groups {
		for _, name := range append(append([]string(nil), group.Proxies...), group.FallbackProxies...) {
			if reference, ok := groupsByName[name]; ok && reference.MihomoOnly && !group.MihomoOnly {
				return fmt.Errorf("PolicyConfig.yaml: policyname %q references MihomoOnly group %q", group.Name, name)
			}
			if name == "$CountryGroups" || isBuiltinPolicy(name) || generatedNames[name] {
				continue
			}
			return fmt.Errorf("PolicyConfig.yaml: policyname %q references unknown proxy group %q", group.Name, name)
		}
	}
	for _, rule := range rules {
		if !isBuiltinPolicy(rule.PolicyName) && !generatedNames[rule.PolicyName] {
			return fmt.Errorf("rule %q: policyname %q is not a generated policy group", rule.RuleID, rule.PolicyName)
		}
	}
	return nil
}

func isBuiltinPolicy(name string) bool {
	switch name {
	case "DIRECT", "REJECT", "REJECT-DROP":
		return true
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
		Default:     append([]string(nil), policy.Default...),
		ProxyServer: append([]string(nil), policy.ProxyServer...),
		Direct:      append([]string(nil), policy.Direct...),
		Fallback:    append([]string(nil), policy.Fallback...),
	}
}
