"""RuleLoader 统一规则模型测试。"""

import sys
import unittest
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[2]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from Generator.core.rule_loader import RuleLoader


class RuleLoaderTestCase(unittest.TestCase):
    """RuleLoader 回归测试。"""

    @classmethod
    def setUpClass(cls):
        cls.loader = RuleLoader()

    def test_normalized_rules_contains_latest_remote_rule_keys(self):
        normalized_rules = self.loader.get_normalized_rules()
        self.assertIn("YouTube", normalized_rules)
        self.assertIn("YouTubeMusic", normalized_rules)
        self.assertIn("LocalNetwork", normalized_rules)
        self.assertIn("LocalNetworkIP", normalized_rules)

    def test_mihomo_rule_providers_use_latest_remote_rule_keys(self):
        providers = self.loader.generate_mihomo_rule_providers()
        self.assertIn("YouTube", providers)
        self.assertIn("YouTubeMusic", providers)
        self.assertIn("LocalNetwork", providers)
        self.assertIn("LocalNetworkIP", providers)

    def test_mihomo_rules_include_local_network_direct_policy(self):
        rules = self.loader.generate_mihomo_rules()
        self.assertIn("RULE-SET,LocalNetwork,DIRECT", rules)
        self.assertIn("RULE-SET,LocalNetworkIP,DIRECT", rules)

    def test_surge_remote_rules_include_youtube_and_local_network(self):
        surge_rules = self.loader.generate_surge_remote_rules()

        joined_rules = "\n".join(surge_rules)
        self.assertIn(
            "RULE-SET,https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Surge/YouTube/YouTube.list,YouTube",
            joined_rules,
        )
        self.assertIn(
            "RULE-SET,https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Surge/YouTubeMusic/YouTubeMusic.list,YouTube",
            joined_rules,
        )
        self.assertIn(
            "RULE-SET,https://ruleset.skk.moe/List/non_ip/lan.conf,DIRECT",
            joined_rules,
        )
        self.assertIn(
            "RULE-SET,https://ruleset.skk.moe/List/ip/lan.conf,DIRECT",
            joined_rules,
        )


if __name__ == "__main__":
    unittest.main()
