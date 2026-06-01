package render

import (
	"fmt"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/catalog"
	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/repository"
)

var surgeSectionOrder = []string{"General", "Proxy", "Proxy Group", "Rule"}

type SurgeRenderer struct {
	base         repository.BaseData
	ruleResolver *RuleResolver
}

func NewSurgeRenderer(base repository.BaseData) *SurgeRenderer {
	return &SurgeRenderer{
		base:         base,
		ruleResolver: NewRuleResolver(base),
	}
}

func (r *SurgeRenderer) Render(plan domain.PolicyPlan) (string, error) {
	head, err := r.base.Head("surge")
	if err != nil {
		return "", err
	}

	rules, err := r.ruleLines(plan.Rules)
	if err != nil {
		return "", err
	}

	return ComposeSectioned(
		head,
		textHeadPlaceholders(plan, r.base),
		map[string]map[string]string{"General": r.generalOverrides(plan)},
		map[string][]string{
			"Proxy":       r.proxyLines(),
			"Proxy Group": r.proxyGroupLines(plan.Proxy.Groups),
			"Rule":        rules,
		},
		surgeSectionOrder,
		[]string{"# Surge Configuration", "# Author: Higanoneko", "# Update: 2025-11-26", "# Surge Version: 5.x"},
	)
}

func (r *SurgeRenderer) generalOverrides(plan domain.PolicyPlan) map[string]string {
	ipv6Value := "true"
	ipv6VIF := "auto"
	if !plan.DNS.IPv6 {
		ipv6Value = "false"
		ipv6VIF = "disabled"
	}

	dns := projectGenericDNS(plan.DNS)
	overrides := map[string]string{
		"wifi-access-http-port":   fmt.Sprintf("%d", plan.Ports.HTTP),
		"wifi-access-socks5-port": fmt.Sprintf("%d", plan.Ports.Socks5),
		"ipv6":                    ipv6Value,
		"ipv6-vif":                ipv6VIF,
		"internet-test-url":       plan.TestURLs.Internet,
		"proxy-test-url":          plan.TestURLs.Proxy,
		"dns-server":              strings.Join(dns.BootstrapResolvers, ", ") + ", system",
		"always-real-ip":          strings.Join(r.base.SurgeAlwaysRealIP, ", "),
	}

	if len(dns.Nameserver) > 0 {
		overrides["encrypted-dns-server"] = strings.Join(dns.Nameserver, ", ")
	}

	return overrides
}

func (r *SurgeRenderer) proxyLines() []string {
	return []string{
		"# 在此添加你的代理节点",
		"# 使用 #!include <ProfileName>.conf 关联其他配置，或在下方的 `Proxies` 中添加订阅地址（仅限单条），若使用 include 则需要删除下方的 policy-path 字段",
	}
}

func (r *SurgeRenderer) proxyGroupLines(groups []domain.ProxyGroup) []string {
	lines := make([]string, 0, len(groups)+3)
	for _, group := range groups {
		if group.Name == "GLOBAL" {
			continue
		}
		lines = append(lines, r.proxyGroupLine(group))
	}
	lines = append(lines,
		"",
		"# Proxies",
		"Proxies = select, policy-path=<Your Node List Link Here>, update-interval=0, no-alert=0, hidden=0, include-all-proxies=1, icon-url=https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Proxy.png",
	)
	return lines
}

func (r *SurgeRenderer) proxyGroupLine(group domain.ProxyGroup) string {
	switch group.Type {
	case "select":
		if group.IncludeAll {
			if group.ExcludeFilter != "" {
				return fmt.Sprintf(
					"%s = select, include-other-group=Proxies, update-interval=0, policy-regex-filter=^(?!.*(%s)), icon-url=%s",
					group.Name,
					r.countryPattern(group.Name),
					group.Icon,
				)
			}
			return fmt.Sprintf("%s = select, include-other-group=Proxies, update-interval=0, icon-url=%s", group.Name, group.Icon)
		}
		return fmt.Sprintf("%s = select, %s, icon-url=%s", group.Name, strings.Join(group.Proxies, ", "), group.Icon)
	case "url-test":
		if group.Filter != "" {
			return fmt.Sprintf(
				"%s = smart, include-other-group=Proxies, update-interval=0, policy-regex-filter=(%s), icon-url=%s",
				group.Name,
				r.countryPattern(group.Name),
				group.Icon,
			)
		}
		return fmt.Sprintf("%s = smart, %s, icon-url=%s", group.Name, strings.Join(group.Proxies, ", "), group.Icon)
	default:
		return fmt.Sprintf("%s = select, %s, icon-url=%s", group.Name, strings.Join(group.Proxies, ", "), group.Icon)
	}
}

func (r *SurgeRenderer) ruleLines(bindings []domain.RuleBinding) ([]string, error) {
	remoteRules, err := r.ruleResolver.SurgeRemoteRules(bindings)
	if err != nil {
		return nil, err
	}
	remoteRules = append(remoteRules,
		"",
		"# 局域网地址",
		"RULE-SET,LAN,DIRECT",
		"",
		"# GeoIP CN",
		"GEOIP,CN,直接连接",
		"",
		"# Final",
		"FINAL,选择代理,dns-failed",
	)
	return remoteRules, nil
}

func (r *SurgeRenderer) countryPattern(groupName string) string {
	countryName := strings.TrimSuffix(groupName, "节点")
	for _, country := range catalog.Countries() {
		if country.Name == countryName {
			return strings.ReplaceAll(country.Pattern, "(?i)", "")
		}
	}
	if countryName != "其他" {
		return ".*"
	}

	patterns := make([]string, 0, len(catalog.Countries()))
	for _, country := range catalog.Countries() {
		patterns = append(patterns, strings.ReplaceAll(country.Pattern, "(?i)", ""))
	}
	return strings.Join(patterns, "|")
}
