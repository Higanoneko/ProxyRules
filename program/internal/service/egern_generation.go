package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/luluape/ProxyRules/internal/domain"
)

func (s *BuildService) validateEgernTarget(
	targets map[domain.Target]bool,
	outputRoot string,
) error {
	if !targets[domain.TargetEgern] {
		return nil
	}

	if len(targets) != 1 {
		return fmt.Errorf("Egern local generation must run separately")
	}

	if os.Getenv("EGERN_SUBSCRIPTION_URL") == "" {
		return fmt.Errorf("EGERN_SUBSCRIPTION_URL is required")
	}

	if outputRoot == "" {
		return fmt.Errorf("Egern requires explicit --output outside repository")
	}

	absOutput, err := filepath.Abs(outputRoot)
	if err != nil {
		return err
	}

	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return err
	}

	relative, err := filepath.Rel(absRoot, absOutput)
	if err != nil {
		return err
	}

	if relative == "." ||
		(relative != ".." &&
			!strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return fmt.Errorf("Egern output must be outside public repository")
	}

	return nil
}

func (s *BuildService) generateEgern(
	outputRoot string,
	ipv6True domain.PolicyPlan,
	ipv6False domain.PolicyPlan,
	generatedAt time.Time,
) error {
	url := strings.TrimSpace(os.Getenv("EGERN_SUBSCRIPTION_URL"))
	if url == "" {
		return fmt.Errorf("missing Egern subscription URL")
	}

	outputDir := filepath.Join(outputRoot, "Egern")
	subscriptions := []string{url}

	files := []struct {
		name string
		plan domain.PolicyPlan
	}{
		{"Egern_config.yaml", ipv6True},
		{"Egern_config_no_ipv6.yaml", ipv6False},
	}

	// Render both configurations before writing either one.
	rendered := make([]string, len(files))
	for i, file := range files {
		content, err := s.egern.RenderFull(file.plan, subscriptions)
		if err != nil {
			return fmt.Errorf("render %s: %w", file.name, err)
		}
		rendered[i] = content
	}

	for i, file := range files {
		if err := writeGeneratedFile(
			filepath.Join(outputDir, file.name),
			rendered[i],
			generatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}
