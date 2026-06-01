package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Higanoneko/ProxyRules/internal/postprocess"
)

func easytierJSSourcePath(projectRoot string) string {
	return filepath.Join(projectRoot, "Base", "Wireguard", "Easytier", "Easytier.js")
}

func easytierSurgeModuleSourcePath(projectRoot string) string {
	return filepath.Join(projectRoot, "Base", "Wireguard", "Easytier", "Easytier.sgmodule")
}

func generateEasytierBundle(mihomoOutputDir string, easytierSourcePath string, outputDir string, generatedAt time.Time) error {
	if _, err := os.Stat(mihomoOutputDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mihomo source directory not found: %s; generate Mihomo configs first", mihomoOutputDir)
		}
		return err
	}

	easytierSource, err := os.ReadFile(easytierSourcePath)
	if err != nil {
		return fmt.Errorf("read easytier source: %w", err)
	}

	entries, err := os.ReadDir(mihomoOutputDir)
	if err != nil {
		return err
	}

	generatedCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !isMihomoConvertScript(entry.Name()) {
			continue
		}

		content, err := os.ReadFile(filepath.Join(mihomoOutputDir, entry.Name()))
		if err != nil {
			return err
		}

		merged := postprocess.MergeEasytier(stripGeneratedHeader(entry.Name(), string(content)), string(easytierSource))
		if err := writeGeneratedFile(filepath.Join(outputDir, entry.Name()), merged, generatedAt); err != nil {
			return err
		}
		generatedCount++
	}

	if generatedCount == 0 {
		return fmt.Errorf("no mihomo_convert_*.js files found in %s; generate Mihomo configs first", mihomoOutputDir)
	}

	return nil
}

func copyFile(sourcePath string, destinationPath string, generatedAt time.Time) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return writeGeneratedFile(destinationPath, stripGeneratedHeader(destinationPath, string(content)), generatedAt)
}

func copyEasytierSurgeModule(sourcePath string, outputRoot string, generatedAt time.Time) error {
	destinationPath := filepath.Join(resolveWireguardOutputDir(outputRoot, "Surge"), "Easytier.sgmodule")
	if err := copyFile(sourcePath, destinationPath, generatedAt); err != nil {
		return fmt.Errorf("copy easytier surge module: %w", err)
	}
	return nil
}

func resolveEasytierOutputDir(outputRoot string) string {
	return resolveWireguardOutputDir(outputRoot, "Mihomo")
}

func resolveWireguardOutputDir(outputRoot string, subdir string) string {
	cleanRoot := filepath.Clean(outputRoot)
	if strings.EqualFold(filepath.Base(cleanRoot), "Config") {
		return filepath.Join(filepath.Dir(cleanRoot), "Wireguard_Easytier", subdir)
	}
	return filepath.Join(cleanRoot, "Wireguard_Easytier", subdir)
}

func isMihomoConvertScript(name string) bool {
	return strings.HasPrefix(name, "mihomo_convert_") && strings.HasSuffix(name, ".js")
}
