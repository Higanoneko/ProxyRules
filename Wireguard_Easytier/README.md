# Higanoneko 的 Easytier WireGuard 配置

此处存放为 [Easytier](https://github.com/EasyTier/EasyTier) WireGuard 组网准备的附属配置，作用如下：

* 将 Easytier WireGuard 节点添加到 Mihomo 或 Surge
* 将指定的 Easytier 虚拟网段路由到该节点
* 不管理 DNS、域名嗅探或主配置的分流规则

## 当前支持情况

| 工具 | 状态 | 说明 |
|------|------|------|
| [Mihomo](#mihomo-clash-meta) | ✅ 已支持 | 在订阅转换脚本中追加 Easytier 节点与路由 |
| [Surge](#surge) | ✅ 已支持 | `.sgmodule` 模块 |

## 快速开始

### Mihomo (Clash Meta)

**Easytier 覆写脚本 (.js)**

以下脚本会在原有 Mihomo 订阅转换结果中追加 Easytier WireGuard 节点和路由：

  - [mihomo_convert_args.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_args.js) - 可传入参数的版本 ⭐ 推荐
  - [mihomo_convert_ipv6-1_dns-1_full-0.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-1_dns-1_full-0.js) - 启用 IPv6，启用 DNS，基础配置
  - [mihomo_convert_ipv6-1_dns-1_full-1.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-1_dns-1_full-1.js) - 启用 IPv6，启用 DNS，完整配置
  - [mihomo_convert_ipv6-0_dns-1_full-0.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-0_dns-1_full-0.js) - 禁用 IPv6，启用 DNS，基础配置
  - [mihomo_convert_ipv6-0_dns-1_full-1.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-0_dns-1_full-1.js) - 禁用 IPv6，启用 DNS，完整配置
  - [mihomo_convert_ipv6-1_dns-0_full-0.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-1_dns-0_full-0.js) - 启用 IPv6，不配置 DNS 或域名嗅探，基础配置
  - [mihomo_convert_ipv6-1_dns-0_full-1.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-1_dns-0_full-1.js) - 启用 IPv6，不配置 DNS 或域名嗅探，完整配置
  - [mihomo_convert_ipv6-0_dns-0_full-0.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-0_dns-0_full-0.js) - 禁用 IPv6，不配置 DNS 或域名嗅探，基础配置
  - [mihomo_convert_ipv6-0_dns-0_full-1.js](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Mihomo/mihomo_convert_ipv6-0_dns-0_full-1.js) - 禁用 IPv6，不配置 DNS 或域名嗅探，完整配置

导入方式与主目录 README 中的 Mihomo 覆写脚本一致。`--tool easytier` 依赖已生成的 `Config/Mihomo/mihomo_convert_*.js`；建议使用 `--tool mihomo,easytier` 或 `--tool all` 一并生成。

### Surge

**Easytier 模块 (.sgmodule)**

  - [Easytier.sgmodule](https://raw.githubusercontent.com/Higanoneko/ProxyRules/main/Wireguard_Easytier/Surge/Easytier.sgmodule) - 为 Surge 添加 Easytier WireGuard 节点和相关网段路由

在 Surge 的「模块」页面导入上方链接，并按下方说明完成 WireGuard 参数配置。

## WireGuard 参数配置

请先修改 `Base/Wireguard/Easytier/` 中的源文件，再重新生成产物；不要直接编辑本目录的 `.js` 或 `.sgmodule` 文件。

| 平台 | 源文件 | 需要填写的内容 |
|------|--------|----------------|
| Mihomo | `Easytier.js` | Endpoint、端口、客户端 Address、服务端公钥、客户端私钥、Easytier 网段 |
| Surge | `Easytier.sgmodule` | Endpoint、客户端 Address、服务端公钥、客户端私钥、Easytier 网段 |

配置完成后，在项目的 `program/` 目录运行：

```powershell
go run ./cmd/proxyrules --tool mihomo,easytier
```

> 本目录除 `README.md` 外均为生成产物。请通过修改 `Base/` 或 `program/` 后重新生成，不要手工修改生成结果。
