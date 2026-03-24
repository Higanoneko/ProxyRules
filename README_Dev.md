# ProxyRules 生成器开发说明

这份文档面向维护者、Fork 使用者和想基于本仓库生成“自己的代理配置集”的用户。

---

## 1. 先理解仓库结构

本项目的生成流程可以理解为：

```text
Base/            -> 原始配置、模板、规则定义
program/         -> Go 生成器
Config/          -> 生成产物
Wireguard_Easytier/ -> Easytier_Wireguard 相关产物
```

最重要的目录如下：

- `Base/`
  - `DNS.yaml`：DNS 基础数据
  - `Ports.yaml`：端口配置
  - `Fake_IP_Filter.yaml`：Fake-IP 相关列表
  - `Test_URL.yaml`：连通性测试地址
  - `Head/`：各平台代理工具的头模板
  - `Rules/RemoteRules.yaml`：规则条目定义
  - `Rules/RemoteRulesLinkBase.yaml`：规则源 URL 模板与工具映射
- `program/`
  - `cmd/proxyrules/`：CLI 入口
  - `internal/service/`：生成流程
  - `internal/render/`：各工具渲染器
- `Config/`
  - 生成后的 Mihomo / Stash / Loon / Surge / Box4Root 配置
- `Wireguard_Easytier/`
  - Easytier 相关脚本与模块输出

---

## 2. 这个生成器生成的到底是什么

这个程序生成的是“规则框架 + DNS 策略 + 策略组 + 规则源引用 + 平台模板整合”后的配置集。

它**不是**节点订阅管理器，也**不会**替你保存真实机场订阅。

通常的使用方式是：

- Mihomo / Stash：把生成出的覆写文件或脚本导入到客户端，再让客户端加载你自己的订阅
- Loon / Surge：生成带策略组和规则的配置骨架，再接入你自己的节点来源
- Box4Root：生成可直接修改订阅地址的完整模板文件

如果你只是想做一套“自己的规则与模板风格”，一般只需要改 `Base/` 下面的文件，不需要改 Go 代码。

---

## 3. 环境要求

- Go `1.23` 或更高版本
- Windows / macOS / Linux 均可，只要能运行 Go

检查版本：

```bash
go version
```

---

## 4. 常用命令

先进入生成器目录：

```bash
cd program
```

生成全部配置：

```bash
go run ./cmd/proxyrules --tool all
```

只生成 Mihomo 系：

```bash
go run ./cmd/proxyrules --tool mihomo
```

> `--tool mihomo` 会同时生成：
> - `Config/Mihomo/`
> - `Config/Box4Root/`

只生成 Stash：

```bash
go run ./cmd/proxyrules --tool stash
```

一次生成多个目标：

```bash
go run ./cmd/proxyrules --tool mihomo,stash,surge
```

输出到自定义目录：

```bash
go run ./cmd/proxyrules --tool all --output D:\Temp\ProxyRulesOutput
```

用测试节点预览国家分组效果：

```bash
go run ./cmd/proxyrules --tool mihomo --test
```

> `--test` 不会读取你的真实订阅，只会使用内置测试节点名来验证国家分组、策略组和脚本逻辑。

运行测试：

```bash
go test ./...
```

---

## 5. 各目标会生成什么

### `--tool mihomo`

输出到：

- `Config/Mihomo/`
- `Config/Box4Root/`

会生成：

- `mihomo_config_ipv6-1_full-0.yaml`
- `mihomo_config_ipv6-1_full-1.yaml`
- `mihomo_config_ipv6-0_full-0.yaml`
- `mihomo_config_ipv6-0_full-1.yaml`
- `mihomo_convert_args.js`
- `mihomo_convert_ipv6-1_full-0.js`
- `mihomo_convert_ipv6-1_full-1.js`
- `mihomo_convert_ipv6-0_full-0.js`
- `mihomo_convert_ipv6-0_full-1.js`
- `Box4Root_mihomo_config.yaml`
- `Box4Root_mihomo_config_no_ipv6.yaml`
- `Box4Root_mihomo_config_tun.yaml`
- `Box4Root_mihomo_config_tun_no_ipv6.yaml`

说明：

