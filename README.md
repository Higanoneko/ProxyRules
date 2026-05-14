# PianCat 的代理规则仓库

此处存放我为多个代理工具编写的覆写规则，本仓库实现如下：

*  集成 [SukkaW/Surge](https://github.com/SukkaW/Surge) 、 [Cats-Team/AdRules](https://github.com/Cats-Team/AdRules) 、 [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script) 和 [dler-io/Rules](https://github.com/dler-io/Rules) 规则
*  包含多种分流策略
*  支持 Mihomo、Stash、Loon 等代理工具的覆写文件或配置文件自动生成
*  查看 [生成器使用文档](README_Dev.md) 了解如何使用生成器生成属于自己的配置文件
*  Go 生成程序现已集中放在 `program/` 目录


## 当前支持情况

| 工具 | 状态 | 说明 |
|------|------|------|
| [Mihomo](#Mihomo) | ✅ 已支持 | `.yaml` 覆写文件 和 `.js` 覆写脚本 |
| [Stash](#Stash) | ✅ 已支持 | `.yaml` 覆写文件 和 `.stoverride` 覆写文件 |
| [Loon](#Loon) | ✅ 已支持 | `.lcf` 配置文件 |
| [Surge](#Surge) | ✅ 已支持 | `.conf` 配置文件 |
| QuantumultX | ❌ 未支持 | 与其他代理工具配置差异较大，暂不支持 |


## 快速开始

### Mihomo (Clash Meta)

**覆写文件 (.yaml) ⭐ 优先使用**
  - [mihomo_config_ipv6-1_full-0.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-1_full-0.yaml) - 启用 IPv6，基础配置 ⭐ 推荐
  - [mihomo_config_ipv6-0_full-0.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-0_full-0.yaml) - 禁用 IPv6，基础配置
  - [mihomo_config_ipv6-0_full-1.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-0_full-1.yaml) - 禁用 IPv6，完整配置
  - [mihomo_config_ipv6-1_full-1.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-1_full-1.yaml) - 启用 IPv6，完整配置

**覆写脚本 (.js)**
  - [mihomo_convert_ipv6-1_full-0.js](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_ipv6-1_full-0.js) - 启用 IPv6，基础配置 ⭐ 推荐
  - [mihomo_convert_ipv6-0_full-0.js](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_ipv6-0_full-0.js) - 禁用 IPv6，基础配置
  - [mihomo_convert_ipv6-0_full-1.js](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_ipv6-0_full-1.js) - 禁用 IPv6，完整配置
  - [mihomo_convert_ipv6-1_full-1.js](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_ipv6-1_full-1.js) - 启用 IPv6，完整配置

**Box4Root 配置 (.yaml)**

用于 [Box4Root](https://github.com/boxproxy/box) 等运行在 Shell 环境中的工具。
  - [Box4Root_mihomo_config.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Box4Root/Box4Root_mihomo_config.yaml) - 启用 IPv6，禁用 TUN。 ⭐ 推荐（需要更改 Box4Root 的运行模式为 Enhanced 等非 TUN 模式）
  - [Box4Root_mihomo_config_no_ipv6.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Box4Root/Box4Root_mihomo_config_no_ipv6.yaml) - 禁用 IPv6，禁用 TUN。（需要更改 Box4Root 的运行模式为 Enhanced 等非 TUN 模式）
  - [Box4Root_mihomo_config_tun.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Box4Root/Box4Root_mihomo_config_tun.yaml) - 启用 IPv6，启用 TUN。
  - [Box4Root_mihomo_config_tun_no_ipv6.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Box4Root/Box4Root_mihomo_config_tun_no_ipv6.yaml) - 禁用 IPv6，启用 TUN。


**Sparkle/Clash Party 使用方法**

推荐使用 [Sparkle](https://github.com/xishang0128/sparkle)

1. 打开覆写
2. 在上方地址栏中粘贴上方腹泻脚本的链接
3. 点击「导入」按钮，（可选）将导入后的文件更改名称为你认为合适的名称，并且开启全局
4. 为对应的配置文件添加该覆写（如果已开启全局则不需要）

**Sparkle/Clash Party 特别设置**

需要注意，Sparkle/Clash Party 在默认设置下还会接管 DNS 和 SNI（域名嗅探），需要手动在设置中关闭「控制 DNS 设置」和「控制域名嗅探」两个选项。

**SubStore 使用方法**

  -  [mihomo_convert_args.js](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_args.js) - SubStore 可导入参数脚本

可传入参数，传入多个参数时，用`&`分隔：
* `ipv6`：是否启用 IPv6，取值 `0`（禁用）或 `1`（启用），默认值 `1`
* `full`：是否使用完整配置，取值 `0`（基础配置）或 `1`（完整配置），默认值 `0`

用例：
```
https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_convert_args.js#ipv6=1&full=0
```

**ShellCrash 使用方法**

使用 `curl` 下载配置文件到本地：

```bash
curl -o /path/to/config/mihomo_config.yaml https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-1_full-0.yaml
```

使用 `wget` 下载配置文件到本地：

```bash
wget -P /path/to/config/mihomo_config.yaml https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Mihomo/mihomo_config_ipv6-1_full-0.yaml
```
> 以上 Sample 均使用启用 IPv6 的基础配置文件作为示例，请根据需要替换为其他版本的配置文件链接。


在 [ShellCrash](https://github.com/juewuy/ShellCrash) 的 配置文件管理中选择 `本地生成配置文件(基于内核providers,推荐！)`

先选择 `选择规则模版` 修改为下载的配置文件路径，再选择 `生成配置文件` 即可生成最终配置文件。

### Stash

**覆写文件 (.stoverride) ⭐ 优先使用**
- 点击以下链接直接导入到 Stash（推荐）：
  - [一键导入 Stash_override.stoverride](https://intradeus.github.io/http-protocol-redirector?r=stash://install-override?url=https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Stash/Stash_override.stoverride) - 启用 IPv6 版本 ⭐ 推荐
  - [一键导入 Stash_override_no_ipv6.stoverride](https://intradeus.github.io/http-protocol-redirector?r=stash://install-override?url=https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Stash/Stash_override_no_ipv6.stoverride) - 禁用 IPv6 版本

**配置文件 (.yaml)**
  - [Stash_config_full.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Stash/Stash_config_full.yaml) - 启用 IPv6 版本 ⭐ 推荐
  - [Stash_config_full_no_ipv6.yaml](https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Stash/Stash_config_full_no_ipv6.yaml) - 禁用 IPv6 版本


### Surge

**Surge 模块 (.sgmodule) ⭐ 优先使用**
  - [一键导入 Surge_override.sgmodule](https://intradeus.github.io/http-protocol-redirector?r=surge:///install-module?url=https%3A%2F%2Fraw%2Egithubusercontent%2Ecom%2FPianCat%2FProxyRules%2Fmain%2FConfig%2FSurge%2FSurge_override%2Esgmodule) - 启用 IPv6 版本 ⭐ 推荐
  - [一键导入 Surge_override_no_ipv6.sgmodule](https://intradeus.github.io/http-protocol-redirector?r=surge:///install-module?url=https%2F%2Fraw%2Egithubusercontent%2Ecom%2FPianCat%2FProxyRules%2Fmain%2FConfig%2FSurge%2FSurge_override_no_ipv6%2Esgmodule) - 禁用 IPv6 版本

**配置文件 (.conf)**
  - [一键导入 Surge_config.conf](https://intradeus.github.io/http-protocol-redirector?r=surge:///install-config?url=https%3A%2F%2Fraw%2Egithubusercontent%2Ecom%2FPianCat%2FProxyRules%2Fmain%2FConfig%2FSurge%2FSurge_config%2Econf) - 启用 IPv6 版本 ⭐ 推荐
  - [一键导入 Surge_config_no_ipv6.conf](https://intradeus.github.io/http-protocol-redirector?r=surge:///install-config?url=https%3A%2F%2Fraw%2Egithubusercontent%2Ecom%2FPianCat%2FProxyRules%2Fmain%2FConfig%2FSurge%2FSurge_config_no_ipv6%2Econf) - 禁用 IPv6 版本

### Loon

**配置文件 (.lcf)**
- 点击以下链接直接导入到 Loon（推荐）：
  - [一键导入 Loon_config.lcf](https://intradeus.github.io/http-protocol-redirector?r=loon://import?sub=https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Loon/Loon_config.lcf) - 启用 IPv6 版本 ⭐ 推荐
  - [一键导入 Loon_config_no_ipv6.lcf](https://intradeus.github.io/http-protocol-redirector?r=loon://import?sub=https://raw.githubusercontent.com/PianCat/ProxyRules/main/Config/Loon/Loon_config_no_ipv6.lcf) - 禁用 IPv6 版本

**使用方法：**
1. 下载配置文件到本地
2. 在 Surge 中选择「从文件导入配置」
3. 选择下载的 `.conf` 文件
4. 在配置中添加你的代理节点订阅（替换 `policy-path=订阅地址`）


## 分流策略

本仓库包含以下分流策略组：

### 规则列表

| 策略组 | 包含的规则 |
|--------|-----------|
| AI | AI |
| Telegram | Telegram |
| YouTube | YouTube, YouTubeMusic |
| Netflix | Netflix |
| TikTok | TikTok |
| Spotify | Spotify |
| Steam | Steam |
| Game | Game, Playhorny, Nikke |
| E-Hentai | E-Hentai |
| PornSite | PornSite, Furrybar |
| US Media | Stream\_US |
| Taiwan Media | Stream\_TW |
| Japan Media | Stream\_JP |
| Global Media | Stream\_Global |
| Apple | Apple |
| Microsoft | Microsoft |
| Google | Google |
| Google FCM | GoogleFCM |
| Sogou Privacy | SogouPrivacy |
| ADBlock | ADBlock |
| 直接连接 | LocalNetwork, LocalNetworkIP |

### 节点组

| 名称 | 说明 |
|------|------|
| 香港节点 | HongKong |
| 台湾节点 | Taiwan |
| 新加坡节点 | Singapore |
| 美国节点 | Unite State |
| 日本节点 | Japan |
| 其他节点 | 其他地区节点 |

## Wireguard 配置（Easytier）

本仓库还提供了适用于 Wireguard 的 Easytier 配置文件，仅拥有适用于 JavaScript 的 Mihomo 覆写脚本版本以及 Surge Module 。

相关文件位置处于 `Wireguard_Easytier` 文件夹下，请自行查阅使用。

## 自定义规则

规则定义完全由 `Base/Rules/RemoteRules.yaml` 驱动，**无需修改 Go 代码**。

文件分为两个区域：

- **`BaseRules`**：基础规则集，每条包含完整的 `policyname`（策略组归属）、`tagname`（展示名）
- **`CustomRules`**：自定义规则，通过 `parenttag` 归入已有策略组（如 Playhorny 的 `parenttag: "Game"` 使其归入 Game）

### 新增一条规则

在 `CustomRules` 下添加：

```yaml
CustomRules:
  MyRule:
    name: "MyRule"
    category: "PianCat"              # 规则源分类，对应 RemoteRulesLinkBase.yaml
    behavior: "classical"             # domain / classical / ip
    remotefile: "./MyRule/MyRule.list" # 远程规则文件路径
    parenttag: "Game"                # 归入 Game 策略组
```

| 字段 | 必需 | 说明 |
|------|------|------|
| `name` | 是 | 规则显示名称 |
| `category` | 是 | 规则源分类，对应 `RemoteRulesLinkBase.yaml` 中的 Categories |
| `behavior` | 是 | `domain` / `classical` / `ip` |
| `remotefile` | 是 | 远程规则文件路径，拼接基础 URL 形成完整下载链接 |
| `policyname` | BaseRules 必需 | 归属的策略组名称 |
| `tagname` | 可选 | 展示标签名，默认使用 `name` |
| `parenttag` | 可选 | 父规则 RuleID，子规则继承父规则的 `policyname` |
| `surgeoption` | 可选 | Surge 专用参数（如 `extended-matching`） |

> 规则顺序与 YAML 书写顺序一致，BaseRules 在前、CustomRules 在后。

## 开发

生成器源码、`go.mod` 和内部包都位于 `program/`。

本地重新生成配置时，请在 `program/` 目录执行：

```bash
cd program
go run ./cmd/proxyrules --tool all
```

GitHub Actions 现在只保留 `auto_generate.yml`，专门负责生成并提交 `Config/` 与 `Wireguard_Easytier/`。

## 感谢

本仓库集成了以下优秀的规则源项目，感谢所有开发者的贡献：

### 规则源项目

- **[SukkaW/Surge](https://github.com/SukkaW/Surge)** - 提供 Telegram、流媒体、搜狗输入法等规则
- **[Cats-Team/AdRules](https://github.com/Cats-Team/AdRules)** - 提供广告拦截规则
- **[blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)** - 提供 YouTube、Netflix、TikTok、Spotify、Steam、游戏、Apple、Microsoft、Google 等服务规则
- **[dler-io/Rules](https://github.com/dler-io/Rules)** - 提供 AI 服务规则

### 工具与资源

- **[MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)** - 提供 GeoIP 和 IPASN 数据库

---

**注意**：使用本仓库的配置文件前，请确保你已经配置好代理节点。配置文件中的策略组需要你手动指定对应的代理节点。
