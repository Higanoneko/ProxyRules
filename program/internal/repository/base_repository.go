package repository

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PianCat/ProxyRules/internal/domain"
	"gopkg.in/yaml.v3"
)

const (
	defaultInternetTestURL = "http://connect.rom.miui.com/generate_204"
	defaultProxyTestURL    = "http://www.gstatic.com/generate_204"
)

type BaseRepository struct {
	root string
}

type CategoryConfig struct {
	URL string `yaml:"url"`
}

type BaseData struct {
	Root              string
	DNS               domain.DNSSource
	FakeIPFilter      []string
	SurgeAlwaysRealIP []string
	Ports             domain.Ports
	TestURLs          domain.TestURLs
	Heads             map[string]string
	RawRules          domain.RuleSections
	Categories        map[string]CategoryConfig
	ToolMappings      map[string]map[string]string
	FileTypeMappings  map[string]map[string]string
}

type dnsConfig struct {
	DNS struct {
		IP             []string `yaml:"IP"`
		DefaultDoH     []string `yaml:"Default_DoH"`
		ProxyServerDoH []string `yaml:"Proxy_Server_DoH"`
		DirectDoH      []string `yaml:"Direct_DoH"`
		FallbackDoH    []string `yaml:"Fallback_DoH"`
	} `yaml:"DNS"`
	LegacyDNSIP        []string `yaml:"DNS_IP"`
	LegacyDNSDoH       []string `yaml:"DNS_DoH"`
	LegacyDNSDoHDirect []string `yaml:"DNS_DoH_Direct"`
}

type portsConfig struct {
	Mihomo struct {
		MixedPort int `yaml:"Mixed_Port"`
	} `yaml:"Mihomo"`
	General struct {
		HTTPPort   int `yaml:"HTTP_Port"`
		SOCKS5Port int `yaml:"SOCKS5_Port"`
	} `yaml:"General"`
}

type fakeIPConfig struct {
	FakeIPFilter      []string `yaml:"Fake_IP_Filter"`
	SurgeAlwaysRealIP []string `yaml:"Surge_Always_Real_IP"`
}

type linkBaseConfig struct {
	Categories            map[string]CategoryConfig    `yaml:"Categories"`
	CategoriesToolsList   map[string]map[string]string `yaml:"Categories_Tools_List"`
	CategoriesFiletypeMap map[string]map[string]string `yaml:"Categories_Filetype_List"`
}

func NewBaseRepository(root string) *BaseRepository {
	return &BaseRepository{root: root}
}

func (r *BaseRepository) Load() (BaseData, error) {
	baseDir := filepath.Join(r.root, "Base")

	var dns dnsConfig
	if err := readYAML(filepath.Join(baseDir, "DNS.yaml"), &dns); err != nil {
		return BaseData{}, err
	}

	var ports portsConfig
	if err := readYAML(filepath.Join(baseDir, "Ports.yaml"), &ports); err != nil {
		return BaseData{}, err
	}

	var fakeIP fakeIPConfig
	if err := readYAML(filepath.Join(baseDir, "Fake_IP_Filter.yaml"), &fakeIP); err != nil {
		return BaseData{}, err
	}

	rules, err := readRuleSections(filepath.Join(baseDir, "Rules", "RemoteRules.yaml"))

	var linkBase linkBaseConfig
	if err := readYAML(filepath.Join(baseDir, "Rules", "RemoteRulesLinkBase.yaml"), &linkBase); err != nil {
		return BaseData{}, err
	}

	testURLs, err := readTestURLs(filepath.Join(baseDir, "Test_URL.yaml"))
	if err != nil {
		return BaseData{}, err
	}

	heads := map[string]string{}
	for key, filename := range map[string]string{
		"mihomo":   "Head_Mihomo.yaml",
		"box4root": "Head_Box4Root.yaml",
		"stash":    "Head_Stash.yaml",
		"loon":     "Head_Loon.conf",
		"surge":    "Head_Surge.conf",
	} {
		content, err := os.ReadFile(filepath.Join(baseDir, "Head", filename))
		if err != nil {
			return BaseData{}, fmt.Errorf("read head %s: %w", filename, err)
		}
		heads[key] = string(content)
	}

	return BaseData{
		Root:              r.root,
		DNS:               dns.source(),
		FakeIPFilter:      cloneStrings(fakeIP.FakeIPFilter),
		SurgeAlwaysRealIP: cloneStrings(fakeIP.SurgeAlwaysRealIP),
		Ports: domain.Ports{
			Mixed:  ports.Mihomo.MixedPort,
			HTTP:   ports.General.HTTPPort,
			Socks5: ports.General.SOCKS5Port,
		},
		TestURLs:         testURLs,
		Heads:            heads,
		RawRules:         rules,
		Categories:       linkBase.Categories,
		ToolMappings:     linkBase.CategoriesToolsList,
		FileTypeMappings: linkBase.CategoriesFiletypeMap,
	}, nil
}

