package render

import (
	"strings"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

func yamlHeadPlaceholders(plan domain.PolicyPlan) map[string]any {
	dns := projectGenericDNS(plan.DNS)
	return map[string]any{
		"DNS_IP_List":         append([]string(nil), dns.BootstrapResolvers...),
		"DNS_DoH_List":        append([]string(nil), dns.Nameserver...),
		"Fake_IP_Filter_List": append([]string(nil), dns.FakeIPFilter...),
	}
}

func textHeadPlaceholders(plan domain.PolicyPlan, base repository.BaseData) map[string]string {
	dns := projectGenericDNS(plan.DNS)
	return map[string]string{
		"DNS_IP_list":               strings.Join(dns.BootstrapResolvers, ", "),
		"DNS_DoH_list":              strings.Join(dns.Nameserver, ", "),
		"Fake_IP_Filter_list":       strings.Join(dns.FakeIPFilter, ", "),
		"Surge_Always_Real_IP_List": strings.Join(base.SurgeAlwaysRealIP, ", "),
	}
}

func mihomoDNSNode(plan domain.PolicyPlan, includeEnable bool) *yaml.Node {
	projected := projectClashDNS(plan.DNS)
	dns := newMappingNode()
	if includeEnable {
		appendMappingValue(dns, "enable", newScalarNode(projected.Enable))
	}
	appendMappingValue(dns, "ipv6", newScalarNode(projected.IPv6))
	appendMappingValue(dns, "enhanced-mode", newScalarNode(projected.EnhancedMode))
	appendMappingValue(dns, "default-nameserver", stringSequenceNode(projected.DefaultNameserver))
	if projected.HasScopedResolvers {
		appendMappingValue(dns, "proxy-server-nameserver", stringSequenceNode(projected.ProxyServerNameserver))
		appendMappingValue(dns, "direct-nameserver", stringSequenceNode(projected.DirectNameserver))
	}
	appendMappingValue(dns, "nameserver", stringSequenceNode(projected.Nameserver))
	appendMappingValue(dns, "fake-ip-filter", stringSequenceNode(projected.FakeIPFilter))
	return dns
}

func geoxNode() *yaml.Node {
	geox := newMappingNode()
	for _, key := range []string{"geoip", "geosite", "mmdb", "asn"} {
		appendMappingValue(geox, key, newScalarNode(repository.DefaultGeoXURLs()[key]))
	}
	return geox
}

func stringSequenceNode(values []string) *yaml.Node {
	sequence := newSequenceNode()
	for _, value := range values {
		sequence.Content = append(sequence.Content, newScalarNode(value))
	}
	return sequence
}
