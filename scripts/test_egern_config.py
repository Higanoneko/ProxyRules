#!/usr/bin/env python3
"""Regression tests for Egern rules-fragment generation."""

import copy
import unittest

from generate_egern_config import (
    CATALOG, REMOTE, RULES,
    read_yaml, build_rules, generate,
)


class EgernConfigTests(unittest.TestCase):

    def setUp(self):
        self.catalog = copy.deepcopy(read_yaml(CATALOG))
        self.remote = copy.deepcopy(read_yaml(REMOTE))

    def build(self):
        return build_rules(self.catalog, self.remote)

    def test_current_rules_and_order(self):
        rules = self.build()
        actual = [
            (r["rule_set"]["match"].rsplit("/", 1)[-1],
             r["rule_set"]["policy"])
            for r in rules
        ]

        self.assertEqual(actual, [
            ("telegramcidr.yaml", "Telegram"),
            ("lancidr.yaml", "DIRECT"),
            ("direct.yaml", "DIRECT"),
            ("proxy.yaml", "选择代理"),
        ])

    def test_enable_icloud_requires_remote_sync(self):
        self.catalog["rules"]["icloud"]["enabled"] = True

        with self.assertRaises(ValueError):
            self.build()

    def test_disable_direct_requires_remote_sync(self):
        self.catalog["rules"]["direct"]["enabled"] = False

        with self.assertRaises(ValueError):
            self.build()

    def test_policy_mapping(self):
        spec = self.catalog["rules"]["proxy"]
        spec["policy"] = "直接连接"
        self.remote["CustomRules"]["LS_Proxy"]["policyname"] = "直接连接"

        rules = self.build()
        self.assertEqual(rules[-1]["rule_set"]["policy"], "DIRECT")

    def test_policy_mismatch_rejected(self):
        self.remote["CustomRules"]["LS_Proxy"]["policyname"] = "Other"

        with self.assertRaises(ValueError):
            self.build()

    def test_unknown_loyalsoldier_rule_rejected(self):
        self.remote["CustomRules"]["LS_Unknown"] = {
            "category": "Loyalsoldier",
        }

        with self.assertRaises(ValueError):
            self.build()

    def test_generated_yaml(self):
        rules, content = generate()
        import yaml
        parsed = yaml.safe_load(content)

        self.assertEqual(parsed["rules"], rules)
        self.assertEqual(len(rules), 4)

    def test_rule_files_exist(self):
        for name in ("telegramcidr", "lancidr", "direct", "proxy"):
            self.assertTrue((RULES / f"{name}.yaml").is_file())


if __name__ == "__main__":
    unittest.main(verbosity=2)
