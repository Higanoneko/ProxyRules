#!/usr/bin/env python3
"""Synchronize Loyalsoldier catalog into RemoteRules.yaml."""

import argparse
import os
import sys
import tempfile
from pathlib import Path

from ruamel.yaml import YAML
from ruamel.yaml.comments import CommentedMap

ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / "Base/Rules/LoyalsoldierCatalog.yaml"
REMOTE = ROOT / "Base/Rules/RemoteRules.yaml"

SECTIONS = ("BaseRules", "CustomRules")
BLOCKED = {"reject", "applications"}

yaml = YAML()
yaml.preserve_quotes = True
yaml.width = 4096


def load(path):
    with path.open(encoding="utf-8") as f:
        return yaml.load(f)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def make_rule(name, spec, existing=None):
    if existing is None:
        existing = CommentedMap()

    rule_id = spec["rule_id"]

    defaults = {
        "name": rule_id,
        "policyname": spec["policy"],
        "tagname": f"Loyalsoldier {name}",
        "category": "Loyalsoldier",
        "behavior": spec["behavior"],
        "remotefile": name,
    }

    for key, value in defaults.items():
        if key not in existing or key in (
            "name",
            "policyname",
            "category",
            "behavior",
            "remotefile",
        ):
            existing[key] = value

    return existing


def resolve_anchor(name, catalog):
    """Follow disabled Loyalsoldier anchors to an active anchor."""
    section = catalog[name]["section"]
    anchor = catalog[name]["after"]
    visited = {name}

    while anchor is not None:
        matches = [
            (n, spec)
            for n, spec in catalog.items()
            if spec["rule_id"] == anchor
        ]

        if not matches:
            return anchor

        parent_name, parent = matches[0]

        require(
            parent_name not in visited,
            f"{name}: circular anchor dependency",
        )
        visited.add(parent_name)

        require(
            parent["section"] == section,
            f"{name}: anchor crosses sections",
        )

        if parent["enabled"]:
            return anchor

        anchor = parent["after"]

    raise ValueError(f"{name}: no usable anchor")


