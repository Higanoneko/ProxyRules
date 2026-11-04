from pathlib import Path
import sys
import yaml

ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / "Base/Rules/LoyalsoldierCatalog.yaml"
REMOTE = ROOT / "Base/Rules/RemoteRules.yaml"

EXPECTED = {
    "reject", "icloud", "apple", "google",
    "proxy", "direct", "private", "gfw",
    "tld-not-cn", "telegramcidr", "cncidr",
    "lancidr", "applications",
}
BEHAVIORS = {"domain", "ipcidr", "classical"}
CLIENTS = {
    "Mihomo": "yaml",
    "Loon": "list",
    "Surge": "list",
}

errors = []

def fail(message):
    errors.append(message)
    print(f"FAIL {message}")

def load_yaml(path):
    with path.open(encoding="utf-8") as f:
        return yaml.safe_load(f)

catalog_doc = load_yaml(CATALOG)
remote = load_yaml(REMOTE)

catalog = catalog_doc.get("rules", {})
if set(catalog) != EXPECTED:
    fail(f"Catalog groups mismatch: {set(catalog) ^ EXPECTED}")

active = {}
for section_name in ("BaseRules", "CustomRules"):
    for rule_id, config in remote.get(section_name, {}).items():
        if config.get("category") != "Loyalsoldier":
            continue
        name = config.get("remotefile")
        if name in active:
            fail(f"Duplicate active group: {name}")
        active[name] = (rule_id, config)

for name, spec in catalog.items():
    behavior = spec.get("behavior")
    policy = spec.get("policy")
    enabled = spec.get("enabled")

    if behavior not in BEHAVIORS:
        fail(f"{name}: invalid behavior {behavior}")
    if not isinstance(policy, str) or not policy:
        fail(f"{name}: invalid policy")
    if not isinstance(enabled, bool):
        fail(f"{name}: enabled must be boolean")

    for client, extension in CLIENTS.items():
        path = ROOT / f"Rules/{client}/{name}.{extension}"
        if not path.is_file():
            fail(f"{name}: missing {path.relative_to(ROOT)}")
        elif path.stat().st_size == 0:
            fail(f"{name}: empty {path.relative_to(ROOT)}")

    if enabled and name not in active:
        fail(f"{name}: enabled but absent from RemoteRules.yaml")
    if not enabled and name in active:
        fail(f"{name}: disabled but present in RemoteRules.yaml")

    if name in active:
        rule_id, config = active[name]
        if config.get("behavior") != behavior:
            fail(f"{name}: behavior mismatch in {rule_id}")
        if config.get("policyname") != policy:
            fail(f"{name}: policy mismatch in {rule_id}")

    print(f"{'ON ' if enabled else 'OFF'} {name:15} {behavior:10} {policy}")

for name in active:
    if name not in catalog:
        fail(f"Unknown active Loyalsoldier group: {name}")

if errors:
    print(f"\nFAILED: {len(errors)} issue(s)")
    sys.exit(1)

print("\nPASS: 13 groups validated; active rules consistent")
