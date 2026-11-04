package main

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ruleFile struct {
	Payload []string `yaml:"payload"`
}

type source struct {
	Name     string
	Behavior string
}

func main() {
	input := "/tmp/proxyrules-loyalsoldier"
	output := "../Rules"

	sources := []source{
		{"reject", "domain"},
		{"icloud", "domain"},
		{"apple", "domain"},
		{"google", "domain"},
		{"proxy", "domain"},
		{"direct", "domain"},
		{"private", "domain"},
		{"gfw", "domain"},
		{"tld-not-cn", "domain"},
		{"telegramcidr", "ipcidr"},
		{"cncidr", "ipcidr"},
		{"lancidr", "ipcidr"},
		{"applications", "classical"},
	}

	for _, src := range sources {
		data, err := os.ReadFile(
			filepath.Join(input, src.Name+".txt"),
		)
		check(err)

		var document ruleFile
		check(yaml.Unmarshal(data, &document))

		if len(document.Payload) == 0 {
			check(fmt.Errorf("%s: empty payload", src.Name))
		}

		seen := make(map[string]bool)
		var mihomo []string
		var classical []string

		for _, raw := range document.Payload {
			value := strings.TrimSpace(raw)

			if value == "" {
				check(fmt.Errorf("%s: empty rule", src.Name))
			}
			if seen[value] {
				continue
			}
			seen[value] = true

			var converted string

			switch src.Behavior {
			case "domain":
				switch {
				case strings.HasPrefix(value, "+."):
					domain := strings.TrimPrefix(value, "+.")
					check(validateDomain(domain))
					converted = "DOMAIN-SUFFIX," + domain
				case strings.ContainsAny(value, "*?"):
					check(fmt.Errorf(
						"%s: unsupported wildcard %q",
						src.Name, value,
					))
				default:
					check(validateDomain(value))
					converted = "DOMAIN," + value
				}
			case "classical":
				parts := strings.SplitN(value, ",", 2)
				if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
					check(fmt.Errorf("%s: invalid classical rule %q", src.Name, value))
				}
				switch parts[0] {
				case "PROCESS-NAME", "PROCESS-PATH",
					"DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD",
					"IP-CIDR", "IP-CIDR6", "GEOIP",
					"DST-PORT", "SRC-PORT", "NETWORK":
					converted = value
				default:
					check(fmt.Errorf("%s: unsupported classical rule %q", src.Name, value))
				}
			case "ipcidr":
				prefix, err := netip.ParsePrefix(value)
				check(err)

				if prefix.Addr().Is4() {
					converted = "IP-CIDR," + value
				} else {
					converted = "IP-CIDR6," + value
				}
			}

			mihomo = append(mihomo, value)
			classical = append(classical, converted)
		}

		// Mihomo / Stash: YAML payload
		yamlData, err := yaml.Marshal(ruleFile{
			Payload: mihomo,
		})
		check(err)

		write(
			filepath.Join(output, "Mihomo", src.Name+".yaml"),
			yamlData,
		)

		// Loon / Surge: classical rule lines
		content := strings.Join(classical, "\n") + "\n"

		for _, target := range []string{"Loon", "Surge"} {
			write(
				filepath.Join(output, target, src.Name+".list"),
				[]byte(content),
			)
		}

		fmt.Printf(
			"%s: %d rules converted\n",
			src.Name, len(mihomo),
		)
	}
}

func validateDomain(value string) error {
	if value == "" ||
		strings.ContainsAny(value, " \t\r\n,/*?") ||
		strings.HasPrefix(value, ".") ||
		strings.HasSuffix(value, ".") {
		return fmt.Errorf("invalid domain: %q", value)
	}
	return nil
}

func write(path string, content []byte) {
	check(os.MkdirAll(filepath.Dir(path), 0755))
	check(os.WriteFile(path, content, 0644))
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
