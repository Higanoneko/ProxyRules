#!/usr/bin/env python3
import argparse
from pathlib import Path
from urllib.parse import urlparse
import yaml

parser = argparse.ArgumentParser()
parser.add_argument("--config-dir", type=Path, required=True)
args = parser.parse_args()

root = Path(__file__).resolve().parents[1]
rules_dir = root / "Rules/Egern"
allowed = {
    "no_resolve", "domain_set", "domain_suffix_set",
    "domain_keyword_set", "domain_regex_set",
    "domain_wildcard_set", "ip_cidr_set", "ip_cidr6_set",
    "user_agent_set", "url_regex_set", "geoip_set",
    "asn_set", "dest_port_set", "ssid_set", "bssid_set",
    "cellular_set", "protocol_set", "and_set", "or_set",
    "not_set",
}

for filename, expected_ipv6 in (
    ("Egern_config.yaml", True),
    ("Egern_config_no_ipv6.yaml", False),
):
    path = args.config_dir / filename
    doc = yaml.safe_load(path.read_text(encoding="utf-8"))
    groups = doc["policy_groups"]
    rules = doc["rules"]
    refs = [item["rule_set"] for item in rules if "rule_set" in item]

    assert doc["ipv6"] is expected_ipv6, filename
    assert len(groups) == 29, (filename, len(groups))
    assert len(refs) == 30, (filename, len(refs))
    assert len(rules) == 31, (filename, len(rules))
    assert rules[-1]["default"]["policy"] == "选择代理"
    assert doc["dns"]["bootstrap"] == ["system"]

    seen = set()
    for ref in refs:
        url = urlparse(ref["match"])
        assert url.scheme == "https", ref
        assert url.hostname == "raw.githubusercontent.com", ref
        assert url.path.startswith(
            "/luluape/ProxyRules/main/Rules/Egern/"
        ), ref
        name = Path(url.path).name
        assert name not in seen, f"Duplicate rule: {name}"
        seen.add(name)

        rule_path = rules_dir / name
        assert rule_path.is_file(), f"Missing: {rule_path}"
        data = yaml.safe_load(rule_path.read_text(encoding="utf-8"))
        assert isinstance(data, dict), rule_path
        assert not set(data).difference(allowed), rule_path
        assert any(isinstance(v, list) and v for v in data.values())

    print(f"PASS {filename}: 29 groups, 30 linked rule files")

print("PASS: Egern native static validation completed")
