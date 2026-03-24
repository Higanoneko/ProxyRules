package render

import (
	"testing"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
)

func TestApplyTextPlaceholdersSupportsDollarSyntax(t *testing.T) {
	template := `
[General]
dns-server = "$DNS_IP_list",system
doh-server = "$DNS_Default_DoH_List"
real-ip = "$Fake_IP_Filter_list"
`

	result := applyTextPlaceholders(template, map[string]string{
		"DNS_IP_list":          "119.29.29.29, 1.1.1.1",
		"DNS_Default_DoH_List": "https://doh.pub/dns-query, https://dns.google/dns-query",
		"Fake_IP_Filter_list":  "*.lan, *.local",
	})

	for expected, label := range map[string]string{
		`dns-server = 119.29.29.29, 1.1.1.1,system`:                                  "dns-server",
		`doh-server = https://doh.pub/dns-query, https://dns.google/dns-query`:       "doh-server",
		`real-ip = *.lan, *.local`:                                                    "real-ip",
	} {
		if !containsString(result, expected) {
			t.Fatalf("expected %s placeholder replacement, got:\n%s", label, result)
		}
	}
}

func TestTextHeadPlaceholdersExposeSplitDNSLists(t *testing.T) {
	placeholders := textHeadPlaceholders(domain.PolicyPlan{
		DNS: domain.DNSPolicy{
			BootstrapResolvers: []string{"119.29.29.29", "1.1.1.1"},
			Upstreams: domain.DNSUpstreamPolicy{
				Default:     []string{"https://default.example/dns-query"},
				ProxyServer: []string{"https://proxy.example/dns-query"},
				Direct:      []string{"https://direct.example/dns-query"},
			},
			FakeIPFilter: []string{"*.lan", "*.local"},
		},
	}, repository.BaseData{
		SurgeAlwaysRealIP: []string{"*.lan", "*.local"},
	})

	expectations := map[string]string{
		"DNS_IP_list":               "119.29.29.29, 1.1.1.1",
		"DNS_Default_DoH_List":      "https://default.example/dns-query",
		"DNS_Proxy_Server_DoH_List": "https://proxy.example/dns-query",
		"DNS_Direct_DoH_List":       "https://direct.example/dns-query",
		"Fake_IP_Filter_list":       "*.lan, *.local",
		"Surge_Always_Real_IP_List": "*.lan, *.local",
	}

	for key, expected := range expectations {
		if got := placeholders[key]; got != expected {
			t.Fatalf("unexpected placeholder %s: got %q want %q", key, got, expected)
		}
	}
}

func containsString(content string, expected string) bool {
	return len(content) >= len(expected) && (content == expected || containsStringAtAnyOffset(content, expected))
}

func containsStringAtAnyOffset(content string, expected string) bool {
	for index := 0; index+len(expected) <= len(content); index++ {
		if content[index:index+len(expected)] == expected {
			return true
		}
	}
	return false
}
