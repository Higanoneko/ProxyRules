package service

import (
	"fmt"

	"github.com/PianCat/ProxyRules/internal/catalog"
	"github.com/PianCat/ProxyRules/internal/domain"
)

type RuleCatalogBuilder struct {
	specIndex map[string]domain.RemoteRuleSpec
}

func NewRuleCatalogBuilder() *RuleCatalogBuilder {
	return &RuleCatalogBuilder{specIndex: catalog.RuleSpecIndex()}
}

func (b *RuleCatalogBuilder) Build(rawRules map[string]any) ([]domain.RuleBinding, error) {
	normalized := flattenRules(rawRules, b.specIndex)
	bindings := make([]domain.RuleBinding, 0, len(catalog.CanonicalRuleOrder()))

	for _, ruleID := range catalog.CanonicalRuleOrder() {
		spec, ok := b.specIndex[ruleID]
		if !ok {
			return nil, fmt.Errorf("missing rule spec: %s", ruleID)
		}

		config, ok := normalized[ruleID]
		if !ok {
			return nil, fmt.Errorf("missing rule config: %s", ruleID)
		}

		binding := domain.RuleBinding{
			RuleID:           ruleID,
			ProviderName:     stringValue(config["name"], ruleID),
			PolicyName:       spec.PolicyName,
			TagName:          spec.TagName,
			MihomoPolicyName: spec.MihomoPolicyName,
			SurgeOption:      spec.SurgeOption,
			Source: domain.RuleSourceRef{
				Category:   stringValue(config["category"], ""),
				Behavior:   stringValue(config["behavior"], "classical"),
				RemoteFile: stringValue(config["remotefile"], ""),
			},
		}

		if binding.Source.Category == "" || binding.Source.RemoteFile == "" {
			return nil, fmt.Errorf("incomplete rule source for %s", ruleID)
		}

		bindings = append(bindings, binding)
	}

	return bindings, nil
}

func flattenRules(raw map[string]any, specIndex map[string]domain.RemoteRuleSpec) map[string]map[string]any {
	flattened := map[string]map[string]any{}
	for key, value := range raw {
		valueMap, ok := toStringAnyMap(value)
		if !ok {
			continue
		}

		if _, hasName := valueMap["name"]; hasName {
			if _, hasCategory := valueMap["category"]; hasCategory {
				flattened[resolveRuleID(key, valueMap, specIndex)] = valueMap
				continue
			}
		}

		nested := flattenRules(valueMap, specIndex)
		for nestedKey, nestedValue := range nested {
			flattened[nestedKey] = nestedValue
		}
	}
	return flattened
}

func resolveRuleID(rawKey string, config map[string]any, specIndex map[string]domain.RemoteRuleSpec) string {
	if _, ok := specIndex[rawKey]; ok {
		return rawKey
	}
	if name, ok := config["name"].(string); ok {
		if _, found := specIndex[name]; found {
			return name
		}
	}
	return rawKey
}

func toStringAnyMap(value any) (map[string]any, bool) {
	if converted, ok := value.(map[string]any); ok {
		return converted, true
	}
	if rawMap, ok := value.(map[any]any); ok {
		converted := make(map[string]any, len(rawMap))
		for rawKey, rawValue := range rawMap {
			stringKey, ok := rawKey.(string)
			if !ok {
				return nil, false
			}
			converted[stringKey] = rawValue
		}
		return converted, true
	}
	return nil, false
}

func stringValue(value any, fallback string) string {
	stringValue, ok := value.(string)
	if !ok {
		return fallback
	}
	return stringValue
}
