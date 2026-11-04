#!/usr/bin/env python3
"""Publish only credential-free, structurally validated Egern templates."""
import argparse
import os
from pathlib import Path
import tempfile
import yaml

PLACEHOLDER = "https://example.invalid/REPLACE_WITH_YOUR_EGERN_SUBSCRIPTION"
FILES = ("Egern_config.yaml", "Egern_config_no_ipv6.yaml")


def verify(path: Path, expected_ipv6: bool) -> str:
    content = path.read_text(encoding="utf-8")
    doc = yaml.safe_load(content)
    if not isinstance(doc, dict):
        raise ValueError(f"not a YAML mapping: {path}")
    if doc.get("ipv6") is not expected_ipv6:
        raise ValueError(f"wrong IPv6 setting: {path}")
    groups = doc.get("policy_groups", [])
    rules = doc.get("rules", [])
    if len(groups) != 29 or len(rules) != 31:
        raise ValueError(f"wrong groups/rules count: {path}")
    if sum("rule_set" in r for r in rules) != 30:
        raise ValueError(f"wrong remote rule count: {path}")
    if rules[-1] != {"default": {"policy": "选择代理"}}:
        raise ValueError(f"wrong default rule: {path}")
    subscription_urls = []
    for group in groups:
        if not isinstance(group, dict) or len(group) != 1:
            raise ValueError(f"invalid group: {path}")
        details = next(iter(group.values()))
        subscription_urls.extend(details.get("urls", []))
    if subscription_urls != [PLACEHOLDER]:
        raise ValueError(
            "REFUSING TO PUBLISH: subscription URL is not the public placeholder"
        )
    if content.count(PLACEHOLDER) != 1:
        raise ValueError("unexpected placeholder count")
    return content


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--destination", type=Path, required=True)
    args = parser.parse_args()
    rendered = []
    for name, ipv6 in zip(FILES, (True, False)):
        content = verify(args.source / name, ipv6)
        banner = (
            "# PUBLIC BETA TEMPLATE — NO PROXY NODES INCLUDED\n"
            "# Replace the placeholder subscription URL with your own Egern-compatible HTTPS URL.\n"
            "# Never commit a personalized copy of this file to a public repository.\n\n"
        )
        rendered.append((name, banner + content))
    args.destination.mkdir(parents=True, exist_ok=True)
    for name, content in rendered:
        target = args.destination / name
        fd, tmp_name = tempfile.mkstemp(prefix=".egern-", dir=args.destination)
        try:
            with os.fdopen(fd, "w", encoding="utf-8") as output:
                output.write(content)
            os.replace(tmp_name, target)
        finally:
            if os.path.exists(tmp_name):
                os.unlink(tmp_name)
        print(f"PASS public template: {target}")


if __name__ == "__main__":
    main()
