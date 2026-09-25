package render

import (
	"fmt"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
)

var loonSectionOrder = []string{
	"General",
	"Plugin",
	"Proxy",
	"Remote Proxy",
	"Remote Filter",
	"Proxy Group",
	"Remote Rule",
	"Rule",
}

type LoonRenderer struct {
	base         repository.BaseData
	ruleResolver *RuleResolver
}

func NewLoonRenderer(base repository.BaseData) *LoonRenderer {
	return &LoonRenderer{
		base:         base,
		ruleResolver: NewRuleResolver(base),
	}
}

func (r *LoonRenderer) Render(plan domain.PolicyPlan) (string, error) {
	head, err := r.base.Head("loon")
	if err != nil {
		return "", err
	}

	remoteRules, err := r.ruleResolver.LoonRemoteRules(plan.Rules)
	if err != nil {
		return "", err
	}

	replacements := map[string][]string{
		"Remote Filter": r.remoteFilterLines(plan.Proxy.Groups),
		"Proxy Group":   r.proxyGroupLines(plan.Proxy.Groups),
		"Remote Rule":   remoteRules,
		"Rule":          []string{"GEOIP, CN, 直接连接", "FINAL, 选择代理"},
	}

	if !strings.Contains(head, "[Plugin]") {
		replacements["Plugin"] = []string{"https://raw.githubusercontent.com/Peng-YM/Loon-Gallery/master/loon-gallery.plugin, enable = true"}
	}
	if !strings.Contains(head, "[Proxy]") {
		replacements["Proxy"] = []string{"# Your Proxy Nodes Here"}
	}
	if !strings.Contains(head, "[Remote Proxy]") {
		replacements["Remote Proxy"] = []string{"# Your Node or Proxy Subscription Links Here"}
	}

	return ComposeSectioned(
		head,
		textHeadPlaceholders(plan, r.base),
		map[string]map[string]string{"General": r.generalOverrides(plan)},
		replacements,
		loonSectionOrder,
		[]string{"# UpdateTime: 2025.11.05 18:00:00 +0000", "# Author: Higanoneko"},
	)
}

func (r *LoonRenderer) generalOverrides(plan domain.PolicyPlan) map[string]string {
	ipMode := "dual"
	ipv6VIF := "auto"
	if !plan.DNS.IPv6 {
		ipMode = "ipv4-only"
		ipv6VIF = "off"
	}

	dns := projectGenericDNS(plan.DNS)
	return map[string]string{
		"ip-mode":                  ipMode,
		"ipv6-vif":                 ipv6VIF,
		"dns-server":               strings.Join(dns.BootstrapResolvers, ", ") + ", system",
		"doh-server":               strings.Join(dns.Nameserver, ", "),
		"wifi-access-http-port":    fmt.Sprintf("%d", plan.Ports.HTTP),
		"wifi-access-socket5-port": fmt.Sprintf("%d", plan.Ports.Socks5),
		"internet-test-url":        plan.TestURLs.Internet,
		"proxy-test-url":           plan.TestURLs.Proxy,
		"real-ip":                  strings.Join(dns.FakeIPFilter, ", "),
	}
}

func (r *LoonRenderer) remoteFilterLines(groups []domain.ProxyGroup) []string {
	lines := make([]string, 0, len(groups)*3)
	for _, group := range groups {
		if group.LoonFilter == "" {
			continue
		}
		pattern := group.Filter
		if group.ExcludeFilter != "" {
			pattern = fmt.Sprintf("^(?!.*(%s))", strings.TrimPrefix(group.ExcludeFilter, "(?i)"))
		} else if pattern == "" {
			pattern = ".*"
		}
		lines = append(lines, "# "+group.Name+"筛选")
		lines = append(lines, fmt.Sprintf("%s = NameRegex, FilterKey = %q", group.LoonFilter, pattern), "")
	}
	return lines
}

func (r *LoonRenderer) proxyGroupLines(groups []domain.ProxyGroup) []string {
	lines := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.MihomoOnly {
			continue
		}
		lines = append(lines, r.proxyGroupLine(group))
	}
	return lines
}

func (r *LoonRenderer) proxyGroupLine(group domain.ProxyGroup) string {
	iconPart := ""
	if group.Icon != "" {
		iconPart = ", img-url = " + group.Icon
	}

	switch group.Type {
	case "select":
		if group.IncludeAll {
			return fmt.Sprintf("%s = select, %s%s", group.Name, group.LoonFilter, iconPart)
		}
		return fmt.Sprintf("%s = select, %s%s", group.Name, strings.Join(group.Proxies, ", "), iconPart)
	case "url-test", "fallback":
		if group.Filter != "" {
			return fmt.Sprintf(
				"%s = %s, %s, url = %s, interval = %d, tolerance = %d%s",
				group.Name,
				group.Type,
				group.LoonFilter,
				group.URL,
				group.Interval,
				group.Tolerance,
				iconPart,
			)
		}
		return fmt.Sprintf(
			"%s = %s, %s, url = %s, interval = %d, tolerance = %d%s",
			group.Name,
			group.Type,
			strings.Join(group.Proxies, ", "),
			group.URL,
			group.Interval,
			group.Tolerance,
			iconPart,
		)
	default:
		return fmt.Sprintf("%s = select, %s%s", group.Name, strings.Join(group.Proxies, ", "), iconPart)
	}
}
