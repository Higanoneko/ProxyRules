package render

import (
	"strings"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
	"gopkg.in/yaml.v3"
)

func yamlHeadPlaceholders(plan domain.PolicyPlan) map[string]any {
	return map[string]any{
		"DNS_IP_List":         append([]string(nil), plan.DNS.DefaultNameserver...),
		"DNS_DoH_List":        append([]string(nil), plan.DNS.Nameserver...),
		"Fake_IP_Filter_List": append([]string(nil), plan.DNS.FakeIPFilter...),
	}
}

func textHeadPlaceholders(plan domain.PolicyPlan, base repository.BaseData) map[string]string {
	return map[string]string{
		"DNS_IP_list":               strings.Join(plan.DNS.DefaultNameserver, ", "),
		"DNS_DoH_list":              strings.Join(plan.DNS.Nameserver, ", "),
		"Fake_IP_Filter_list":       strings.Join(plan.DNS.FakeIPFilter, ", "),
		"Surge_Always_Real_IP_List": strings.Join(base.SurgeAlwaysRealIP, ", "),
	}
}

func mihomoDNSNode(plan domain.PolicyPlan, includeEnable bool) *yaml.Node {
	dns := newMappingNode()
	if includeEnable {
		appendMappingValue(dns, "enable", newScalarNode(plan.DNS.Enable))
	}
	appendMappingValue(dns, "ipv6", newScalarNode(plan.DNS.IPv6))
	appendMappingValue(dns, "enhanced-mode", newScalarNode(plan.DNS.EnhancedMode))
	appendMappingValue(dns, "default-nameserver", stringSequenceNode(plan.DNS.DefaultNameserver))
	appendMappingValue(dns, "nameserver", stringSequenceNode(plan.DNS.Nameserver))
	appendMappingValue(dns, "fake-ip-filter", stringSequenceNode(plan.DNS.FakeIPFilter))
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
