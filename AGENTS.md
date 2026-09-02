# AGENTS.md

## 项目简介

ProxyRules 是一个用 Go 编写的代理配置生成器：读取 `Base/` 下的数据与头模板，生成 Mihomo / Stash / Loon / Surge / Box4Root 以及 Easytier 相关配置。

本仓库的核心关系：

- `Base/` 是唯一数据源（DNS、端口、Fake-IP、测试 URL、各平台头模板、规则定义）
- `program/` 是 Go 生成器代码
- `Config/` 是生成产物（主要产物，已提交入库，但禁止手改）
- `Wireguard_Easytier/` 是生成产物（附属产物，同样禁止手改）；其中 `Wireguard_Easytier/README.md` 是人工维护的使用文档，不属于生成产物

## 开发主旨（按优先级排序）

### 1. 严格遵守 Functional Programming（函数式编程）

这是本项目的第一开发主旨，任何新代码、重构和修复都必须先满足函数式编程约束：

- **纯函数优先**：相同输入必须产生相同输出。业务逻辑中禁止依赖隐藏状态、全局变量、环境变量或当前时间。
- **不可变数据**：禁止修改入参的 slice / map / 结构体。需要变更时先复制（copy-on-write），再返回新值。现有代码中的 `cloneStrings`、`append([]string(nil), ...)` 就是标准写法。
- **副作用隔离**：I/O（读文件、写文件、网络、时间）只允许出现在边界层：`internal/repository`（读取 Base）、`internal/service/output_writer.go`（写产物）、`cmd/proxyrules`（CLI）。`domain`、`catalog`、`render`、`policy_plan_builder` 等业务层保持纯函数。
- **显式数据流**：依赖通过参数传入，结果通过返回值传递；禁止捕获可变状态的闭包、包级可变变量、单例或隐式全局配置。
- **组合优于继承**：用小函数组合完成流程，不做带隐式状态的"对象式"设计；结构体只承载数据。
- **错误即返回值**：函数返回 `(T, error)`，不 panic；panic 仅允许在 `main` 的退出路径和测试辅助函数中。
- **确定性输出**：除产物头部的生成时间戳（由 `output_writer` 在边界注入）外，相同 Base + 相同参数必须生成完全一致的内容。

允许的包级变量只有不可变常量类数据：模板字符串（`go:embed`）、编译期正则、目录/键名列表、只读目录数据（`catalog`）。

### 2. 产物纪律：只生成，不手改

- `Config/` 与 `Wireguard_Easytier/` 是程序输出，任何改动必须通过修改 `Base/` 或 `program/` 后重新生成完成。
- 禁止直接编辑 `Config/` 或 `Wireguard_Easytier/` 下的产物；唯一例外是 `Wireguard_Easytier/README.md`。生成的头部标记 `Generated at (UTC): ...` 不可手工伪造。
- CI（`.github/workflows/auto_generate.yml`）在 `Base/**` 或 `program/**` 变更后会重新生成并自动提交产物。

### 3. 数据优先，代码最小化

- 新增/删除/调整规则组、DNS、端口、Fake-IP、头模板等需求，优先改 `Base/`，不需要写 Go 代码。
- 只有当现有模型或渲染器无法表达需求时，才修改 `program/`。
- 规则定义唯一数据源是 `Base/Rules/RemoteRules.yaml`（`BaseRules` / `CustomRules`），自定义规则通过 `parenttag` 继承策略组。

## 目录结构

```text
Base/
  DNS.yaml                     # DNS 基础数据
  Ports.yaml                   # 端口配置
  Fake_IP_Filter.yaml          # Fake-IP / Surge real-ip 列表
  Test_URL.yaml                # 连通性测试地址
  Head/                        # 各平台头模板（含占位符）
  Rules/
    RemoteRules.yaml           # 规则条目定义（唯一数据源）
    RemoteRulesLinkBase.yaml   # 规则源 URL 模板与工具映射
program/
  cmd/proxyrules/              # CLI 入口（参数解析、边界调用）
  internal/domain/             # 领域模型（纯数据）
  internal/repository/         # 读取 Base（I/O 边界）
  internal/catalog/            # 国家/策略组模板（只读目录）
  internal/service/            # 流程编排与策略计划构建
  internal/render/             # 各平台渲染器（纯函数）
  internal/postprocess/        # Easytier 产物合并
  internal/projectroot/        # 项目根目录定位
Config/                        # 生成产物（禁止手改）
Wireguard_Easytier/            # Easytier 生成产物（禁止手改）
  README.md                    # Easytier/WireGuard 使用文档（可手改）
Temp/                          # 临时目录（已 gitignore）
.github/workflows/             # CI 自动生成
```

## 数据流

```text
Base/ ──read──> repository ──> domain（纯数据）
                                 │
                    service/policy_plan_builder（纯函数）
                                 │
                    render/*（纯函数，按平台渲染）
                                 │
                    postprocess（Easytier 合并）
                                 │
                    output_writer（唯一写产物边界）
                                 │
                                 v
                     Config/ 与 Wireguard_Easytier/
```

## 硬性规则

- 业务代码（`domain` / `catalog` / `render` / `service` 中的计算部分）禁止出现 `os.*`、`time.Now()`、网络调用、全局可变状态。
- 禁止修改函数入参中的 slice / map；跨层传递数据时使用防御性复制。
- 新增平台渲染器必须复用 `internal/render` 的公共节点工具（`yaml_nodes.go`、`RuleResolver` 等），不重复实现规则解析。
- 占位符语义保持兼容：YAML 头模板使用 `"$XXX"`，文本头模板使用 `{XXX}` / `"$XXX"`；`"$ProxyRules_Pack"` 是规则包插入锚点。
- 保持小函数、显式参数、显式错误处理；不做过度抽象，不为"未来可能"提前设计。
- 保持最小外部依赖：目前仅允许 `gopkg.in/yaml.v3`。新增依赖必须先说明理由。

## 常用命令

在 `program/` 目录下执行：

```powershell
go run ./cmd/proxyrules --tool all
go run ./cmd/proxyrules --tool mihomo,stash,loon,surge,easytier
go run ./cmd/proxyrules --tool mihomo --test     # 用测试节点预览分组
go run ./cmd/proxyrules --tool all --output D:\Temp\ProxyRulesOutput
go test ./...
```

注意：`--tool mihomo` 会同时生成 `Config/Mihomo/` 与 `Config/Box4Root/`；`--tool easytier` 依赖 `Mihomo/` 下已生成的 `mihomo_convert_*.js`，建议直接使用 `--tool all` 或 `--tool mihomo,easytier`。

## 完成改动前的检查清单

- 修改 `program/` 后：`go test ./...` 必须通过。
- 修改 `Base/` 后：运行生成器验证产物，且不手改产物。
- 确认没有把产物修改混入业务改动；产物由 CI 统一提交。
- 确认新代码满足函数式编程约束（纯函数、不可变、副作用隔离、显式数据流）。
