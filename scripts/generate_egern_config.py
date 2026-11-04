#!/usr/bin/env python3
"""Generate an Egern rules fragment from the Loyalsoldier catalog."""

import argparse
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / "Base/Rules/LoyalsoldierCatalog.yaml"
REMOTE = ROOT / "Base/Rules/RemoteRules.yaml"
RULES = ROOT / "Rules/Egern"
OUTPUT = ROOT / "Config/Egern/Egern_rules.yaml"

BASE_URL = (
    "https://raw.githubusercontent.com/luluape/ProxyRules/"
    "main/Rules/Egern"
)

POLICY_MAP = {
    "直接连接": "DIRECT",
    "DIRECT": "DIRECT",
    "REJECT": "REJECT",
}

SECTIONS = ("BaseRules", "CustomRules")
UPDATE_INTERVAL = 86400


def read_yaml(path):
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"Invalid YAML mapping: {path}")
    return data


def build_rules(catalog, remote):
    specifications = catalog["rules"]
    if not isinstance(specifications, dict):
        raise ValueError("Invalid catalog rules")

    by_id = {}
    for name, spec in specifications.items():
        rule_id = spec["rule_id"]
        if rule_id in by_id:
            raise ValueError(f"Duplicate rule ID: {rule_id}")
        by_id[rule_id] = (name, spec)

    result = []
    found = set()

    for section in SECTIONS:
        entries = remote.get(section)
        if not isinstance(entries, dict):
            raise ValueError(f"Missing section: {section}")

        for rule_id, entry in entries.items():
            if rule_id not in by_id:
                if (
                    isinstance(entry, dict)
                    and entry.get("category") == "Loyalsoldier"
                ):
                    raise ValueError(f"Unknown Loyalsoldier rule: {rule_id}")
                continue

            name, spec = by_id[rule_id]

            if not spec["enabled"]:
                raise ValueError(f"Disabled rule is active: {name}")

            if spec["section"] != section:
                raise ValueError(f"Wrong section: {name}")

            if entry["category"] != "Loyalsoldier":
                raise ValueError(f"Wrong category: {name}")

            if entry["remotefile"] != name:
                raise ValueError(f"Wrong remotefile: {name}")

            if entry["behavior"] != spec["behavior"]:
                raise ValueError(f"Wrong behavior: {name}")

            if entry["policyname"] != spec["policy"]:
                raise ValueError(f"Policy mismatch: {name}")

            if spec["behavior"] not in ("domain", "ipcidr"):
                raise ValueError(
                    f"Egern converter does not support: {name}"
                )

            if not (RULES / f"{name}.yaml").is_file():
                raise ValueError(f"Missing Egern rule set: {name}")

            policy = POLICY_MAP.get(spec["policy"], spec["policy"])

            result.append({
                "rule_set": {
                    "match": f"{BASE_URL}/{name}.yaml",
                    "policy": policy,
                    "update_interval": UPDATE_INTERVAL,
                }
            })
            found.add(rule_id)

    expected = {
        spec["rule_id"]
        for spec in specifications.values()
        if spec["enabled"]
    }

    if found != expected:
        raise ValueError(
            f"Enabled rules mismatch. Missing: {sorted(expected - found)}"
        )

    return result


def generate():
    catalog = read_yaml(CATALOG)
    remote = read_yaml(REMOTE)
    rules = build_rules(catalog, remote)

    content = yaml.safe_dump(
        {"rules": rules},
        allow_unicode=True,
        sort_keys=False,
        default_flow_style=False,
    )

    return rules, content


def main():
    parser = argparse.ArgumentParser()
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--write", action="store_true")
    mode.add_argument("--check", action="store_true")
    args = parser.parse_args()

    rules, content = generate()

    current = (
        OUTPUT.read_text(encoding="utf-8")
        if OUTPUT.is_file() else None
    )

    if args.check:
        if current != content:
            print("FAIL: Egern rules fragment is out of sync")
            return 1
        print(f"PASS: {len(rules)} Egern rules synchronized")
        return 0

    if current == content:
        print(f"PASS: already synchronized; {len(rules)} rules")
        return 0

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT.write_text(content, encoding="utf-8")
    print(f"PASS: generated {OUTPUT.relative_to(ROOT)}")
    print(f"Active rules: {len(rules)}")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, KeyError, TypeError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        sys.exit(1)
