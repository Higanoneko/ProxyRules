#!/usr/bin/env python3
"""Read-only Loyalsoldier catalog synchronization preflight."""

from pathlib import Path
import sys
import yaml

ROOT = Path(__file__).resolve().parents[1]

catalog = yaml.safe_load(
    (ROOT / "Base/Rules/LoyalsoldierCatalog.yaml")
    .read_text(encoding="utf-8")
)["rules"]

remote = yaml.safe_load(
    (ROOT / "Base/Rules/RemoteRules.yaml")
    .read_text(encoding="utf-8")
)

PLACEMENTS = {
    "telegramcidr": ("BaseRules", "LS_TelegramIP", "Telegram"),
    "lancidr": ("BaseRules", "LS_LanCIDR", "LocalNetworkIP"),
    "direct": ("CustomRules", "LS_Direct", "Nikke"),
    "proxy": ("CustomRules", "LS_Proxy", "LS_Direct"),
}

errors = []

for name, (section, rule_id, anchor) in PLACEMENTS.items():
    spec = catalog[name]
    rules = remote[section]
    keys = list(rules)

    if not spec["enabled"]:
        errors.append(f"{name}: expected enabled")
        continue

    if rule_id not in rules:
        errors.append(f"{name}: missing {rule_id}")
        continue

    config = rules[rule_id]

    if config.get("category") != "Loyalsoldier":
        errors.append(f"{name}: wrong category")

    if config.get("remotefile") != name:
        errors.append(f"{name}: wrong remote file")

    if config.get("behavior") != spec["behavior"]:
        errors.append(f"{name}: behavior mismatch")

    if config.get("policyname") != spec["policy"]:
        errors.append(f"{name}: policy mismatch")

    if anchor not in keys:
        errors.append(f"{name}: missing anchor {anchor}")
    elif keys.index(rule_id) != keys.index(anchor) + 1:
        errors.append(f"{name}: expected immediately after {anchor}")

    print(f"CHECK {name}: {section}/{rule_id} after {anchor}")

if errors:
    for error in errors:
        print("FAIL", error)
    sys.exit(1)

print("\nPASS: Loyalsoldier placement preflight")