- `ipv6-1` 表示启用 IPv6
- `ipv6-0` 表示禁用 IPv6
- `full-1` 表示完整配置
- `full-0` 表示基础配置
- `Box4Root_mihomo_config_tun*` 表示 `tun.enable = true`
- `Box4Root_mihomo_config*` 表示 `tun.enable = false`

### `--tool stash`

输出到 `Config/Stash/`：

- `Stash_config_full.yaml`
- `Stash_config_full_no_ipv6.yaml`
- `Stash_override.stoverride`
- `Stash_override_no_ipv6.stoverride`

### `--tool loon`

输出到 `Config/Loon/`：

- `Loon_config.lcf`
- `Loon_config_no_ipv6.lcf`

### `--tool surge`

输出到 `Config/Surge/`：

- `Surge_config.conf`
- `Surge_config_no_ipv6.conf`

### `--tool easytier`

输出到：

- 默认输出根目录是 `Config` 时：`Wireguard_Easytier/Mihomo/`
- 自定义输出根目录时：`<你的输出目录>/Wireguard_Easytier/Mihomo/`

额外说明：

- 每次运行生成器时，都会同步刷新 `Wireguard_Easytier/Surge/Easytier.sgmodule`
- 当选择 `--tool easytier` 时，还会额外生成 Mihomo 版 Easytier JS 产物

注意：

- `easytier` 依赖输出根目录下 `Mihomo/` 中已经生成好的 `mihomo_convert_*.js`
- 最稳妥的用法是直接执行：

```bash
go run ./cmd/proxyrules --tool all
```

或者：

```bash
go run ./cmd/proxyrules --tool mihomo,easytier
```

---

## 6. 你应该改哪些文件

### 6.1 改 DNS：`Base/DNS.yaml`

当前推荐结构：

```yaml
DNS:
  IP:
    - 119.29.29.29
    - '2402:4e00::'
  Default_DoH:
    - https://doh.pub/dns-query
    - https://cloudflare-dns.com/dns-query
  Proxy_Server_DoH:
    - https://doh.pub/dns-query
  Direct_DoH:
    - https://doh.pub/dns-query
  Fallback_DoH:
    - https://cloudflare-dns.com/dns-query
    - https://dns.google/dns-query
```

当前生成器中的主要含义：

- `IP`：用于 `default-nameserver`
- `Default_DoH`：用于主 `nameserver`
- `Proxy_Server_DoH`：用于 Mihomo / Stash / Box4Root 的 `proxy-server-nameserver`
- `Direct_DoH`：用于 `direct-nameserver`
- `Fallback_DoH`：保留在通用 DNS 模型中，作为补充上游池

如果你只想改 DNS，通常只需要改这一个文件，再重新生成。

### 6.2 改 Fake-IP / Always-Real-IP：`Base/Fake_IP_Filter.yaml`

这里维护：

- `Fake_IP_Filter`
- `Surge_Always_Real_IP`

影响：

- Mihomo / Stash / Box4Root 的 `fake-ip-filter`
- Loon / Surge 的真实 IP 相关配置

### 6.3 改端口：`Base/Ports.yaml`

这里控制：

- Mihomo `mixed-port`
- 通用 HTTP / SOCKS5 端口

### 6.4 改测试地址：`Base/Test_URL.yaml`

这里控制：

- `internet-test-url`
- `proxy-test-url`

用于 Loon / Surge 等平台的连通性测试设置。

### 6.5 改头模板：`Base/Head/`

这是最常改的地方。

当前模板文件：

- `Base/Head/Head_Mihomo.yaml`
- `Base/Head/Head_Box4Root.yaml`
- `Base/Head/Head_Stash.yaml`
- `Base/Head/Head_Loon.conf`
- `Base/Head/Head_Surge.conf`

适合在这里改的内容：

- DNS 块里的固定字段
- 日志级别
- `sniffer`
- `tun`
- `profile`
- UI 设置
- 默认注释

如果你有自己的固定风格，优先改头模板，而不是每次手动改生成结果。

### 6.6 改规则内容：`Base/Rules/RemoteRules.yaml`

这里定义“有哪些规则组”“每个规则组的名字是什么”“对应哪个规则源分类”“行为类型是什么”“远程文件路径是什么”。

适合做的事：

- 新增规则组
- 删除不需要的规则组
- 调整规则名称
- 更换某个规则组对应的远程规则文件

### 6.7 改规则源映射：`Base/Rules/RemoteRulesLinkBase.yaml`

