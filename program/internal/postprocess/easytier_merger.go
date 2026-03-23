package postprocess

import (
	"regexp"
	"strings"
)

const (
	configStart = "// ============================================\n// Wireguard_Easytier 用户配置区域"
	configEnd   = "// Wireguard_Easytier 用户配置区域结束\n// ============================================"
)

var mainFunctionPattern = regexp.MustCompile(`\bfunction\s+main\s*\(`)

func MergeEasytier(convertScript string, easytierScript string) string {
	configBlock, logicBlock := splitConfigAndLogic(easytierScript)
	renamedSource := renameMainFunction(convertScript, "_originalMain")
	renamedSource = insertConfigAfterHeader(renamedSource, configBlock)
	renamedLogic := renameMainFunction(logicBlock, "_easytierEnhance")

	parts := []string{
		strings.TrimRight(renamedSource, "\n"),
		"",
		"// ============ Wireguard_Easytier Start ============",
		strings.TrimSpace(renamedLogic),
		"// ============ Wireguard_Easytier End ============",
		"",
		buildBridgeFunction(),
	}
	return strings.Join(parts, "\n")
}

func renameMainFunction(content string, newName string) string {
	firstMatch := mainFunctionPattern.FindStringIndex(content)
	if firstMatch == nil {
		return content
	}
	return content[:firstMatch[0]] + "function " + newName + "(" + content[firstMatch[1]:]
}

func splitConfigAndLogic(content string) (string, string) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	startIndex := strings.Index(normalized, configStart)
	endIndex := strings.Index(normalized, configEnd)
	if startIndex < 0 || endIndex < 0 {
		return "", normalized
	}

	configEndIndex := endIndex + len(configEnd)
	configBlock := strings.TrimSpace(normalized[startIndex:configEndIndex])
	logicBlock := strings.TrimSpace(normalized[configEndIndex:])
	return configBlock, logicBlock
}

func insertConfigAfterHeader(content string, configBlock string) string {
	if strings.TrimSpace(configBlock) == "" {
		return content
	}

	headerEnd := strings.Index(content, "*/")
	if headerEnd < 0 {
		return configBlock + "\n\n" + content
	}

	insertAt := headerEnd + 2
	for insertAt < len(content) && (content[insertAt] == '\r' || content[insertAt] == '\n') {
		insertAt++
	}
	return content[:insertAt] + "\n" + configBlock + "\n\n" + content[insertAt:]
}

func buildBridgeFunction() string {
	return "// ============ Wireguard_Easytier Bridge ============\n" +
		"function main(config) {\n" +
		"    return _easytierEnhance(_originalMain(config));\n" +
		"}\n"
}
