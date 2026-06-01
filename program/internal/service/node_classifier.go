package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/catalog"
	"github.com/Higanoneko/ProxyRules/internal/domain"
)

type NodeClassifier struct {
	ispExcludeRegex *regexp.Regexp
	countryRegexes  map[string]*regexp.Regexp
	countries       []domain.CountryMeta
}

func NewNodeClassifier() *NodeClassifier {
	countries := catalog.Countries()
	countryRegexes := make(map[string]*regexp.Regexp, len(countries))
	for _, country := range countries {
		countryRegexes[country.Name] = regexp.MustCompile(country.Pattern)
	}
	return &NodeClassifier{
		ispExcludeRegex: regexp.MustCompile(catalog.ISPExcludePattern),
		countryRegexes:  countryRegexes,
		countries:       countries,
	}
}

func (c *NodeClassifier) IdentifyCountry(nodeName string, excludeISP bool) (string, bool) {
	if excludeISP && c.ispExcludeRegex.MatchString(nodeName) {
		return "", false
	}
	for _, country := range c.countries {
		if c.countryRegexes[country.Name].MatchString(nodeName) {
			return country.Name, true
		}
	}
	return "其他", true
}

func (c *NodeClassifier) ParseCountryInfos(nodeNames []string, minCount int) []domain.CountryInfo {
	counts := map[string]int{}
	for _, nodeName := range nodeNames {
		country, ok := c.IdentifyCountry(nodeName, true)
		if !ok {
			continue
		}
		counts[country]++
	}

	results := make([]domain.CountryInfo, 0, len(c.countries)+1)
	for _, country := range c.countries {
		count := counts[country.Name]
		if count < minCount {
			continue
		}
		results = append(results, domain.CountryInfo{
			Name:    country.Name,
			Count:   count,
			Pattern: country.Pattern,
			IconURL: country.IconURL,
		})
	}

	if otherCount := counts["其他"]; otherCount > 0 {
		results = append(results, domain.CountryInfo{
			Name:    "其他",
			Count:   otherCount,
			Pattern: c.OtherPattern(),
			IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png",
		})
	}

	return results
}

func (c *NodeClassifier) DefaultCountryInfos() []domain.CountryInfo {
	results := make([]domain.CountryInfo, 0, len(c.countries)+1)
	for _, country := range c.countries {
		results = append(results, domain.CountryInfo{
			Name:    country.Name,
			Count:   0,
			Pattern: country.Pattern,
			IconURL: country.IconURL,
		})
	}
	results = append(results, domain.CountryInfo{
		Name:    "其他",
		Count:   0,
		Pattern: c.OtherPattern(),
		IconURL: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png",
	})
	return results
}

func (c *NodeClassifier) OtherPattern() string {
	excludePatterns := []string{
		strings.ReplaceAll(catalog.ISPExcludePattern, "(?i)", ""),
	}
	for _, country := range c.countries {
		excludePatterns = append(excludePatterns, strings.ReplaceAll(country.Pattern, "(?i)", ""))
	}
	return fmt.Sprintf("^(?!.*(%s)).*$", strings.Join(excludePatterns, "|"))
}

func (c *NodeClassifier) CountryExcludePattern() string {
	patterns := make([]string, 0, len(c.countries))
	for _, country := range c.countries {
		patterns = append(patterns, strings.ReplaceAll(country.Pattern, "(?i)", ""))
	}
	return "(?i)" + strings.Join(patterns, "|")
}
