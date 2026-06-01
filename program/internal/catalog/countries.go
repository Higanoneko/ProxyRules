package catalog

import "github.com/Higanoneko/ProxyRules/internal/domain"

const ISPExcludePattern = `(?i)家宽|家庭|家庭宽带|商宽|商业宽带|星链|Starlink|落地`

var defaultCountries = []domain.CountryMeta{
	{
		Name:    "香港",
		Pattern: `(?i)香港|港|HK|hk|Hong Kong|HongKong|Hongkong|Hong kong|hongkong|hong kong|🇭🇰`,
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Hong_Kong.png",
	},
	{
		Name:    "台湾",
		Pattern: `(?i)台|新北|彰化|TW|Taiwan|TaiWan|Tai wan|Tai Wan|taiwan|tai wan|🇹🇼`,
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Taiwan.png",
	},
	{
		Name:    "新加坡",
		Pattern: `(?i)新加坡|坡|狮城|SG|Singapore|🇸🇬`,
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Singapore.png",
	},
	{
		Name:    "日本",
		Pattern: `(?i)日本|川日|东京|大阪|泉日|埼玉|沪日|深日|JP|Japan|🇯🇵`,
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Japan.png",
	},
	{
		Name:    "美国",
		Pattern: `(?i)美国|美|US|United States|🇺🇸`,
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/United_States.png",
	},
}

func Countries() []domain.CountryMeta {
	result := make([]domain.CountryMeta, len(defaultCountries))
	copy(result, defaultCountries)
	return result
}
