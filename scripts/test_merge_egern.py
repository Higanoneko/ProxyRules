#!/usr/bin/env python3
"""Tests for non-destructive Egern profile merging."""

import copy
import unittest

from merge_egern import merge, dump

URL = "https://raw.githubusercontent.com/luluape/ProxyRules/main/Rules/Egern/"

class MergeEgernTests(unittest.TestCase):

    def setUp(self):
        self.original = {
            "dns": {"upstreams": {"HomeAdGuard": ["192.168.100.64"]}},
            "rules": [
                {"domain": {
                    "match": "jellyfin.127369.xyz",
                    "policy": "DIRECT",
                }},
                {"domain": {
                    "match": "immich.127369.xyz",
                    "policy": "DIRECT",
                }},
                {"default": {"policy": "Proxy"}},
            ],
            "policy_groups": [
                {"select": {"name": "Proxy", "policies": ["DIRECT"]}},
                {"select": {"name": "Telegram", "policies": ["Proxy"]}},
            ],
        }
        self.fragment = {
            "rules": [
                {"rule_set": {
                    "match": URL + name + ".yaml",
                    "policy": policy,
                    "update_interval": 86400,
                }}
                for name, policy in (
                    ("telegramcidr", "Telegram"),
                    ("lancidr", "DIRECT"),
                    ("direct", "DIRECT"),
                    ("proxy", "选择代理"),
                )
            ]
        }
        self.mapping = {"选择代理": "Proxy"}

    def test_four_rules_inserted(self):
        result, count = merge(
            self.original, self.fragment, self.mapping
        )
        self.assertEqual(count, 4)
        self.assertEqual(len(result["rules"]), 7)

    def test_idempotency(self):
        first, count1 = merge(
            self.original, self.fragment, self.mapping
        )
        second, count2 = merge(
            first, self.fragment, self.mapping
        )
        self.assertEqual((count1, count2), (4, 0))
        self.assertEqual(first, second)

    def test_original_unchanged(self):
        before = copy.deepcopy(self.original)
        merge(self.original, self.fragment, self.mapping)
        self.assertEqual(self.original, before)

    def test_dns_and_home_rules_preserved(self):
        result, _ = merge(
            self.original, self.fragment, self.mapping
        )
        self.assertEqual(result["dns"], self.original["dns"])
        self.assertEqual(
            result["rules"][:2], self.original["rules"][:2]
        )
        self.assertEqual(
            result["policy_groups"], self.original["policy_groups"]
        )
        self.assertIn("default", result["rules"][-1])

    def test_missing_policy_rejected(self):
        with self.assertRaisesRegex(ValueError, "Undefined"):
            merge(self.original, self.fragment, {})

    def test_conflicting_existing_rule_rejected(self):
        config = copy.deepcopy(self.original)
        config["rules"].insert(2, {
            "rule_set": {
                "match": URL + "proxy.yaml",
                "policy": "DIRECT",
            }
        })
        with self.assertRaisesRegex(ValueError, "different policy"):
            merge(config, self.fragment, self.mapping)

    def test_invalid_default_position_rejected(self):
        config = copy.deepcopy(self.original)
        config["rules"].insert(0, config["rules"].pop())
        with self.assertRaisesRegex(ValueError, "default"):
            merge(config, self.fragment, self.mapping)

    def test_output_yaml_parseable(self):
        from ruamel.yaml import YAML
        result, _ = merge(
            self.original, self.fragment, self.mapping
        )
        text = dump(result)
        parsed = YAML(typ="safe").load(text)
        self.assertEqual(parsed["rules"], result["rules"])


if __name__ == "__main__":
    unittest.main(verbosity=2)
