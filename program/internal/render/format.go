package render

import "strings"

func addBlankLinesBeforeSections(content string, sections []string) string {
	lines := strings.Split(content, "\n")
	rewritten := make([]string, 0, len(lines)+len(sections))

	for index, line := range lines {
		lineTrimmed := strings.TrimSpace(line)
		for _, section := range sections {
			if !strings.HasPrefix(lineTrimmed, section) {
				continue
			}
			if index > 0 && len(rewritten) > 0 && strings.TrimSpace(rewritten[len(rewritten)-1]) != "" {
				rewritten = append(rewritten, "")
			}
			break
		}
		rewritten = append(rewritten, line)
	}

	return strings.Join(rewritten, "\n")
}
