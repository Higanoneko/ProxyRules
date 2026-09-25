package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Higanoneko/ProxyRules/internal/domain"
)

type NodeClassifier struct {
	excludePattern string
	excludeRegex   *regexp.Regexp
	countryRegexes map[string]*regexp.Regexp
	countries      []domain.CountryMeta
}

func NewNodeClassifier(config domain.PolicyConfig) (*NodeClassifier, error) {
	excludeRegex, err := regexp.Compile(config.NodeExcludePattern)
	if err != nil {
		return nil, fmt.Errorf("NodeExcludePattern: %w", err)
	}
	countries := make([]domain.CountryMeta, 0)
	countryRegexes := make(map[string]*regexp.Regexp)
	for _, group := range config.Groups {
		if group.Country == "" || group.Country == "其他" {
			continue
		}
		countryRegex, err := regexp.Compile(group.Filter)
		if err != nil {
			return nil, fmt.Errorf("policyname %q Filter: %w", group.Name, err)
		}
		countries = append(countries, domain.CountryMeta{Name: group.Country, Pattern: group.Filter})
		countryRegexes[group.Country] = countryRegex
	}
	return &NodeClassifier{
		excludePattern: config.NodeExcludePattern,
		excludeRegex:   excludeRegex,
		countryRegexes: countryRegexes,
		countries:      countries,
	}, nil
}

func (c *NodeClassifier) IdentifyCountry(nodeName string, excludeISP bool) (string, bool) {
	if excludeISP && c.excludeRegex.MatchString(nodeName) {
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
		if ok {
			counts[country]++
		}
	}
	results := make([]domain.CountryInfo, 0, len(c.countries)+1)
	for _, country := range c.countries {
		if count := counts[country.Name]; count >= minCount {
			results = append(results, domain.CountryInfo{Name: country.Name, Count: count, Pattern: country.Pattern})
		}
	}
	if count := counts["其他"]; count > 0 {
		results = append(results, domain.CountryInfo{Name: "其他", Count: count, Pattern: c.OtherPattern()})
	}
	return results
}

func (c *NodeClassifier) DefaultCountryInfos() []domain.CountryInfo {
	results := make([]domain.CountryInfo, 0, len(c.countries)+1)
	for _, country := range c.countries {
		results = append(results, domain.CountryInfo{Name: country.Name, Pattern: country.Pattern})
	}
	return append(results, domain.CountryInfo{Name: "其他", Pattern: c.OtherPattern()})
}

func (c *NodeClassifier) OtherPattern() string {
	patterns := []string{stripCaseFlag(c.excludePattern)}
	for _, country := range c.countries {
		patterns = append(patterns, stripCaseFlag(country.Pattern))
	}
	return fmt.Sprintf("^(?!.*(%s)).*$", strings.Join(patterns, "|"))
}

func (c *NodeClassifier) CountryExcludePattern() string {
	patterns := make([]string, 0, len(c.countries))
	for _, country := range c.countries {
		patterns = append(patterns, stripCaseFlag(country.Pattern))
	}
	return "(?i)" + strings.Join(patterns, "|")
}

func stripCaseFlag(pattern string) string {
	return strings.TrimPrefix(pattern, "(?i)")
}