func (d BaseData) Head(key string) (string, error) {
	content, ok := d.Heads[key]
	if !ok || strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("missing head template: %s", key)
	}
	return content, nil
}

func DefaultGeoXURLs() map[string]string {
	return map[string]string{
		"geoip":   "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip-lite.dat",
		"geosite": "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat",
		"mmdb":    "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb",
		"asn":     "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/GeoLite2-ASN.mmdb",
	}
}

func readRuleSections(path string) (domain.RuleSections, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.RuleSections{}, fmt.Errorf("read %s: %w", path, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return domain.RuleSections{}, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return domain.RuleSections{}, fmt.Errorf("empty document: %s", path)
	}

	topLevel := root.Content[0]
	if topLevel.Kind != yaml.MappingNode {
		return domain.RuleSections{}, fmt.Errorf("expected mapping at top level: %s", path)
	}

	baseRules, err := extractOrderedSection(topLevel, "BaseRules")
	if err != nil {
		return domain.RuleSections{}, fmt.Errorf("BaseRules: %w", err)
	}
	customRules, err := extractOrderedSection(topLevel, "CustomRules")
	if err != nil {
		return domain.RuleSections{}, fmt.Errorf("CustomRules: %w", err)
	}

	return domain.RuleSections{
		BaseRules:   baseRules,
		CustomRules: customRules,
	}, nil
}

func extractOrderedSection(parent *yaml.Node, key string) (domain.OrderedSection, error) {
	sectionNode := findMappingValue(parent, key)
	if sectionNode == nil {
		return domain.OrderedSection{}, fmt.Errorf("missing section: %s", key)
	}
	if sectionNode.Kind != yaml.MappingNode {
		return domain.OrderedSection{}, fmt.Errorf("expected mapping for %s", key)
	}

	keys := make([]string, 0, len(sectionNode.Content)/2)
	entries := make(map[string]map[string]any, len(sectionNode.Content)/2)

	for i := 0; i < len(sectionNode.Content); i += 2 {
		entryKey := sectionNode.Content[i].Value
		entryValue := sectionNode.Content[i+1]

		if entryValue.Kind != yaml.MappingNode {
			continue
		}

		entryMap := make(map[string]any)
		if err := entryValue.Decode(&entryMap); err != nil {
			continue
		}

		if _, ok := entryMap["category"]; !ok {
			continue
		}

		keys = append(keys, entryKey)
		entries[entryKey] = entryMap
	}

	return domain.OrderedSection{Keys: keys, Entries: entries}, nil
}

func findMappingValue(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1]
		}
	}
	return nil
}

func readYAML(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(content, target); err != nil {
		return fmt.Errorf("unmarshal %s: %w", path, err)
	}
	return nil
}

func readTestURLs(path string) (domain.TestURLs, error) {
	file, err := os.Open(path)
	if err != nil {
		return domain.TestURLs{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	testURLs := domain.TestURLs{
		Internet: defaultInternetTestURL,
		Proxy:    defaultProxyTestURL,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		switch strings.TrimSpace(key) {
		case "internet-test-url":
			testURLs.Internet = strings.TrimSpace(value)
		case "proxy-test-url":
			testURLs.Proxy = strings.TrimSpace(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.TestURLs{}, fmt.Errorf("scan %s: %w", path, err)
	}

	return testURLs, nil
}

func cloneStrings(values []string) []string {
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func (c dnsConfig) source() domain.DNSSource {
	return domain.DNSSource{
		BootstrapResolvers: cloneStrings(firstConfiguredGroup(c.DNS.IP, c.LegacyDNSIP)),
		Upstreams: domain.DNSUpstreamPolicy{
			Default:     cloneStrings(firstConfiguredGroup(c.DNS.DefaultDoH, c.LegacyDNSDoH)),
			ProxyServer: cloneStrings(firstConfiguredGroup(c.DNS.ProxyServerDoH)),
			Direct:      cloneStrings(firstConfiguredGroup(c.DNS.DirectDoH, c.LegacyDNSDoHDirect)),
			Fallback:    cloneStrings(firstConfiguredGroup(c.DNS.FallbackDoH)),
		},
	}
}

func firstConfiguredGroup(groups ...[]string) []string {
	for _, group := range groups {
		if len(group) > 0 {
			return group
		}
	}
	return nil
}
