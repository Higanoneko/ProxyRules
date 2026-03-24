package render

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	sectionPattern    = regexp.MustCompile(`^\[(?P<name>[^\]]+)\]\s*$`)
	assignmentPattern = regexp.MustCompile(`^(?P<indent>\s*)(?P<key>[^#;\s][^=]*?)\s*=\s*(?P<value>.*)$`)
)

type parsedSections struct {
	preamble []string
	order    []string
	sections map[string][]string
}

func ComposeSectioned(
	templateContent string,
	placeholders map[string]string,
	assignmentOverrides map[string]map[string]string,
	replacementSections map[string][]string,
	sectionOrder []string,
	fallbackPreamble []string,
) (string, error) {
	preparedTemplate := applyTextPlaceholders(templateContent, placeholders)
	parsed := parseSections(preparedTemplate)
	effectivePreamble := parsed.preamble
	if len(effectivePreamble) == 0 {
		effectivePreamble = append([]string(nil), fallbackPreamble...)
	}

	orderedNames := make([]string, 0, len(parsed.order)+len(sectionOrder))
	seen := map[string]struct{}{}
	for _, name := range parsed.order {
		orderedNames = appendIfMissing(orderedNames, seen, name)
	}
	for _, name := range sectionOrder {
		orderedNames = appendIfMissing(orderedNames, seen, name)
	}
	for name := range assignmentOverrides {
		orderedNames = appendIfMissing(orderedNames, seen, name)
	}
	for name := range replacementSections {
		orderedNames = appendIfMissing(orderedNames, seen, name)
	}

	blocks := make([]string, 0, len(orderedNames)+1)
	if len(effectivePreamble) > 0 {
		blocks = append(blocks, strings.TrimRight(strings.Join(effectivePreamble, "\n"), "\n"))
	}

	for _, sectionName := range orderedNames {
		body, ok := resolveSectionBody(
			sectionName,
			parsed.sections[sectionName],
			assignmentOverrides[sectionName],
			replacementSections[sectionName],
		)
		if !ok {
			continue
		}

		sectionBlock := append([]string{fmt.Sprintf("[%s]", sectionName)}, body...)
		blocks = append(blocks, strings.TrimRight(strings.Join(sectionBlock, "\n"), "\n"))
	}

	return strings.TrimRight(strings.Join(blocks, "\n\n"), "\n") + "\n", nil
}

func applyTextPlaceholders(templateContent string, placeholders map[string]string) string {
	preparedContent := templateContent
	for key, value := range placeholders {
		for _, token := range []string{
			`"$` + key + `"`,
			`'$` + key + `'`,
			"{" + key + "}",
			"$" + key,
		} {
			preparedContent = strings.ReplaceAll(preparedContent, token, value)
		}
	}
	return preparedContent
}

func parseSections(templateContent string) parsedSections {
	result := parsedSections{
		preamble: []string{},
		order:    []string{},
		sections: map[string][]string{},
	}

	var currentSection string
	var currentBody []string

	for _, line := range strings.Split(templateContent, "\n") {
		match := sectionPattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 0 {
			if currentSection == "" {
				result.preamble = append(result.preamble, line)
			} else {
				currentBody = append(currentBody, line)
			}
			continue
		}

		if currentSection != "" {
			result.order = append(result.order, currentSection)
			result.sections[currentSection] = currentBody
		}

		currentSection = match[1]
		currentBody = []string{}
	}

	if currentSection != "" {
		result.order = append(result.order, currentSection)
		result.sections[currentSection] = currentBody
	}

	return result
}

func resolveSectionBody(
	sectionName string,
	templateBody []string,
	assignments map[string]string,
	replacementBody []string,
) ([]string, bool) {
	_ = sectionName
	if replacementBody != nil {
		return append([]string(nil), replacementBody...), true
	}
	if assignments != nil {
		return mergeAssignments(templateBody, assignments), true
	}
	if templateBody != nil {
		return append([]string(nil), templateBody...), true
	}
	return nil, false
}

func mergeAssignments(templateBody []string, assignments map[string]string) []string {
	merged := make([]string, 0, len(templateBody)+len(assignments))
	used := map[string]struct{}{}

	for _, line := range templateBody {
		match := assignmentPattern.FindStringSubmatch(line)
		if len(match) == 0 {
			merged = append(merged, line)
			continue
		}

		key := strings.TrimSpace(match[2])
		value, ok := assignments[key]
		if !ok {
			merged = append(merged, line)
			continue
		}

		used[key] = struct{}{}
		merged = append(merged, match[1]+key+" = "+value)
	}

	for key, value := range assignments {
		if _, ok := used[key]; ok {
			continue
		}
		merged = append(merged, key+" = "+value)
	}

	return merged
}

func appendIfMissing(values []string, seen map[string]struct{}, value string) []string {
	if _, ok := seen[value]; ok {
		return values
	}
	seen[value] = struct{}{}
	return append(values, value)
}
