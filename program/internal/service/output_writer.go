package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const generatedAtLabel = "Generated at (UTC): "

func writeFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeGeneratedFile(path string, content string, generatedAt time.Time) error {
	return writeFile(path, prependGeneratedHeader(path, content, generatedAt))
}

func writeRenderResult(path string, generatedAt time.Time, renderer func() (string, error)) error {
	content, err := renderer()
	if err != nil {
		return err
	}
	return writeGeneratedFile(path, content, generatedAt)
}

func prependGeneratedHeader(path string, content string, generatedAt time.Time) string {
	headerLine := generatedHeaderLine(path, generatedAt)
	if headerLine == "" {
		return content
	}
	if filepath.Ext(path) == ".sgmodule" {
		return prependGeneratedHeaderAfterMetadata(content, headerLine)
	}

	if strings.HasSuffix(content, "\n") {
		return headerLine + "\n\n" + content
	}
	return headerLine + "\n\n" + content + "\n"
}

func prependGeneratedHeaderAfterMetadata(content string, headerLine string) string {
	lines := strings.SplitAfter(content, "\n")
	index := 0
	for index < len(lines) && strings.HasPrefix(strings.TrimRight(lines[index], "\r\n"), "#!") {
		index++
	}
	if index == 0 {
		if strings.HasSuffix(content, "\n") {
			return headerLine + "\n\n" + content
		}
		return headerLine + "\n\n" + content + "\n"
	}

	prefix := strings.Join(lines[:index], "")
	suffix := strings.Join(lines[index:], "")
	if suffix == "" {
		return prefix + headerLine + "\n"
	}
	return prefix + headerLine + "\n\n" + suffix
}

func stripGeneratedHeader(path string, content string) string {
	headerPrefix := generatedHeaderPrefix(path)
	if headerPrefix == "" {
		return content
	}

	if filepath.Ext(path) == ".sgmodule" {
		return stripGeneratedHeaderAfterMetadata(content, headerPrefix)
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], headerPrefix+" "+generatedAtLabel) {
		return content
	}

	remaining := lines[1:]
	for len(remaining) > 0 && strings.TrimSpace(remaining[0]) == "" {
		remaining = remaining[1:]
	}
	return strings.Join(remaining, "\n")
}

func stripGeneratedHeaderAfterMetadata(content string, headerPrefix string) string {
	lines := strings.Split(content, "\n")
	index := 0
	for index < len(lines) && strings.HasPrefix(lines[index], "#!") {
		index++
	}
	if index >= len(lines) || !strings.HasPrefix(lines[index], headerPrefix+" "+generatedAtLabel) {
		return content
	}

	remaining := append([]string{}, lines[:index]...)
	tail := lines[index+1:]
	for len(tail) > 0 && strings.TrimSpace(tail[0]) == "" {
		tail = tail[1:]
	}
	remaining = append(remaining, tail...)
	return strings.Join(remaining, "\n")
}

func generatedHeaderLine(path string, generatedAt time.Time) string {
	prefix := generatedHeaderPrefix(path)
	if prefix == "" {
		return ""
	}
	return fmt.Sprintf("%s %s%s", prefix, generatedAtLabel, generatedAt.UTC().Format(time.RFC3339))
}

func generatedHeaderPrefix(path string) string {
	switch filepath.Ext(path) {
	case ".js":
		return "//"
	case ".yaml", ".yml", ".conf", ".lcf", ".stoverride", ".sgmodule":
		return "#"
	default:
		return ""
	}
}
