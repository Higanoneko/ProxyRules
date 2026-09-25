package render

import (
	"fmt"
	"strings"

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
	providerName := r.providerGroupName()

	return ComposeSectioned(
		head,
		textHeadPlaceholders(plan, r.base),
		map[string]map[string]string{"General": r.generalOverrides(plan)},
		map[string][]string{
			"Proxy":       r.proxyLines(providerName),
			"Proxy Group": r.proxyGroupLines(plan.Proxy.Groups, providerName),
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

func (r *SurgeRenderer) proxyLines(providerName string) []string {
	return []string{
		"# 在此添加你的代理节点",
		fmt.Sprintf("# 使用 #!include <ProfileName>.conf 关联其他配置，或在下方的 `%s` 中添加订阅地址（仅限单条），若使用 include 则需要删除下方的 policy-path 字段", providerName),
	}
}

func (r *SurgeRenderer) providerGroupName() string {
	for _, spec := range r.base.PolicyConfig.Groups {
		if spec.SurgeOnly {
			return spec.Name
		}
	}
	return ""
}

func (r *SurgeRenderer) proxyGroupLines(groups []domain.ProxyGroup, providerName string) []string {
	lines := make([]string, 0, len(groups)+3)
	groupsByName := make(map[string]domain.ProxyGroup, len(groups))
	for _, group := range groups {
		groupsByName[group.Name] = group
	}
	for _, spec := range r.base.PolicyConfig.Groups {
		if spec.SurgeOnly {
			lines = append(lines, "", "# "+spec.Name, fmt.Sprintf(
				"%s = %s, policy-path=%s, update-interval=0, no-alert=0, hidden=0, include-all-proxies=1, icon-url=%s",
				spec.Name, spec.Type, spec.PolicyPath, spec.Icon,
			))
			continue
		}
		if group, ok := groupsByName[spec.Name]; ok && !group.MihomoOnly {
			lines = append(lines, r.proxyGroupLine(group, providerName))
		}
	}
	return lines
}

func (r *SurgeRenderer) proxyGroupLine(group domain.ProxyGroup, providerName string) string {
	switch group.Type {
	case "select":
		if group.IncludeAll {
			if group.ExcludeFilter != "" {
				return fmt.Sprintf(
					"%s = select, include-other-group=%s, update-interval=0, policy-regex-filter=^(?!.*(%s)), icon-url=%s",
					group.Name,
					providerName,
					strings.TrimPrefix(group.ExcludeFilter, "(?i)"),
					group.Icon,
				)
			}
			return fmt.Sprintf("%s = select, include-other-group=%s, update-interval=0, icon-url=%s", group.Name, providerName, group.Icon)
		}
		return fmt.Sprintf("%s = select, %s, icon-url=%s", group.Name, strings.Join(group.Proxies, ", "), group.Icon)
	case "url-test":
		if group.Filter != "" {
			return fmt.Sprintf(
				"%s = smart, include-other-group=%s, update-interval=0, policy-regex-filter=(%s), icon-url=%s",
				group.Name,
				providerName,
				strings.TrimPrefix(group.Filter, "(?i)"),
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
