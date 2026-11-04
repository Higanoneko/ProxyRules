#!/usr/bin/env python3
import argparse
import ipaddress
import urllib.parse
import urllib.request
from collections import defaultdict
from pathlib import Path
import yaml

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "Rules/Egern"
BASE = ROOT / "Base/Rules"
SOURCES = {"Sukka", "dler", "ADBlock"}
TYPES = {
    "DOMAIN": "domain_set",
    "DOMAIN-SUFFIX": "domain_suffix_set",
    "DOMAIN-KEYWORD": "domain_keyword_set",
    "DOMAIN-WILDCARD": "domain_wildcard_set",
    "DOMAIN-REGEX": "domain_regex_set",
    "IP-CIDR": "ip_cidr_set",
    "IP-CIDR6": "ip_cidr6_set",
    "USER-AGENT": "user_agent_set",
    "URL-REGEX": "url_regex_set",
    "GEOIP": "geoip_set",
    "ASN": "asn_set",
    "DEST-PORT": "dest_port_set",
}

def get_url(rule, links):
    c = rule["category"]
    template = links["Categories"][c]["url"]
    tool = links["Categories_Tools_List"].get(c, {}).get("fallback", "Surge")
    ext = links["Categories_Filetype_List"].get(c, {}).get("fallback", "")
    remote = rule["remotefile"].removeprefix("./")
    if "{filetype}" in template and Path(remote).suffix:
        remote = str(Path(remote).with_suffix(""))
    url = template.replace("{proxytools}", tool)
    url = url.replace("{remotefile}", remote).replace("{filetype}", ext)
    return urllib.parse.quote(url, safe=":/?&=%#")

def convert(text, name):
    data = defaultdict(list)
    skipped = []
    unsupported = []
    no_resolve = False

    for number, raw in enumerate(text.splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith(("#", ";", "//")):
            continue
        if line.startswith("[") and line.endswith("]"):
            continue
        parts = [x.strip() for x in line.split(",")]
        kind = parts[0].upper()

        if kind == "PROCESS-NAME":
            skipped.append(f"{name}:{number}: {line}")
            continue

        if kind not in TYPES or len(parts) < 2 or not parts[1]:
            unsupported.append(f"{name}:{number}: {line[:120]}")
            continue

        value = parts[1]
        options = [x.lower() for x in parts[2:]]
        if any(x != "no-resolve" for x in options):
            unsupported.append(f"{name}:{number}: {line[:120]}")
            continue

        if "no-resolve" in options:
            if kind not in ("IP-CIDR", "IP-CIDR6", "GEOIP", "ASN"):
                unsupported.append(f"{name}:{number}: {line[:120]}")
                continue
            no_resolve = True

        if kind in ("IP-CIDR", "IP-CIDR6"):
            try:
                ip = ipaddress.ip_network(value, strict=False)
                if ip.version != (4 if kind == "IP-CIDR" else 6):
                    raise ValueError("IP version mismatch")
            except ValueError:
                unsupported.append(f"{name}:{number}: {line[:120]}")
                continue

        data[TYPES[kind]].append(value)

    if unsupported or not data:
        raise ValueError(
            f"{name}: unsupported={len(unsupported)} "
            + "; ".join(unsupported[:5])
        )

    result = {}
    if no_resolve:
        result["no_resolve"] = True
    for key, values in data.items():
        result[key] = list(dict.fromkeys(values))
    count = sum(len(v) for v in result.values() if isinstance(v, list))
    return yaml.safe_dump(
        result, allow_unicode=True, sort_keys=False, width=120
    ), count, skipped

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true")
    args = parser.parse_args()

    rules = yaml.safe_load((BASE / "RemoteRules.yaml").read_text())
    links = yaml.safe_load((BASE / "RemoteRulesLinkBase.yaml").read_text())
    files = {}
    skipped_all = []
    failures = []

    for section in ("BaseRules", "CustomRules"):
        for name, rule in rules[section].items():
            category = rule.get("category")
            if category not in SOURCES:
                continue
            try:
                url = get_url(rule, links)
                request = urllib.request.Request(
                    url, headers={"User-Agent": "ProxyRules-Egern/1.0"}
                )
                with urllib.request.urlopen(request, timeout=60) as response:
                    text = response.read().decode("utf-8-sig")
                content, count, skipped = convert(text, name)
                filename = f"external_{category}_{name}.yaml"
                files[filename] = content
                skipped_all.extend(skipped)
                print(f"PASS {name}: {count} rules, {len(skipped)} process rules")
            except Exception as exc:
                failures.append(name)
                print(f"FAIL {name}: {exc}")

    print(f"\nConverted: {len(files)}/10")
    print(f"Unsupported PROCESS-NAME: {len(skipped_all)}")

    if failures or len(files) != 10:
        print("Incomplete conversion; no files written")
        return 1

    if args.write:
        OUT.mkdir(parents=True, exist_ok=True)
        for filename, content in files.items():
            (OUT / filename).write_text(content, encoding="utf-8")
        (OUT / "unsupported_process_remaining.txt").write_text(
            "\n".join(skipped_all) + "\n", encoding="utf-8"
        )
        print("PASS: wrote 10 YAML files")
    else:
        print("Dry run: no files written")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
