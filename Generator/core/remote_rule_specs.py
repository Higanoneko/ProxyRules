"""
RemoteRules 统一规则规格定义。

该模块提供：
1. 统一的规则 ID 与策略映射（避免各生成器各自维护一份）；
2. 各目标平台的规则输出顺序；
3. 历史键名兼容别名。
"""

from dataclasses import dataclass
from typing import Dict, List


@dataclass(frozen=True)
class RemoteRuleSpec:
    """统一规则规格。"""

    rule_id: str
    policy_name: str
    tag_name: str
    mihomo_policy_name: str | None = None
    surge_option: str = ""


REMOTE_RULE_SPECS: List[RemoteRuleSpec] = [
    RemoteRuleSpec("AI", "AI", "AI"),
    RemoteRuleSpec("Telegram", "Telegram", "Telegram"),
    RemoteRuleSpec("YouTube", "YouTube", "YouTube"),
    RemoteRuleSpec("YouTubeMusic", "YouTube", "YouTube Music"),
    RemoteRuleSpec("Netflix", "Netflix", "Netflix"),
    RemoteRuleSpec("TikTok", "TikTok", "TikTok"),
    RemoteRuleSpec("Spotify", "Spotify", "Spotify"),
    RemoteRuleSpec("Steam", "Steam", "Steam"),
    RemoteRuleSpec("Game", "Game", "Game"),
    RemoteRuleSpec("E-Hentai", "E-Hentai", "E-Hentai"),
    RemoteRuleSpec("PornSite", "PornSite", "PornSite"),
    RemoteRuleSpec("Furrybar", "PornSite", "Furrybar"),
    RemoteRuleSpec("Stream_US", "US Media", "US Media"),
    RemoteRuleSpec("Stream_TW", "Taiwan Media", "Taiwan Media"),
    RemoteRuleSpec("Playhorny", "Taiwan Media", "Playhorny"),
    RemoteRuleSpec("Stream_JP", "Japan Media", "Japan Media"),
    RemoteRuleSpec("Stream_Global", "Global Media", "Global Media"),
    RemoteRuleSpec("Apple", "Apple", "Apple"),
    RemoteRuleSpec("Microsoft", "Microsoft", "Microsoft"),
    RemoteRuleSpec("Google", "Google", "Google"),
    RemoteRuleSpec("GoogleFCM", "Google FCM", "Google FCM"),
    RemoteRuleSpec("SogouPrivacy", "Sogou Privacy", "Sogou Privacy"),
    RemoteRuleSpec("ADBlock", "ADBlock", "ADBlock", surge_option="extended-matching"),
    RemoteRuleSpec("LocalNetwork", "DIRECT", "LocalNetwork", mihomo_policy_name="DIRECT"),
    RemoteRuleSpec("LocalNetworkIP", "DIRECT", "LocalNetworkIP", mihomo_policy_name="DIRECT"),
]


REMOTE_RULE_SPEC_INDEX: Dict[str, RemoteRuleSpec] = {
    spec.rule_id: spec for spec in REMOTE_RULE_SPECS
}


# 历史键名兼容映射。
# 基于当前 RemoteRules.yaml（2026-02-09）已不需要额外别名，保留空映射以便未来按需扩展。
REMOTE_RULE_ID_ALIASES: Dict[str, str] = {}


# Mihomo / Stash（Clash 语法）输出顺序
MIHOMO_RULE_ORDER: List[str] = [
    "AI",
    "Telegram",
    "YouTube",
    "YouTubeMusic",
    "Netflix",
    "TikTok",
    "Spotify",
    "Steam",
    "Game",
    "E-Hentai",
    "PornSite",
    "Furrybar",
    "Stream_US",
    "Stream_TW",
    "Playhorny",
    "Stream_JP",
    "Stream_Global",
    "Apple",
    "Microsoft",
    "Google",
    "GoogleFCM",
    "SogouPrivacy",
    "ADBlock",
    "LocalNetwork",
    "LocalNetworkIP",
]


# Loon 输出顺序
LOON_RULE_ORDER: List[str] = [
    "AI",
    "Telegram",
    "YouTube",
    "YouTubeMusic",
    "Netflix",
    "TikTok",
    "Spotify",
    "Steam",
    "Game",
    "E-Hentai",
    "PornSite",
    "Furrybar",
    "Stream_US",
    "Stream_TW",
    "Playhorny",
    "Stream_JP",
    "Stream_Global",
    "Apple",
    "Microsoft",
    "Google",
    "GoogleFCM",
    "SogouPrivacy",
    "ADBlock",
    "LocalNetwork",
    "LocalNetworkIP",
]


# Surge 输出顺序（ADBlock 前置）
SURGE_RULE_ORDER: List[str] = [
    "ADBlock",
    "AI",
    "Telegram",
    "YouTube",
    "YouTubeMusic",
    "Netflix",
    "TikTok",
    "Spotify",
    "Steam",
    "Game",
    "E-Hentai",
    "PornSite",
    "Furrybar",
    "Stream_US",
    "Stream_TW",
    "Playhorny",
    "Stream_JP",
    "Stream_Global",
    "Apple",
    "Microsoft",
    "Google",
    "GoogleFCM",
    "SogouPrivacy",
    "LocalNetwork",
    "LocalNetworkIP",
]