这里定义：

- 每个规则分类的 URL 模板
- 不同平台代理工具在不同规则源中的工具名映射
- 文件扩展名映射

只有当你要接入新的规则源，或者现有规则源的 URL 模板变化时，才需要改这里。

---

## 7. 模板占位符怎么用

### 7.1 Mihomo 系 YAML 占位符

在以下 YAML 模板中使用：

- `Base/Head/Head_Mihomo.yaml`
- `Base/Head/Head_Box4Root.yaml`
- `Base/Head/Head_Stash.yaml`

推荐写法：

```yaml
default-nameserver: "$DNS_IP_List"
direct-nameserver: "$DNS_Direct_DoH_List"
proxy-server-nameserver: "$DNS_Proxy_Server_DoH_List"
nameserver: "$DNS_Default_DoH_List"
fake-ip-filter: "$Fake_IP_Filter_List"
```

当前常用占位符：

- `"$DNS_IP_List"`
- `"$DNS_Default_DoH_List"`
- `"$DNS_Direct_DoH_List"`
- `"$DNS_Proxy_Server_DoH_List"`
- `"$Fake_IP_Filter_List"`

这些 DNS 占位符不仅会影响 Mihomo / Stash / Box4Root 的 YAML 输出，也会同时影响 Mihomo 的 `mihomo_convert_*.js` 与 `mihomo_convert_args.js`，因为 JS 的 DNS 模板现在与 Mihomo 头模板共用同一套 DNS 合成结果。

以下是在 `rules` 字段上的额外占位符 `"$ProxyRules_Pack"`，当前标准写法是：

```yaml
rules:
  - YOUR,RULES_1,HERE
  - "$ProxyRules_Pack"
  - YOUR,RULES_2,HERE
```

作用：

- 生成器会把完整规则包插入到这个占位符所在的位置
- 这样你可以精确控制自定义规则放在规则包前面还是后面

用例：

```yaml
rules:
  - DST-PORT,53,DNS_Hijack
  - "$ProxyRules_Pack"
```

这会让 `DNS_Hijack` 永远位于规则组最上面。

如果 `rules` 中没有 `"$ProxyRules_Pack"`，生成器会回退到“普通生成模式”，直接使用默认生成的规则包，不会把头模板里的自定义 `rules` 与规则包做位置混排。

### 7.3 文本头模板占位符

在以下文本模板中使用：

- `Base/Head/Head_Loon.conf`
- `Base/Head/Head_Surge.conf`

当前文本头模板同时支持两种写法：

```text
{DNS_IP_list}
{DNS_Default_DoH_List}
{DNS_DoH_list}
{Fake_IP_Filter_list}
{Surge_Always_Real_IP_List}
"$DNS_IP_list"
"$DNS_Default_DoH_List"
"$Fake_IP_Filter_list"
"$Surge_Always_Real_IP_List"
```

例如：

```ini
dns-server = "$DNS_IP_list",system
doh-server = "$DNS_Default_DoH_List"
real-ip = "$Fake_IP_Filter_list"
```

常用文本占位符：

- `"$DNS_IP_list"`
- `"$DNS_Default_DoH_List"`
- `"$DNS_Proxy_Server_DoH_List"`
- `"$DNS_Direct_DoH_List"`
- `"$Fake_IP_Filter_list"`
- `"$Surge_Always_Real_IP_List"`

---

## 8. 一个推荐的定制流程

如果你是第一次做自己的配置集，建议按这个顺序：

### 第一步：先定 DNS

编辑：

- `Base/DNS.yaml`
- `Base/Fake_IP_Filter.yaml`

### 第二步：再定模板风格

编辑：

- `Base/Head/Head_Mihomo.yaml`
- `Base/Head/Head_Box4Root.yaml`
- `Base/Head/Head_Stash.yaml`
- `Base/Head/Head_Loon.conf`
- `Base/Head/Head_Surge.conf`

### 第三步：最后定规则集

编辑：

- `Base/Rules/RemoteRules.yaml`
- `Base/Rules/RemoteRulesLinkBase.yaml`

### 第四步：生成并检查

```bash
cd program
go test ./...
go run ./cmd/proxyrules --tool all
```

重点检查：

