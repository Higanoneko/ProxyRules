#!/usr/bin/env python3
"""Safely merge generated ProxyRules into an existing Egern profile."""

import argparse
import difflib
import sys
from copy import deepcopy
from io import StringIO
from pathlib import Path

from ruamel.yaml import YAML

ROOT = Path(__file__).resolve().parents[1]
FRAGMENT = ROOT / "Config/Egern/Egern_rules.yaml"

yaml = YAML(typ="rt")
yaml.preserve_quotes = True
yaml.indent(mapping=2, sequence=4, offset=2)
yaml.width = 120


def load(path):
    with path.open(encoding="utf-8") as f:
        data = yaml.load(f)
    if not isinstance(data, dict):
        raise ValueError(f"Invalid YAML mapping: {path}")
    return data


def dump(data):
    buf = StringIO()
    yaml.dump(data, buf)
    return buf.getvalue()


def policy_names(config):
    names = {"DIRECT", "REJECT"}

    for section in ("policy_groups", "proxies"):
        entries = config.get(section, [])
        if not isinstance(entries, list):
            raise ValueError(f"{section} must be a list")

        for item in entries:
            if not isinstance(item, dict) or len(item) != 1:
                raise ValueError(f"Invalid {section} entry")

            body = next(iter(item.values()))
            if not isinstance(body, dict):
                raise ValueError(f"Invalid {section} entry body")

            name = body.get("name")
            if isinstance(name, str) and name:
                if name in names:
                    raise ValueError(f"Duplicate policy name: {name}")
                names.add(name)

    return names


def resolve_policy(value, mapping):
    return mapping.get(value, value)


def merge(config, fragment, mapping):
    result = deepcopy(config)
    current = result.get("rules")

    if not isinstance(current, list):
        raise ValueError("Profile must contain a rules list")

    generated = fragment.get("rules")
    if not isinstance(generated, list) or not generated:
        raise ValueError("Fragment must contain rules")

    available = policy_names(result)

    defaults = [
        i for i, item in enumerate(current)
        if isinstance(item, dict) and "default" in item
    ]

    if len(defaults) != 1 or defaults[0] != len(current) - 1:
        raise ValueError(
            "Expected exactly one final default rule; "
            "manual review required"
        )

    existing = {}
    for item in current:
        if not isinstance(item, dict):
            raise ValueError("Invalid rule entry")
        if "rule_set" not in item:
            continue
        body = item["rule_set"]
        if not isinstance(body, dict):
            raise ValueError("Invalid rule_set")
        url = body.get("match")
        if isinstance(url, str):
            if url in existing:
                raise ValueError(f"Duplicate existing rule set: {url}")
            existing[url] = body

    additions = []

    for item in generated:
        if not isinstance(item, dict) or set(item) != {"rule_set"}:
            raise ValueError("Unexpected generated rule")

        body = deepcopy(item["rule_set"])
        if not isinstance(body, dict):
            raise ValueError("Invalid generated rule_set")

        url = body.get("match")
        policy = body.get("policy")
        if not isinstance(url, str) or not url.startswith("https://"):
            raise ValueError("Invalid remote rule URL")
        if not isinstance(policy, str):
            raise ValueError("Invalid generated policy")

        body["policy"] = resolve_policy(policy, mapping)
        if body["policy"] not in available:
            raise ValueError(
                f"Undefined Egern policy: {body['policy']}. "
                "Use --map-policy to specify an existing policy."
            )

        if url in existing:
            old = existing[url]
            if old.get("policy") != body["policy"]:
                raise ValueError(
                    f"Existing rule has different policy: {url}"
                )
            continue

        if any(x["rule_set"]["match"] == url for x in additions):
            raise ValueError(f"Duplicate generated rule set: {url}")

        additions.append({"rule_set": body})

    # Preserve all old rules and insert immediately before final default.
    for item in additions:
        current.insert(len(current) - 1, item)

    return result, len(additions)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True, type=Path)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--write", action="store_true")
    parser.add_argument(
        "--map-policy", action="append", default=[],
        metavar="SOURCE=TARGET"
    )
    args = parser.parse_args()

    if args.write:
        if args.output is None:
            parser.error("--write requires --output")
        if args.input.resolve() == args.output.resolve():
            parser.error("Refusing to overwrite original profile")
        if args.output.exists():
            parser.error("Refusing to overwrite existing output")

    mapping = {}
    for entry in args.map_policy:
        if "=" not in entry:
            parser.error("--map-policy requires SOURCE=TARGET")
        source, target = entry.split("=", 1)
        if not source or not target or source in mapping:
            parser.error("Invalid or duplicate policy mapping")
        mapping[source] = target

    original = args.input.read_text(encoding="utf-8")
    config = load(args.input)
    fragment = load(FRAGMENT)

    merged, count = merge(config, fragment, mapping)
    rendered = dump(merged)

    changes = "".join(difflib.unified_diff(
        original.splitlines(keepends=True),
        rendered.splitlines(keepends=True),
        fromfile=str(args.input),
        tofile=str(args.output or "preview"),
    ))

    print(f"PASS: {count} rules added")
    print(f"PASS: {len(fragment['rules']) - count} rules already present")
    print("PASS: original file is unchanged")

    if args.write:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(rendered, encoding="utf-8")
        print(f"OUTPUT: {args.output}")
    else:
        print(changes or "No differences")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        sys.exit(1)
