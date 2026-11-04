package render

import (
	"fmt"
	"strings"

	"github.com/luluape/ProxyRules/internal/domain"
	"gopkg.in/yaml.v3"
)

type EgernRenderer struct{}

func NewEgernRenderer() *EgernRenderer {
	return &EgernRenderer{}
}

type egernPolicyGroup struct {
	Name           string   `yaml:"name"`
	Policies       []string `yaml:"policies,omitempty"`
	URLs           []string `yaml:"urls,omitempty"`
	Flatten        bool     `yaml:"flatten,omitempty"`
	Filter         string   `yaml:"filter,omitempty"`
	Icon           string   `yaml:"icon,omitempty"`
	Interval       int      `yaml:"interval,omitempty"`
	Tolerance      int      `yaml:"tolerance,omitempty"`
	LatencyTestURL string   `yaml:"latency_test_url,omitempty"`
	UpdateInterval int      `yaml:"update_interval,omitempty"`
}

type egernPolicyDocument struct {
	PolicyGroups []map[string]egernPolicyGroup `yaml:"policy_groups"`
}

// RenderPolicyGroups renders shared groups from PolicyPlan.
// Subscription URLs are supplied separately and are not embedded
// in the public repository.
func (r *EgernRenderer) RenderPolicyGroups(
	plan domain.PolicyPlan,
	subscriptionURLs []string,
) (string, error) {
	if len(subscriptionURLs) == 0 {
		return "", fmt.Errorf("Egern requires a node subscription")
	}
	for _, url := range subscriptionURLs {
		if !strings.HasPrefix(url, "https://") {
			return "", fmt.Errorf("invalid Egern subscription URL")
		}
	}

	groups := make([]map[string]egernPolicyGroup, 0)
	names := make(map[string]bool)

	for _, group := range plan.Proxy.Groups {
		if group.MihomoOnly {
			continue
		}
		if names[group.Name] {
			return "", fmt.Errorf("duplicate Egern group %q", group.Name)
		}
		names[group.Name] = true

		if strings.TrimSpace(group.Icon) == "" {
			return "", fmt.Errorf("missing icon: %s", group.Name)
		}

		entry := egernPolicyGroup{
			Name:     group.Name,
			Icon:     group.Icon,
			Policies: make([]string, 0, len(group.Proxies)),
		}

		for _, name := range group.Proxies {
			if name == "REJECT-DROP" {
				name = "REJECT"
			}
			found := false
			for _, existing := range entry.Policies {
				if existing == name {
					found = true
					break
				}
			}
			if !found {
				entry.Policies = append(entry.Policies, name)
			}
		}

		var kind string
		switch group.Type {
		case "select":
			kind = "select"
		case "url-test":
			kind = "auto_test"
			entry.Interval = group.Interval
			entry.Tolerance = group.Tolerance
			entry.LatencyTestURL = group.URL
		default:
			return "", fmt.Errorf(
				"unsupported Egern type %q: %s",
				group.Type, group.Name,
			)
		}

		switch {
		case group.Name == "手动选择":
			// Central subscription group; all regional groups
			// expand nodes from this source.
			entry.Policies = nil
			entry.URLs = append([]string(nil), subscriptionURLs...)
			entry.UpdateInterval = 86400

		case group.IncludeAll:
			entry.Policies = []string{"手动选择"}
			entry.Flatten = true

			if group.ExcludeFilter != "" {
				// Negative lookahead: keep nodes not matching
				// any existing country filter.
				pattern := strings.TrimPrefix(
					group.ExcludeFilter, "(?i)",
				)
				entry.Filter = "(?i)^(?!.*(?:" + pattern + ")).*$"
			} else {
				entry.Filter = group.Filter
			}
		}

		groups = append(groups,
			map[string]egernPolicyGroup{kind: entry},
		)
	}

	if !names["手动选择"] {
		return "", fmt.Errorf("missing 手动选择 group")
	}

	for _, item := range groups {
		for _, body := range item {
			for _, reference := range body.Policies {
				if reference == "DIRECT" || reference == "REJECT" {
					continue
				}
				if !names[reference] {
					return "", fmt.Errorf(
						"group %q references unknown %q",
						body.Name, reference,
					)
				}
			}
		}
	}

	data, err := yaml.Marshal(
		egernPolicyDocument{PolicyGroups: groups},
	)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