- `Config/Mihomo/`
- `Config/Box4Root/`
- `Config/Stash/`
- `Config/Loon/`
- `Config/Surge/`

---

## 9. 几个典型场景

### 场景 A：我只想改 DNS，不想动别的

改：

- `Base/DNS.yaml`
- `Base/Fake_IP_Filter.yaml`

然后执行：

```bash
cd program
go run ./cmd/proxyrules --tool mihomo,stash,loon,surge
```

### 场景 B：我只想改 Mihomo / Box4Root 头部模板

改：

- `Base/Head/Head_Mihomo.yaml`
- `Base/Head/Head_Box4Root.yaml`

然后执行：

```bash
cd program
go run ./cmd/proxyrules --tool mihomo
```

### 场景 C：我只想新增一个规则组

改：

- `Base/Rules/RemoteRules.yaml`

如果新规则来自全新规则源，还要同时改：

- `Base/Rules/RemoteRulesLinkBase.yaml`

然后执行：

```bash
cd program
go run ./cmd/proxyrules --tool all
```

### 场景 D：我想预览国家分组是否合理

执行：

```bash
cd program
go run ./cmd/proxyrules --tool mihomo --test
```

---

## 10. 关于 Mihomo JS 参数脚本

`Config/Mihomo/mihomo_convert_args.js` 是可调参版本。

需要注意：

- 这个 JS 的 DNS 块并不是一套独立配置
- 它会继承 Mihomo 头模板中的 DNS 合成结果
- 因此你在 `Base/Head/Head_Mihomo.yaml` 中使用的 `"$DNS_IP_List"`、`"$DNS_Default_DoH_List"`、`"$DNS_Proxy_Server_DoH_List"`、`"$DNS_Direct_DoH_List"`、`"$Fake_IP_Filter_List"` 也会同步影响 JS 输出

目前支持的参数有：

- `ipv6=0|1`
- `full=0|1`
- `threshold=<数字>`

其中：

- `ipv6`：是否启用 IPv6
- `full`：是否使用完整配置
- `threshold`：某国家节点数量小于该值时，不显示该国家分组

示例：

```text
https://example.com/mihomo_convert_args.js#ipv6=1&full=0&threshold=2
```

---

## 11. 常见坑

### 11.1 `--tool mihomo` 为什么还生成了 Box4Root

这是设计如此。

因为 Box4Root 使用 Mihomo 系渲染链路，所以它跟随 Mihomo 一起生成。

### 11.2 为什么我改了头模板，某些块又被程序覆盖了

因为生成器对部分区域是“程序负责生成”的：

- `proxy-groups`
- `rule-providers`
- `rules`

Mihomo / Stash / Box4Root 的 DNS 也会由程序把动态值写进去，但头模板中不属于动态字段的固定内容会被保留。

### 11.3 为什么 Easytier 单独生成时报错

因为 Easytier 需要先读取 `Mihomo` 目录下已生成的 `mihomo_convert_*.js`。

请先执行：

```bash
go run ./cmd/proxyrules --tool mihomo,easytier
```

或直接：

```bash
go run ./cmd/proxyrules --tool all
```

### 11.4 为什么生成器没有读取我的机场节点

因为这个程序不是订阅下载器。

它生成的是配置框架，真实节点一般由：

- 客户端的远程订阅
- `proxy-providers`
- 覆写脚本运行时传入的代理列表

来提供。

---

## 12. 建议的维护习惯

- 先改 `Base/`，再生成，不要直接手改 `Config/` 后忘记同步
- 改 DNS 或模板后，至少跑一次 `go run ./cmd/proxyrules --tool all`
- 改渲染逻辑后，先跑 `go test ./...`
- 如果仓库公开，尽量不要把私人订阅 URL 直接提交到 `Base/Head/Head_Box4Root.yaml`

---

## 13. 最短上手路线

如果你只想最快做出自己的版本：

1. Fork 本仓库
2. 修改 `Base/DNS.yaml`
3. 修改 `Base/Head/` 下你关心的平台模板
4. 如有需要，修改 `Base/Rules/RemoteRules.yaml`
5. 执行：

```bash
cd program
go run ./cmd/proxyrules --tool all
```

6. 从 `Config/` 和 `Wireguard_Easytier/` 拿走生成结果

做到这一步，你就已经拥有一套可持续维护、可重复生成的“自己的代理配置集”了。
