#!/usr/bin/env python3
"""Convert Loyalsoldier rule payloads into Egern native rule sets."""

import argparse
import ipaddress
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
INPUT = ROOT / "Rules" / "Mihomo"
OUTPUT = ROOT / "Rules" / "Egern"

DOMAIN_GROUPS = (
    "reject", "icloud", "apple", "google",
    "proxy", "direct", "private", "gfw", "tld-not-cn",
)
IP_GROUPS = ("telegramcidr", "cncidr", "lancidr")


def unique(values):
    return list(dict.fromkeys(values))


def convert_domain(payload):
    exact = []
    suffix = []

    for raw in payload:
        if not isinstance(raw, str):
            raise ValueError(f"Non-string domain rule: {raw!r}")

        value = raw.strip()

        if value.startswith("+."):
            value = value[2:]
            target = suffix
        else:
            target = exact

        if (
            not value
            or value.startswith(".")
            or value.endswith(".")
            or any(c in value for c in " \t\r\n,/*?")
        ):
            raise ValueError(f"Unsupported domain rule: {raw!r}")

        target.append(value)

    result = {}
    if exact:
        result["domain_set"] = unique(exact)
    if suffix:
        result["domain_suffix_set"] = unique(suffix)
    return result


def convert_ip(payload):
    ipv4 = []
    ipv6 = []

    for raw in payload:
        if not isinstance(raw, str):
            raise ValueError(f"Non-string IP rule: {raw!r}")

        value = raw.strip()
        network = ipaddress.ip_network(value, strict=False)

        if network.version == 4:
            ipv4.append(value)
        else:
            ipv6.append(value)

    result = {}
    if ipv4:
        result["ip_cidr_set"] = unique(ipv4)
    if ipv6:
        result["ip_cidr6_set"] = unique(ipv6)
    return result


def generate(name, behavior):
    source = INPUT / f"{name}.yaml"
    data = yaml.safe_load(source.read_text(encoding="utf-8"))

    if not isinstance(data, dict):
        raise ValueError(f"{name}: invalid YAML document")

    payload = data.get("payload")

    if not isinstance(payload, list) or not payload:
        raise ValueError(f"{name}: empty or invalid payload")

    if behavior == "domain":
        result = convert_domain(payload)
    elif behavior == "ipcidr":
        result = convert_ip(payload)
    else:
        raise ValueError(f"Unsupported behavior: {behavior}")

    if not result:
        raise ValueError(f"{name}: empty converted rule set")

    content = yaml.safe_dump(
        result,
        allow_unicode=True,
        sort_keys=False,
        default_flow_style=False,
        width=120,
    )

    # Verify that the resulting YAML is parseable.
    if yaml.safe_load(content) != result:
        raise ValueError(f"{name}: output verification failed")

    return content


def main():
    parser = argparse.ArgumentParser()
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--check", action="store_true")
    mode.add_argument("--write", action="store_true")
    args = parser.parse_args()

    changed = []

    for name in DOMAIN_GROUPS + IP_GROUPS:
        behavior = "domain" if name in DOMAIN_GROUPS else "ipcidr"
        content = generate(name, behavior)
        target = OUTPUT / f"{name}.yaml"

        current = (
            target.read_text(encoding="utf-8")
            if target.is_file()
            else None
        )

        if current != content:
            changed.append(name)
            if args.write:
                OUTPUT.mkdir(parents=True, exist_ok=True)
                target.write_text(content, encoding="utf-8")

        print(f"{name}: {'changed' if current != content else 'OK'}")

    if args.check and changed:
        print(f"FAIL: {len(changed)} Egern files need generation")
        return 1

    print(f"PASS: {len(DOMAIN_GROUPS) + len(IP_GROUPS)} Egern rule sets verified")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, yaml.YAMLError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        sys.exit(1)
