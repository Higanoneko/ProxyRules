package render

import (
	"fmt"
	"strings"

	"github.com/PianCat/ProxyRules/internal/catalog"
	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/repository"
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
		"Remote Filter": r.remoteFilterLines(),
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
		[]string{"# UpdateTime: 2025.11.05 18:00:00 +0000", "# Author: PianCat"},
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

func (r *LoonRenderer) remoteFilterLines() []string {
	lines := []string{
		"# 全部节点筛选",
		`ALL_Filter = NameRegex, FilterKey = ".*"`,
		"",
	}

	for _, country := range catalog.Countries() {
		lines = append(lines, "# "+country.Name+"节点筛选")
		lines = append(lines, fmt.Sprintf(`%s = NameRegex, FilterKey = "%s"`, filterName(country.Name), country.Pattern))
		lines = append(lines, "")
	}

	lines = append(lines, "# 其他节点筛选（排除以上所有地区）")
	lines = append(lines, "# 使用负向预查来排除特定关键词：不含香港、台湾、新加坡、日本、美国")
	lines = append(lines,
		`Other_Filter = NameRegex, FilterKey = "^(?!.*(香港|港|HK|hk|Hong Kong|HongKong|Hongkong|Hong kong|hongkong|hong kong|🇭🇰|台|新北|彰化|TW|Taiwan|TaiWan|Tai wan|Tai Wan|taiwan|tai wan|🇹🇼|美国|美|US|United States|🇺🇸|日本|川日|东京|大阪|泉日|埼玉|沪日|深日|JP|Japan|🇯🇵|新加坡|坡|狮城|SG|Singapore|🇸🇬))"`,
	)
	return lines
}

func (r *LoonRenderer) proxyGroupLines(groups []domain.ProxyGroup) []string {
	lines := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.Name == "GLOBAL" {
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
			return fmt.Sprintf("%s = select, %s%s", group.Name, loonFilterNameForGroup(group.Name), iconPart)
		}
		return fmt.Sprintf("%s = select, %s%s", group.Name, strings.Join(group.Proxies, ", "), iconPart)
	case "url-test", "fallback":
		if group.Filter != "" {
			return fmt.Sprintf(
				"%s = %s, %s, url = %s, interval = %d, tolerance = %d%s",
				group.Name,
				group.Type,
				loonFilterNameForGroup(group.Name),
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

func filterName(country string) string {
	switch country {
	case "香港":
		return "HK_Filter"
	case "台湾":
		return "TW_Filter"
	case "新加坡":
		return "SG_Filter"
	case "日本":
		return "JP_Filter"
	case "美国":
		return "US_Filter"
	default:
		return "ALL_Filter"
	}
}

func loonFilterNameForGroup(groupName string) string {
	switch groupName {
	case "香港节点":
		return "HK_Filter"
	case "台湾节点":
		return "TW_Filter"
	case "新加坡节点":
		return "SG_Filter"
	case "日本节点":
		return "JP_Filter"
	case "美国节点":
		return "US_Filter"
	case "其他节点":
		return "Other_Filter"
	case "手动选择":
		return "ALL_Filter"
	default:
		return "ALL_Filter"
	}
}