def synchronize(catalog, remote):
    require(isinstance(catalog, dict), "Invalid catalog")
    require(isinstance(remote, dict), "Invalid RemoteRules")

    ids = {}
    for name, spec in catalog.items():
        for field in (
            "rule_id",
            "section",
            "after",
            "enabled",
            "policy",
            "behavior",
        ):
            require(field in spec, f"{name}: missing {field}")

        require(
            spec["section"] in SECTIONS,
            f"{name}: invalid section",
        )
        require(
            isinstance(spec["enabled"], bool),
            f"{name}: enabled must be boolean",
        )
        require(
            isinstance(spec["policy"], str) and bool(spec["policy"]),
            f"{name}: invalid policy",
        )
        require(
            spec["behavior"] in ("domain", "ipcidr", "classical"),
            f"{name}: invalid behavior",
        )
        require(
            isinstance(spec["rule_id"], str)
            and bool(spec["rule_id"]),
            f"{name}: invalid rule_id",
        )
        require(
            spec["after"] is None
            or isinstance(spec["after"], str),
            f"{name}: invalid anchor",
        )
        require(
            spec["rule_id"] not in ids,
            f"{name}: duplicate rule_id",
        )

        ids[spec["rule_id"]] = name

        require(
            not (name in BLOCKED and spec["enabled"]),
            f"{name}: automatic enabling is blocked",
        )
        require(
            not (spec["enabled"] and spec["after"] is None),
            f"{name}: enabled rule requires an anchor",
        )

    original = {
        section: list(remote[section])
        for section in SECTIONS
    }

    saved = {}

    # First validate ownership; never remove an unrelated rule.
    for section in SECTIONS:
        rules = remote[section]

        for rule_id, config in list(rules.items()):
            if config.get("category") == "Loyalsoldier":
                require(
                    rule_id in ids,
                    f"{section}/{rule_id}: unknown Loyalsoldier rule",
                )
                name = ids[rule_id]
                require(
                    spec_section := catalog[name]["section"],
                    f"{name}: missing section",
                )
                require(
                    spec_section == section,
                    f"{name}: rule exists in wrong section",
                )
                require(
                    rule_id not in saved,
                    f"{name}: duplicate existing rule",
                )
                saved[rule_id] = config

            elif rule_id in ids:
                raise ValueError(
                    f"{section}/{rule_id}: ID owned by another category"
                )

    # Remove only catalog-owned entries from the working YAML tree.
    for section in SECTIONS:
        rules = remote[section]
        for rule_id in list(rules):
            if rule_id in saved:
                del rules[rule_id]

    # Insert enabled entries in dependency order.
    pending = {
        name for name, spec in catalog.items()
        if spec["enabled"]
    }

    while pending:
        progress = False

        for name in list(catalog):
            if name not in pending:
                continue

            spec = catalog[name]
            section = spec["section"]
            rules = remote[section]
            anchor = resolve_anchor(name, catalog)

            if anchor not in rules:
                # Wait if the anchor is another enabled catalog rule.
                if anchor in ids and ids[anchor] in pending:
                    continue

                raise ValueError(
                    f"{name}: anchor {anchor!r} does not exist "
                    f"in {section}"
                )

            rule_id = spec["rule_id"]

            require(
                rule_id not in rules,
                f"{name}: duplicate insertion",
            )

            entry = make_rule(
                name,
                spec,
                saved.get(rule_id),
            )

            position = list(rules).index(anchor) + 1

            # Preserve catalog order for siblings sharing an anchor.
            while position < len(rules):
                next_id = list(rules)[position]
                next_name = ids.get(next_id)

                if next_name is None:
                    break

                if resolve_anchor(next_name, catalog) != anchor:
                    break

                position += 1

            rules.insert(position, rule_id, entry)

            pending.remove(name)
            progress = True

        require(
            progress,
            f"Unresolved rule dependencies: {sorted(pending)}",
        )

    # Validate resulting entries and preserve unrelated ordering.
    for section in SECTIONS:
        before = [
            rule_id for rule_id in original[section]
            if rule_id not in ids
        ]
        after = [
            rule_id for rule_id in remote[section]
            if rule_id not in ids
        ]

        require(
            before == after,
            f"{section}: unrelated rule ordering changed",
        )

    return original


def main():
    parser = argparse.ArgumentParser()
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument("--check", action="store_true")
    mode.add_argument("--write", action="store_true")
    args = parser.parse_args()

    catalog_doc = load(CATALOG)
    require(
        isinstance(catalog_doc, dict)
        and isinstance(catalog_doc.get("rules"), dict),
        "Invalid catalog structure",
    )

    remote = load(REMOTE)
    catalog = catalog_doc["rules"]

    before = REMOTE.read_text(encoding="utf-8")
    original = load(REMOTE)

    synchronize(catalog, remote)

    # Compare ordered mappings semantically to avoid rewriting YAML
    # solely because of formatting differences.
    changed = any(
        list(original[section].items())
        != list(remote[section].items())
        for section in SECTIONS
    )

    if not changed:
        print("PASS: already synchronized; zero changes")
        return

    if not args.write:
        print("FAIL: RemoteRules.yaml requires synchronization")
        print("Run with --write to apply the changes")
        sys.exit(1)

    fd, tmp = tempfile.mkstemp(
        prefix=".RemoteRules.",
        suffix=".yaml",
        dir=REMOTE.parent,
    )

    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            yaml.dump(remote, f)

        # Check that the output can be parsed.
        load(Path(tmp))

        # Preserve original file permissions.
        os.chmod(tmp, REMOTE.stat().st_mode & 0o777)
        os.replace(tmp, REMOTE)

    finally:
        if os.path.exists(tmp):
            os.unlink(tmp)

    print("PASS: RemoteRules.yaml synchronized")
    print("Review with: git diff -- Base/Rules/RemoteRules.yaml")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, KeyError, TypeError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        sys.exit(1)
