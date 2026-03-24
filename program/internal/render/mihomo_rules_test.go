package render

import "testing"

func TestResolveMihomoRulesReplacesPackInPlace(t *testing.T) {
	head := `
rules:
  - DST-PORT,53,DNS_Hijack
  - <ProxyRules_Pack>
`
	generated := []string{"RULE-SET,AI,AI", "RULE-SET,Telegram,Telegram"}

	resolved, err := resolveMihomoRules(head, nil, generated)
	if err != nil {
		t.Fatalf("resolve mihomo rules: %v", err)
	}

	if len(resolved) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(resolved))
	}
	if resolved[0] != "DST-PORT,53,DNS_Hijack" {
		t.Fatalf("unexpected first rule: %s", resolved[0])
	}
	if resolved[1] != "RULE-SET,AI,AI" || resolved[2] != "RULE-SET,Telegram,Telegram" {
		t.Fatalf("unexpected generated rules order: %#v", resolved)
	}
}

func TestResolveMihomoRulesSupportsPackBeforeCustomRules(t *testing.T) {
	head := `
rules:
  - <ProxyRules_Pack>
  - DST-PORT,53,DNS_Hijack
`
	generated := []string{"RULE-SET,AI,AI", "RULE-SET,Telegram,Telegram"}

	resolved, err := resolveMihomoRules(head, nil, generated)
	if err != nil {
		t.Fatalf("resolve mihomo rules: %v", err)
	}

	if len(resolved) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(resolved))
	}
	if resolved[0] != "RULE-SET,AI,AI" || resolved[1] != "RULE-SET,Telegram,Telegram" {
		t.Fatalf("unexpected generated rules order: %#v", resolved)
	}
	if resolved[2] != "DST-PORT,53,DNS_Hijack" {
		t.Fatalf("unexpected trailing custom rule: %s", resolved[2])
	}
}
