package service

import (
	"fmt"

	"github.com/Higanoneko/ProxyRules/internal/domain"
)

type RuleCatalogBuilder struct{}

func NewRuleCatalogBuilder() *RuleCatalogBuilder {
	return &RuleCatalogBuilder{}
}

func (b *RuleCatalogBuilder) Build(sections domain.RuleSections) ([]domain.RuleBinding, error) {
	allEntries := mergeEntries(sections.BaseRules.Entries, sections.CustomRules.Entries)
	bindings := make([]domain.RuleBinding, 0, len(sections.BaseRules.Keys)+len(sections.CustomRules.Keys))

	sectionsList := []domain.OrderedSection{sections.BaseRules, sections.CustomRules}
	for _, section := range sectionsList {
		for _, ruleID := range section.Keys {
			config := section.Entries[ruleID]
			binding, err := b.buildBinding(ruleID, config, allEntries)
			if err != nil {
				return nil, fmt.Errorf("rule %s: %w", ruleID, err)
			}
			bindings = append(bindings, binding)
		}
	}

	return bindings, nil
}

func (b *RuleCatalogBuilder) buildBinding(ruleID string, config map[string]any, allEntries map[string]map[string]any) (domain.RuleBinding, error) {
	policyName := stringValue(config["policyname"], "")
	if policyName == "" {
		if parentTag, ok := config["parenttag"].(string); ok && parentTag != "" {
			var err error
			policyName, err = resolveParentPolicyName(parentTag, allEntries)
			if err != nil {
				return domain.RuleBinding{}, err
			}
		}
	}
	if policyName == "" {
		return domain.RuleBinding{}, fmt.Errorf("no policyname configured and no valid parenttag")
	}

	tagName := stringValue(config["tagname"], "")
	if tagName == "" {
		tagName = stringValue(config["name"], ruleID)
	}

	return domain.RuleBinding{
		RuleID:       ruleID,
		ProviderName: stringValue(config["name"], ruleID),
		PolicyName:   policyName,
		TagName:      tagName,
		SurgeOption:  stringValue(config["surgeoption"], ""),
		Source: domain.RuleSourceRef{
			Category:   stringValue(config["category"], ""),
			Behavior:   stringValue(config["behavior"], "classical"),
			RemoteFile: stringValue(config["remotefile"], ""),
		},
	}, nil
}

func resolveParentPolicyName(parentTag string, allEntries map[string]map[string]any) (string, error) {
	parentConfig, ok := allEntries[parentTag]
	if !ok {
		return "", fmt.Errorf("unknown parenttag %q", parentTag)
	}
	if policyName := stringValue(parentConfig["policyname"], ""); policyName != "" {
		return policyName, nil
	}
	return "", fmt.Errorf("parent rule %q has no policyname configured", parentTag)
}

func mergeEntries(a, b map[string]map[string]any) map[string]map[string]any {
	merged := make(map[string]map[string]any, len(a)+len(b))
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] = v
	}
	return merged
}

func stringValue(value any, fallback string) string {
	stringValue, ok := value.(string)
	if !ok {
		return fallback
	}
	return stringValue
}
