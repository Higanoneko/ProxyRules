"""
规则加载器。
负责读取和解析 RemoteRules.yaml 与 RemoteRulesLinkBase.yaml，
并基于统一规则规格输出各目标平台的专有配置片段。
"""

from typing import Any, Dict, List, Optional, Tuple
from pathlib import Path
import sys

# 添加项目根目录到 Python 路径
project_root = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(project_root))

from Generator.utils.yaml_helper import YamlHelper
from Generator.utils.file_helper import FileHelper
from Generator.core.remote_rule_specs import (
    LOON_RULE_ORDER,
    MIHOMO_RULE_ORDER,
    REMOTE_RULE_ID_ALIASES,
    REMOTE_RULE_SPEC_INDEX,
    SURGE_RULE_ORDER,
    RemoteRuleSpec,
)


class RuleLoader:
    """规则加载和 URL 生成器。"""

    def __init__(self):
        """初始化规则加载器。"""
        self.project_root = FileHelper.get_project_root()
        self.rules_file = self.project_root / "Base" / "Rules" / "RemoteRules.yaml"
        self.link_base_file = self.project_root / "Base" / "Rules" / "RemoteRulesLinkBase.yaml"

        # 加载规则文件
        self.rules_data = YamlHelper.load_yaml(self.rules_file)
        self.link_base_data = YamlHelper.load_yaml(self.link_base_file)

        # 提取关键数据
        self.rules = self.rules_data.get("rules", {})
        self.categories = self.link_base_data.get("Categories", {})
        self.tools_mapping = self.link_base_data.get("Categories_Tools_List", {})
        self.filetype_mapping = self.link_base_data.get("Categories_Filetype_List", {})

    def get_tool_type_mapping(self, category: str, proxy_tool: str) -> str:
        """
        获取代理工具类型映射。

        Args:
            category: 规则分类（如 Sukka、blackmatrix）。
            proxy_tool: 代理工具名称（如 Mihomo、Loon）。

        Returns:
            映射后的工具类型字符串。
        """
        if category not in self.tools_mapping:
            return proxy_tool

        mapping = self.tools_mapping[category]
        if proxy_tool in mapping:
            return mapping[proxy_tool]

        return mapping.get("fallback", proxy_tool)

    def get_filetype_mapping(self, category: str, proxy_tool: str) -> str:
        """
        获取文件类型映射。

        Args:
            category: 规则分类。
            proxy_tool: 代理工具名称。

        Returns:
            映射后的文件类型字符串。
        """
        if category not in self.filetype_mapping:
            return ""

        mapping = self.filetype_mapping[category]
        if proxy_tool in mapping:
            return mapping[proxy_tool]

        return mapping.get("fallback", "")

    def generate_rule_url(
        self,
        rule_name: str,
        rule_config: Dict[str, Any],
        proxy_tool: str,
    ) -> Optional[str]:
        """
        生成规则的完整 URL。

        Args:
            rule_name: 规则名称（保留参数，兼容旧调用）。
            rule_config: 规则配置字典。
            proxy_tool: 代理工具类型（Mihomo/Loon/Stash/Surge）。

        Returns:
            完整规则 URL；无法生成时返回 None。
        """
        _ = rule_name  # 兼容旧签名，当前 URL 仅依赖 rule_config

        category = rule_config.get("category")
        remotefile = rule_config.get("remotefile", "")

        if not category or category not in self.categories:
            return None

        url_template = self.categories[category].get("url", "")
        if not url_template:
            return None

        mapped_tool = self.get_tool_type_mapping(category, proxy_tool)

        clean_remotefile = remotefile
        if clean_remotefile.startswith("./"):
            clean_remotefile = clean_remotefile[2:]

        if "{filetype}" in url_template:
            if "." in clean_remotefile:
                clean_remotefile = clean_remotefile.rsplit(".", 1)[0]
            filetype = self.get_filetype_mapping(category, proxy_tool)
        else:
            filetype = ""

        url = url_template.replace("{proxytools}", mapped_tool)
        url = url.replace("{remotefile}", clean_remotefile)
        url = url.replace("{filetype}", filetype)
        return url

    def _flatten_rules(
        self,
        rules_dict: Dict[str, Any],
        parent_key: str = "",
    ) -> Dict[str, Dict[str, Any]]:
        """
        将嵌套规则字典扁平化。

        Args:
            rules_dict: 规则字典（可能嵌套）。
            parent_key: 父键名（保留参数，便于兼容后续扩展）。

        Returns:
            扁平化规则字典。
        """
        _ = parent_key

        flattened: Dict[str, Dict[str, Any]] = {}

        for key, value in rules_dict.items():
            if not isinstance(value, dict):
                continue

            if "name" in value and "category" in value:
                flattened[key] = value
                continue

            nested = self._flatten_rules(value, key)
            flattened.update(nested)

        return flattened

    def get_all_rules(self) -> Dict[str, Dict[str, Any]]:
        """
        获取扁平化后的原始规则。

        Returns:
            原始规则键 -> 规则配置。
        """
        return self._flatten_rules(self.rules)

    def _resolve_rule_id(self, raw_rule_key: str, rule_config: Dict[str, Any]) -> str:
        """将原始键名解析为统一规则 ID。"""
        if raw_rule_key in REMOTE_RULE_SPEC_INDEX:
            return raw_rule_key

        alias_rule_id = REMOTE_RULE_ID_ALIASES.get(raw_rule_key)
        if alias_rule_id:
            return alias_rule_id

        rule_name = rule_config.get("name")
        if isinstance(rule_name, str) and rule_name in REMOTE_RULE_SPEC_INDEX:
            return rule_name

        return raw_rule_key

    def get_normalized_rules(self) -> Dict[str, Dict[str, Any]]:
        """
        获取统一规则 ID 视图。

        Returns:
            统一规则 ID -> 规则配置。
        """
        normalized_rules: Dict[str, Dict[str, Any]] = {}

        for raw_rule_key, rule_config in self.get_all_rules().items():
            normalized_rule_id = self._resolve_rule_id(raw_rule_key, rule_config)
            if normalized_rule_id not in normalized_rules:
                normalized_rules[normalized_rule_id] = rule_config

        return normalized_rules

    def _iter_rules_for_target(
        self,
        proxy_tool: str,
        rule_order: List[str],
    ) -> List[Tuple[str, RemoteRuleSpec, Dict[str, Any], str]]:
        """按目标顺序返回可用规则。"""
        normalized_rules = self.get_normalized_rules()
        results: List[Tuple[str, RemoteRuleSpec, Dict[str, Any], str]] = []

        for rule_id in rule_order:
            rule_config = normalized_rules.get(rule_id)
            if not rule_config:
                continue

            url = self.generate_rule_url(rule_id, rule_config, proxy_tool)
            if not url:
                continue

            spec = REMOTE_RULE_SPEC_INDEX[rule_id]
            results.append((rule_id, spec, rule_config, url))

        return results

    @staticmethod
    def _detect_format_type(url: str) -> str:
        """根据 URL 推断规则格式。"""
        if url.endswith(".mrs"):
            return "mrs"
        if url.endswith(".yaml") or url.endswith(".yml"):
            return "yaml"
        return "text"

    @staticmethod
    def _provider_path_ext(format_type: str) -> str:
        """根据格式返回 provider 本地扩展名。"""
        if format_type == "yaml":
            return "yaml"
        if format_type == "mrs":
            return "mrs"
        return "list"

    def generate_mihomo_rule_providers(self) -> Dict[str, Dict[str, Any]]:
        """
        为 Mihomo/Stash 生成 rule-providers。

        Returns:
            rule-providers 字典（键统一为规则 ID）。
        """
        rule_providers: Dict[str, Dict[str, Any]] = {}

        for rule_id, _, rule_config, url in self._iter_rules_for_target("Mihomo", MIHOMO_RULE_ORDER):
            behavior = rule_config.get("behavior", "classical")
            provider_name = rule_config.get("name", rule_id)
            format_type = self._detect_format_type(url)
            path_ext = self._provider_path_ext(format_type)

            rule_providers[rule_id] = {
                "type": "http",
                "behavior": behavior,
                "format": format_type,
                "interval": 86400,
                "url": url,
                "path": f"./ruleset/{provider_name}.{path_ext}",
            }

        return rule_providers

    def generate_mihomo_rules(self) -> List[str]:
        """
        为 Mihomo/Stash 生成规则列表（RULE-SET 语法）。

        Returns:
            规则字符串列表。
        """
        rules: List[str] = []

        for rule_id, spec, _, _ in self._iter_rules_for_target("Mihomo", MIHOMO_RULE_ORDER):
            policy_name = spec.mihomo_policy_name or spec.policy_name
            rules.append(f"RULE-SET,{rule_id},{policy_name}")

        rules.append("GEOIP,CN,直接连接")
        rules.append("MATCH,选择代理")
        return rules

    def generate_loon_remote_rules(self) -> List[str]:
        """
        为 Loon 生成 [Remote Rule] 段规则列表。

        Returns:
            规则字符串列表。
        """
        remote_rules: List[str] = []

        for _, spec, _, url in self._iter_rules_for_target("Loon", LOON_RULE_ORDER):
            remote_rules.append(
                f"{url}, policy = {spec.policy_name}, tag = {spec.tag_name}, enabled = true"
            )

        return remote_rules

    def generate_surge_remote_rules(self) -> List[str]:
        """
        为 Surge 生成 [Rule] 段远程规则列表。

        Returns:
            规则字符串列表。
        """
        remote_rules: List[str] = []

        for _, spec, _, url in self._iter_rules_for_target("Surge", SURGE_RULE_ORDER):
            if remote_rules:
                remote_rules.append("")
            remote_rules.append(f"# {spec.tag_name}")

            if spec.surge_option:
                remote_rules.append(
                    f"RULE-SET,{url},{spec.policy_name},{spec.surge_option}"
                )
            else:
                remote_rules.append(f"RULE-SET,{url},{spec.policy_name}")

        return remote_rules

    def get_rule_names(self) -> List[str]:
        """
        获取所有规则名称列表（去重后）。

        Returns:
            规则名称列表。
        """
        names: List[str] = []
        for rule_config in self.get_normalized_rules().values():
            rule_name = rule_config.get("name")
            if rule_name and rule_name not in names:
                names.append(rule_name)
        return names

    def get_rules_by_category(self, category: str) -> Dict[str, Dict[str, Any]]:
        """
        按分类获取规则。

        Args:
            category: 分类名称。

        Returns:
            该分类下的所有规则（统一规则 ID 视图）。
        """
        return {
            rule_id: rule_config
            for rule_id, rule_config in self.get_normalized_rules().items()
            if rule_config.get("category") == category
        }


if __name__ == "__main__":
    loader = RuleLoader()

    print("=== 测试规则加载器 ===\n")

    normalized_rules = loader.get_normalized_rules()
    print(f"统一规则数量: {len(normalized_rules)}")
    print(f"前 5 个规则 ID: {list(normalized_rules.keys())[:5]}")

    print("\n=== Mihomo Rule Providers (前3个) ===")
    mihomo_providers = loader.generate_mihomo_rule_providers()
    for provider_name, config in list(mihomo_providers.items())[:3]:
        print(f"\n{provider_name}:")
        print(f"  URL: {config['url']}")
        print(f"  Behavior: {config['behavior']}")
        print(f"  Format: {config['format']}")

    print("\n=== Loon Remote Rules (前3个) ===")
    loon_rules = loader.generate_loon_remote_rules()
    for rule in loon_rules[:3]:
        print(f"  {rule}")

    print("\n=== Surge Remote Rules (前6行) ===")
    surge_rules = loader.generate_surge_remote_rules()
    for line in surge_rules[:6]:
        print(f"  {line}")

    print("\n测试完成！")
