package catalog

import "github.com/PianCat/ProxyRules/internal/domain"

var policyTemplates = []domain.PolicyTemplate{
	{
		Name:     "选择代理",
		IconURL:  "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Proxy.png",
		Strategy: domain.StrategySelector,
	},
	{
		Name:     "手动选择",
		IconURL:  "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Round_Robin_1.png",
		Strategy: domain.StrategyManual,
	},
	{Name: "AI", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/AI.png", Strategy: domain.StrategyDefault},
	{Name: "Telegram", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Telegram.png", Strategy: domain.StrategyDefault},
	{Name: "YouTube", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/YouTube.png", Strategy: domain.StrategyDefault},
	{Name: "Netflix", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Netflix.png", Strategy: domain.StrategyDefault},
	{Name: "Spotify", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Spotify.png", Strategy: domain.StrategyDefault},
	{Name: "TikTok", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/TikTok.png", Strategy: domain.StrategyDefault},
	{Name: "Steam", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Steam.png", Strategy: domain.StrategyDefault},
	{Name: "Game", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Game.png", Strategy: domain.StrategyDefault},
	{Name: "E-Hentai", IconURL: "https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Ehentai.png", Strategy: domain.StrategyDefault},
	{Name: "PornSite", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Pornhub.png", Strategy: domain.StrategyDefault},
	{
		Name:                  "US Media",
		IconURL:               "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/United_States.png",
		Strategy:              domain.StrategyMediaPreferred,
		PreferredCountryGroup: "美国节点",
	},
	{
		Name:                  "Taiwan Media",
		IconURL:               "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Taiwan.png",
		Strategy:              domain.StrategyMediaPreferred,
		PreferredCountryGroup: "台湾节点",
	},
	{
		Name:                  "Japan Media",
		IconURL:               "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Japan.png",
		Strategy:              domain.StrategyMediaPreferred,
		PreferredCountryGroup: "日本节点",
	},
	{Name: "Global Media", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/DomesticMedia.png", Strategy: domain.StrategyDefault},
	{Name: "Apple", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Apple.png", Strategy: domain.StrategyDirectFirst},
	{Name: "Microsoft", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Microsoft.png", Strategy: domain.StrategyDirectFirst},
	{Name: "Google", IconURL: "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Google_Search.png", Strategy: domain.StrategyDefault},
	{
		Name:         "Google FCM",
		IconURL:      "https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Firebase.png",
		Strategy:     domain.StrategyFixed,
		FixedProxies: []string{"Google", "直接连接"},
	},
	{
		Name:         "Sogou Privacy",
		IconURL:      "https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Sougou.png",
		Strategy:     domain.StrategyFixed,
		FixedProxies: []string{"直接连接", "REJECT"},
	},
	{
		Name:         "ADBlock",
		IconURL:      "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/AdBlack.png",
		Strategy:     domain.StrategyFixed,
		FixedProxies: []string{"REJECT-DROP", "REJECT", "直接连接"},
	},
	{
		Name:         "直接连接",
		IconURL:      "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Direct.png",
		Strategy:     domain.StrategyFixed,
		FixedProxies: []string{"DIRECT", "选择代理"},
	},
	{
		Name:     "GLOBAL",
		IconURL:  "https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png",
		Strategy: domain.StrategyGlobal,
	},
}

func PolicyTemplates() []domain.PolicyTemplate {
	result := make([]domain.PolicyTemplate, len(policyTemplates))
	copy(result, policyTemplates)
	return result
}
