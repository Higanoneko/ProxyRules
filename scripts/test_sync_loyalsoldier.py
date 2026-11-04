#!/usr/bin/env python3
"""Regression tests for Loyalsoldier catalog synchronization."""

import copy
import unittest

from sync_loyalsoldier import (
    CATALOG,
    REMOTE,
    load,
    synchronize,
)

SECTIONS = ("BaseRules", "CustomRules")


class SyncTests(unittest.TestCase):

    def setUp(self):
        self.catalog = copy.deepcopy(load(CATALOG)["rules"])
        self.remote = copy.deepcopy(load(REMOTE))

    def sync(self):
        synchronize(self.catalog, self.remote)

    def test_current_config_unchanged(self):
        before = copy.deepcopy(self.remote)
        self.sync()
        for section in SECTIONS:
            self.assertEqual(
                list(before[section].items()),
                list(self.remote[section].items()),
            )

    def test_enable_icloud(self):
        self.catalog["icloud"]["enabled"] = True
        self.sync()

        rules = self.remote["BaseRules"]
        keys = list(rules)

        self.assertEqual(
            keys.index("LS_iCloud"),
            keys.index("Apple") + 1,
        )
        self.assertEqual(
            rules["LS_iCloud"]["remotefile"],
            "icloud",
        )

    def test_disable_lancidr(self):
        self.catalog["lancidr"]["enabled"] = False
        self.sync()

        self.assertNotIn(
            "LS_LanCIDR",
            self.remote["BaseRules"],
        )

    def test_change_proxy_policy(self):
        self.catalog["proxy"]["policy"] = "DIRECT"
        self.sync()

        self.assertEqual(
            self.remote["CustomRules"]["LS_Proxy"]["policyname"],
            "DIRECT",
        )

    def test_disabled_anchor_fallback(self):
        self.catalog["apple"]["enabled"] = True
        self.catalog["icloud"]["enabled"] = False
        self.sync()

        keys = list(self.remote["BaseRules"])

        self.assertEqual(
            keys.index("LS_Apple"),
            keys.index("Apple") + 1,
        )

    def test_idempotency(self):
        self.catalog["icloud"]["enabled"] = True
        self.sync()

        first = copy.deepcopy(self.remote)
        self.sync()

        for section in SECTIONS:
            self.assertEqual(
                list(first[section].items()),
                list(self.remote[section].items()),
            )

    def test_missing_anchor(self):
        self.catalog["icloud"]["enabled"] = True
        self.catalog["icloud"]["after"] = "MissingAnchor"

        with self.assertRaises(ValueError):
            self.sync()

    def test_reject_blocked(self):
        self.catalog["reject"]["enabled"] = True

        with self.assertRaises(ValueError):
            self.sync()

    def test_applications_blocked(self):
        self.catalog["applications"]["enabled"] = True

        with self.assertRaises(ValueError):
            self.sync()

    def test_unrelated_order_preserved(self):
        before = {
            section: [
                key
                for key, value in self.remote[section].items()
                if value.get("category") != "Loyalsoldier"
            ]
            for section in SECTIONS
        }

        self.catalog["icloud"]["enabled"] = True
        self.catalog["apple"]["enabled"] = True
        self.catalog["lancidr"]["enabled"] = False
        self.sync()

        for section in SECTIONS:
            after = [
                key
                for key, value in self.remote[section].items()
                if value.get("category") != "Loyalsoldier"
            ]
            self.assertEqual(before[section], after)


if __name__ == "__main__":
    unittest.main(verbosity=2)
