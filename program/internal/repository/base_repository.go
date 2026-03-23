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
	DNSIP             []string
	DNSDoH            []string
	FakeIPFilter      []string
	SurgeAlwaysRealIP []string
	Ports             domain.Ports
	TestURLs          domain.TestURLs
	Heads             map[string]string
	RawRules          map[string]any
	Categories        map[string]CategoryConfig
	ToolMappings      map[string]map[string]string
	FileTypeMappings  map[string]map[string]string
}

type dnsConfig struct {
	DNSIP  []string `yaml:"DNS_IP"`
	DNSDoH []string `yaml:"DNS_DoH"`
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

type rulesConfig struct {
	Rules map[string]any `yaml:"rules"`
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

	var rules rulesConfig
	if err := readYAML(filepath.Join(baseDir, "Rules", "RemoteRules.yaml"), &rules); err != nil {
		return BaseData{}, err
	}

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
		"mihomo":     "Head_Mihomo.yaml",
		"mihomo_tun": "Head_Mihomo_Tun.yaml",
		"stash":      "Head_Stash.yaml",
		"loon":       "Head_Loon.conf",
		"surge":      "Head_Surge.conf",
	} {
		content, err := os.ReadFile(filepath.Join(baseDir, "Head", filename))
		if err != nil {
			return BaseData{}, fmt.Errorf("read head %s: %w", filename, err)
		}
		heads[key] = string(content)
	}

	return BaseData{
		Root:              r.root,
		DNSIP:             cloneStrings(dns.DNSIP),
		DNSDoH:            cloneStrings(dns.DNSDoH),
		FakeIPFilter:      cloneStrings(fakeIP.FakeIPFilter),
		SurgeAlwaysRealIP: cloneStrings(fakeIP.SurgeAlwaysRealIP),
		Ports: domain.Ports{
			Mixed:  ports.Mihomo.MixedPort,
			HTTP:   ports.General.HTTPPort,
			Socks5: ports.General.SOCKS5Port,
		},
		TestURLs:         testURLs,
		Heads:            heads,
		RawRules:         rules.Rules,
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
