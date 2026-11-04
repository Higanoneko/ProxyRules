#!/usr/bin/env python3
"""Convert selected classical rule sources to Egern native YAML."""

import argparse
import ipaddress
import sys
import urllib.request
from collections import defaultdict
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
RULES = ROOT / "Base/Rules/RemoteRules.yaml"
OUTPUT = ROOT / "Rules/Egern"

SOURCES = {"blackmatrix", "Higanoneko"}

KEYS = {
    "DOMAIN": "domain_set",
    "DOMAIN-SUFFIX": "domain_suffix_set",
    "DOMAIN-KEYWORD": "domain_keyword_set",
    "DOMAIN-REGEX": "domain_regex_set",
    "IP-CIDR": "ip_cidr_set",
    "IP-CIDR6": "ip_cidr6_set",
    "GEOIP": "geoip_set",
    "USER-AGENT": "user_agent_set",
    "URL-REGEX": "url_regex_set",
    "DEST-PORT": "dest_port_set",
    "ASN": "asn_set",
}

def source_url(rule):
    path = rule["remotefile"].removeprefix("./")
    if rule["category"] == "blackmatrix":
        return (
            "https://raw.githubusercontent.com/blackmatrix7/"
            "ios_rule_script/master/rule/Surge/" + path
        )
    if rule["category"] == "Higanoneko":
        return (
            "https://raw.githubusercontent.com/luluape/"
            "CustomProxyRuleset/main/" + path
        )
    raise ValueError("Unsupported source category")

def download(url):
    request = urllib.request.Request(
        url,
        headers={"User-Agent": "ProxyRules-Egern-Converter/1.0"},
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        return response.read().decode("utf-8-sig")

def convert(text, rule_id, allow_process=False):
    groups = defaultdict(list)
    unsupported = []
    skipped_process = []
    no_resolve = False

    for number, raw in enumerate(text.splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith(("#", ";", "//")):
            continue
        if line.startswith("[") and line.endswith("]"):
            continue

        parts = [p.strip() for p in line.split(",")]
        kind = parts[0].upper()

        if kind == "PROCESS-NAME":
            if not allow_process:
                unsupported.append(f"line {number}: {line[:120]}")
            else:
                skipped_process.append(f"line {number}: {line}")
            continue

        if kind not in KEYS or len(parts) < 2:
            unsupported.append(f"line {number}: {line[:120]}")
            continue

        value = parts[1]
        if not value:
            unsupported.append(f"line {number}: empty value")
            continue

        options = [p.lower() for p in parts[2:]]
        if any(p not in ("no-resolve",) for p in options):
            unsupported.append(f"line {number}: {line[:120]}")
            continue
        if "no-resolve" in options:
            no_resolve = True

        if kind in ("IP-CIDR", "IP-CIDR6"):
            try:
                ip = ipaddress.ip_network(value, strict=False)
                if ip.version != (4 if kind == "IP-CIDR" else 6):
                    raise ValueError("IP version mismatch")
            except ValueError:
                unsupported.append(f"line {number}: {line[:120]}")
                continue

        groups[KEYS[kind]].append(value)

    if unsupported:
        raise ValueError(
            f"{rule_id}: {len(unsupported)} unsupported rules:\n"
            + "\n".join(unsupported[:10])
        )

    if not groups:
        raise ValueError(f"{rule_id}: no supported rules")

    result = {}
    if no_resolve:
        result["no_resolve"] = True
    for key, values in groups.items():
        result[key] = list(dict.fromkeys(values))

    output = yaml.safe_dump(
        result,
        allow_unicode=True,
        sort_keys=False,
        width=120,
    )
    return (
        output,
        sum(len(v) for v in result.values() if isinstance(v, list)),
        skipped_process,
    )

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true")
    parser.add_argument(
        "--allow-unsupported-process", action="store_true",
        help="Explicitly exclude unsupported PROCESS-NAME rules and report them",
    )
    args = parser.parse_args()

    config = yaml.safe_load(RULES.read_text(encoding="utf-8"))
    results = []
    failures = []
    process_report = []

    for section in ("BaseRules", "CustomRules"):
        for rule_id, rule in config[section].items():
            if rule["category"] not in SOURCES:
                continue

            try:
                url = source_url(rule)
                content, count, skipped = convert(
                    download(url),
                    rule_id,
                    allow_process=args.allow_unsupported_process,
                )
                for item in skipped:
                    process_report.append(f"{rule_id}: {item}")
                results.append((rule_id, content))
                print(
                    f"PASS {rule_id:20} {count:6} rules"
                    f"  unsupported process rules: {len(skipped)}"
                )
            except Exception as exc:
                print(f"FAIL {rule_id}: {exc}")
                failures.append(rule_id)

    print(f"\nPASS: {len(results)} / EXPECTED: 16")
    if failures or len(results) != 16:
        print("Conversion incomplete; no files written.")
        return 1

    if args.write:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for rule_id, content in results:
            (OUTPUT / f"external_{rule_id}.yaml").write_text(
                content, encoding="utf-8"
            )
        report = OUTPUT / "unsupported_process_rules.txt"
        report.write_text(
            "# PROCESS-NAME entries unsupported by Egern native rules\n"
            + "\n".join(process_report)
            + "\n",
            encoding="utf-8",
        )
        print(f"Wrote {len(results)} files to {OUTPUT}")
        print(f"Wrote conversion report to {report}")
    else:
        print("Dry run only; use --write after reviewing results.")
        if process_report:
            print("\nUnsupported PROCESS-NAME entries:")
            for item in process_report:
                print(" ", item)

    return 0

if __name__ == "__main__":
    sys.exit(main())
