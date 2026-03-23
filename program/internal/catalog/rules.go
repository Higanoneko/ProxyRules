package catalog

import "github.com/PianCat/ProxyRules/internal/domain"

var remoteRuleSpecs = []domain.RemoteRuleSpec{
	{RuleID: "AI", PolicyName: "AI", TagName: "AI"},
	{RuleID: "Telegram", PolicyName: "Telegram", TagName: "Telegram"},
	{RuleID: "YouTube", PolicyName: "YouTube", TagName: "YouTube"},
	{RuleID: "YouTubeMusic", PolicyName: "YouTube", TagName: "YouTube Music"},
	{RuleID: "Netflix", PolicyName: "Netflix", TagName: "Netflix"},
	{RuleID: "TikTok", PolicyName: "TikTok", TagName: "TikTok"},
	{RuleID: "Spotify", PolicyName: "Spotify", TagName: "Spotify"},
	{RuleID: "Steam", PolicyName: "Steam", TagName: "Steam"},
	{RuleID: "Game", PolicyName: "Game", TagName: "Game"},
	{RuleID: "E-Hentai", PolicyName: "E-Hentai", TagName: "E-Hentai"},
	{RuleID: "PornSite", PolicyName: "PornSite", TagName: "PornSite"},
	{RuleID: "Furrybar", PolicyName: "PornSite", TagName: "Furrybar"},
	{RuleID: "Stream_US", PolicyName: "US Media", TagName: "US Media"},
	{RuleID: "Stream_TW", PolicyName: "Taiwan Media", TagName: "Taiwan Media"},
	{RuleID: "Playhorny", PolicyName: "Taiwan Media", TagName: "Playhorny"},
	{RuleID: "Stream_JP", PolicyName: "Japan Media", TagName: "Japan Media"},
	{RuleID: "Stream_Global", PolicyName: "Global Media", TagName: "Global Media"},
	{RuleID: "Apple", PolicyName: "Apple", TagName: "Apple"},
	{RuleID: "Microsoft", PolicyName: "Microsoft", TagName: "Microsoft"},
	{RuleID: "Google", PolicyName: "Google", TagName: "Google"},
	{RuleID: "GoogleFCM", PolicyName: "Google FCM", TagName: "Google FCM"},
	{RuleID: "SogouPrivacy", PolicyName: "Sogou Privacy", TagName: "Sogou Privacy"},
	{RuleID: "ADBlock", PolicyName: "ADBlock", TagName: "ADBlock", SurgeOption: "extended-matching"},
	{RuleID: "LocalNetwork", PolicyName: "DIRECT", TagName: "LocalNetwork", MihomoPolicyName: "DIRECT"},
	{RuleID: "LocalNetworkIP", PolicyName: "DIRECT", TagName: "LocalNetworkIP", MihomoPolicyName: "DIRECT"},
}

var canonicalRuleOrder = []string{
	"AI",
	"Telegram",
	"YouTube",
	"YouTubeMusic",
	"Netflix",
	"TikTok",
	"Spotify",
	"Steam",
	"Game",
	"E-Hentai",
	"PornSite",
	"Furrybar",
	"Stream_US",
	"Stream_TW",
	"Playhorny",
	"Stream_JP",
	"Stream_Global",
	"Apple",
	"Microsoft",
	"Google",
	"GoogleFCM",
	"SogouPrivacy",
	"ADBlock",
	"LocalNetwork",
	"LocalNetworkIP",
}

func RuleSpecs() []domain.RemoteRuleSpec {
	result := make([]domain.RemoteRuleSpec, len(remoteRuleSpecs))
	copy(result, remoteRuleSpecs)
	return result
}

func RuleSpecIndex() map[string]domain.RemoteRuleSpec {
	index := make(map[string]domain.RemoteRuleSpec, len(remoteRuleSpecs))
	for _, spec := range remoteRuleSpecs {
		index[spec.RuleID] = spec
	}
	return index
}

func CanonicalRuleOrder() []string {
	result := make([]string, len(canonicalRuleOrder))
	copy(result, canonicalRuleOrder)
	return result
}
