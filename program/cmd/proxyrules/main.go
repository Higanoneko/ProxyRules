package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PianCat/ProxyRules/internal/domain"
	"github.com/PianCat/ProxyRules/internal/projectroot"
	"github.com/PianCat/ProxyRules/internal/service"
)

type multiTargetFlag []string

func (m *multiTargetFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiTargetFlag) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		*m = append(*m, trimmed)
	}
	return nil
}

func main() {
	var toolFlags multiTargetFlag
	var outputPath string
	var testMode bool

	flag.Var(&toolFlags, "tool", "目标工具，可重复传入或使用逗号分隔：mihomo,stash,loon,surge,easytier,all")
	flag.StringVar(&outputPath, "output", "", "输出目录，默认为项目根目录下的 Config")
	flag.BoolVar(&testMode, "test", false, "使用测试节点生成配置")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		exitf("get cwd: %v", err)
	}
	projectRoot, err := projectroot.Find(cwd)
	if err != nil {
		exitf("find project root: %v", err)
	}

	targets, err := parseTargets(toolFlags)
	if err != nil {
		exitf("%v", err)
	}

	buildService, err := service.NewBuildService(projectRoot)
	if err != nil {
		exitf("load project: %v", err)
	}

	var nodeNames []string
	if testMode {
		nodeNames = service.CreateTestNodes()
		fmt.Println("[Test Mode] Using mock nodes")
	}

	if outputPath != "" && !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(projectRoot, outputPath)
	}

	if err := buildService.Generate(targets, outputPath, nodeNames); err != nil {
		exitf("generate: %v", err)
	}

	fmt.Println("Configuration files generated successfully.")
	if outputPath == "" {
		outputPath = filepath.Join(projectRoot, "Config")
	}
	fmt.Printf("Output directory: %s\n", outputPath)
}

func parseTargets(values []string) ([]domain.Target, error) {
	if len(values) == 0 {
		return []domain.Target{domain.TargetAll}, nil
	}

	targets := make([]domain.Target, 0, len(values))
	for _, value := range values {
		switch strings.ToLower(value) {
		case "all":
			return []domain.Target{domain.TargetAll}, nil
		case "mihomo":
			targets = append(targets, domain.TargetMihomo)
		case "stash":
			targets = append(targets, domain.TargetStash)
		case "loon":
			targets = append(targets, domain.TargetLoon)
		case "surge":
			targets = append(targets, domain.TargetSurge)
		case "easytier":
			targets = append(targets, domain.TargetEasytier)
		default:
			return nil, fmt.Errorf("unsupported tool: %s", value)
		}
	}
	return targets, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
