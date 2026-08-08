package service

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Higanoneko/ProxyRules/internal/domain"
	"github.com/Higanoneko/ProxyRules/internal/render"
	"github.com/Higanoneko/ProxyRules/internal/repository"
)

type BuildService struct {
	root         string
	base         repository.BaseData
	planner      *PolicyPlanBuilder
	mihomo       *render.MihomoRenderer
	mihomo4root  *render.Mihomo4RootRenderer
	mihomoScript *render.MihomoScriptRenderer
	stash        *render.StashRenderer
	loon         *render.LoonRenderer
	surge        *render.SurgeRenderer
}

func NewBuildService(root string) (*BuildService, error) {
	base, err := repository.NewBaseRepository(root).Load()
	if err != nil {
		return nil, err
	}

	return &BuildService{
		root:         root,
		base:         base,
		planner:      NewPolicyPlanBuilder(base),
		mihomo:       render.NewMihomoRenderer(base),
		mihomo4root:  render.NewMihomo4RootRenderer(base),
		mihomoScript: render.NewMihomoScriptRenderer(base),
		stash:        render.NewStashRenderer(base),
		loon:         render.NewLoonRenderer(base),
		surge:        render.NewSurgeRenderer(base),
	}, nil
}

func (s *BuildService) Generate(targets []domain.Target, outputRoot string, nodeNames []string) error {
	selectedTargets := normalizeTargets(targets)
	if outputRoot == "" {
		outputRoot = filepath.Join(s.root, "Config")
	}
	generatedAt := time.Now().UTC()
	if err := copyEasytierSurgeModule(easytierSurgeModuleSourcePath(s.root), outputRoot, generatedAt); err != nil {
		return err
	}

	requiresPolicyPlan := selectedTargets[domain.TargetMihomo] ||
		selectedTargets[domain.TargetStash] ||
		selectedTargets[domain.TargetLoon] ||
		selectedTargets[domain.TargetSurge]

	var planIPv6True domain.PolicyPlan
	var planIPv6False domain.PolicyPlan
	if requiresPolicyPlan {
		var err error
		planIPv6True, err = s.planner.Build(true, nodeNames)
		if err != nil {
			return err
		}
		planIPv6False, err = s.planner.Build(false, nodeNames)
		if err != nil {
			return err
		}
	}

	if selectedTargets[domain.TargetMihomo] {
		if err := s.generateMihomo(outputRoot, planIPv6True, planIPv6False, generatedAt); err != nil {
			return err
		}
	}
	if selectedTargets[domain.TargetStash] {
		if err := s.generateStash(outputRoot, planIPv6True, planIPv6False, generatedAt); err != nil {
			return err
		}
	}
	if selectedTargets[domain.TargetLoon] {
		if err := s.generateLoon(outputRoot, planIPv6True, planIPv6False, generatedAt); err != nil {
			return err
		}
	}
	if selectedTargets[domain.TargetSurge] {
		if err := s.generateSurge(outputRoot, planIPv6True, planIPv6False, generatedAt); err != nil {
			return err
		}
	}
	if selectedTargets[domain.TargetEasytier] {
		if err := generateEasytierBundle(
			filepath.Join(outputRoot, "Mihomo"),
			easytierJSSourcePath(s.root),
			resolveEasytierOutputDir(outputRoot),
			generatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func normalizeTargets(targets []domain.Target) map[domain.Target]bool {
	result := map[domain.Target]bool{}
	if len(targets) == 0 {
		result[domain.TargetMihomo] = true
		result[domain.TargetStash] = true
		result[domain.TargetLoon] = true
		result[domain.TargetSurge] = true
		result[domain.TargetEasytier] = true
		return result
	}

	for _, target := range targets {
		if target == domain.TargetAll {
			return normalizeTargets(nil)
		}
		result[target] = true
	}
	return result
}

func (s *BuildService) generateMihomo(outputRoot string, planIPv6True domain.PolicyPlan, planIPv6False domain.PolicyPlan, generatedAt time.Time) error {
	mihomoOutputDir := filepath.Join(outputRoot, "Mihomo")
	mihomo4RootOutputDir := filepath.Join(outputRoot, "Mihomo4Root")
	if err := removeFiles(
		filepath.Join(mihomoOutputDir, "Box4Root_mihomo_config.yaml"),
		filepath.Join(mihomoOutputDir, "Box4Root_mihomo_config_no_ipv6.yaml"),
		filepath.Join(mihomoOutputDir, "Box4Root_mihomo_config_tun.yaml"),
		filepath.Join(mihomoOutputDir, "Box4Root_mihomo_config_tun_no_ipv6.yaml"),
		filepath.Join(outputRoot, "Box4Root", "Box4Root_mihomo_config.yaml"),
		filepath.Join(outputRoot, "Box4Root", "Box4Root_mihomo_config_no_ipv6.yaml"),
		filepath.Join(outputRoot, "Box4Root", "Box4Root_mihomo_config_tun.yaml"),
		filepath.Join(outputRoot, "Box4Root", "Box4Root_mihomo_config_tun_no_ipv6.yaml"),
		filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config.yaml"),
		filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_no_ipv6.yaml"),
		filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_tun.yaml"),
		filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_tun_no_ipv6.yaml"),
	); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_config_ipv6-1_full-0.yaml"), generatedAt, func() (string, error) {
		return s.mihomo.RenderStandard(planIPv6True, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_config_ipv6-1_full-1.yaml"), generatedAt, func() (string, error) {
		return s.mihomo.RenderStandard(planIPv6True, true)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_config_ipv6-0_full-0.yaml"), generatedAt, func() (string, error) {
		return s.mihomo.RenderStandard(planIPv6False, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_config_ipv6-0_full-1.yaml"), generatedAt, func() (string, error) {
		return s.mihomo.RenderStandard(planIPv6False, true)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_tun.yaml"), generatedAt, func() (string, error) {
		return s.mihomo4root.RenderMihomo4Root(planIPv6True, true)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_tun_no_ipv6.yaml"), generatedAt, func() (string, error) {
		return s.mihomo4root.RenderMihomo4Root(planIPv6False, true)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config.yaml"), generatedAt, func() (string, error) {
		return s.mihomo4root.RenderMihomo4Root(planIPv6True, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomo4RootOutputDir, "Mihomo4Root_mihomo_config_no_ipv6.yaml"), generatedAt, func() (string, error) {
		return s.mihomo4root.RenderMihomo4Root(planIPv6False, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_convert_args.js"), generatedAt, func() (string, error) {
		return s.mihomoScript.RenderArgs(planIPv6True)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_convert_ipv6-1_full-0.js"), generatedAt, func() (string, error) {
		return s.mihomoScript.RenderFixed(planIPv6True, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_convert_ipv6-1_full-1.js"), generatedAt, func() (string, error) {
		return s.mihomoScript.RenderFixed(planIPv6True, true)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_convert_ipv6-0_full-0.js"), generatedAt, func() (string, error) {
		return s.mihomoScript.RenderFixed(planIPv6False, false)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(mihomoOutputDir, "mihomo_convert_ipv6-0_full-1.js"), generatedAt, func() (string, error) {
		return s.mihomoScript.RenderFixed(planIPv6False, true)
	}); err != nil {
		return err
	}

	return nil
}

func removeFiles(paths ...string) error {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (s *BuildService) generateStash(outputRoot string, planIPv6True domain.PolicyPlan, planIPv6False domain.PolicyPlan, generatedAt time.Time) error {
	outputDir := filepath.Join(outputRoot, "Stash")
	if err := writeRenderResult(filepath.Join(outputDir, "Stash_config_full.yaml"), generatedAt, func() (string, error) {
		return s.stash.RenderFull(planIPv6True)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(outputDir, "Stash_config_full_no_ipv6.yaml"), generatedAt, func() (string, error) {
		return s.stash.RenderFull(planIPv6False)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(outputDir, "Stash_override.stoverride"), generatedAt, func() (string, error) {
		return s.stash.RenderOverride(planIPv6True)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(outputDir, "Stash_override_no_ipv6.stoverride"), generatedAt, func() (string, error) {
		return s.stash.RenderOverride(planIPv6False)
	}); err != nil {
		return err
	}
	return nil
}

func (s *BuildService) generateLoon(outputRoot string, planIPv6True domain.PolicyPlan, planIPv6False domain.PolicyPlan, generatedAt time.Time) error {
	outputDir := filepath.Join(outputRoot, "Loon")
	if err := writeRenderResult(filepath.Join(outputDir, "Loon_config.lcf"), generatedAt, func() (string, error) {
		return s.loon.Render(planIPv6True)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(outputDir, "Loon_config_no_ipv6.lcf"), generatedAt, func() (string, error) {
		return s.loon.Render(planIPv6False)
	}); err != nil {
		return err
	}
	return nil
}

func (s *BuildService) generateSurge(outputRoot string, planIPv6True domain.PolicyPlan, planIPv6False domain.PolicyPlan, generatedAt time.Time) error {
	outputDir := filepath.Join(outputRoot, "Surge")
	if err := writeRenderResult(filepath.Join(outputDir, "Surge_config.conf"), generatedAt, func() (string, error) {
		return s.surge.Render(planIPv6True)
	}); err != nil {
		return err
	}
	if err := writeRenderResult(filepath.Join(outputDir, "Surge_config_no_ipv6.conf"), generatedAt, func() (string, error) {
		return s.surge.Render(planIPv6False)
	}); err != nil {
		return err
	}
	return nil
}

func CreateTestNodes() []string {
	return []string{
		"香港 IEPL 01", "香港 IEPL 02", "香港 IEPL 03",
		"台湾 HiNet 01", "台湾 HiNet 02", "台湾 HiNet 03",
		"美国 洛杉矶 01", "美国 洛杉矶 02", "美国 洛杉矶 03",
		"日本 东京 01", "日本 东京 02", "日本 东京 03",
		"新加坡 01", "新加坡 02", "新加坡 03",
		"韩国 首尔 01",
	}
}
