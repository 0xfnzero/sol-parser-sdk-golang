# 版本记录

## v0.9.0 — 2026-10-22

同步 SDK、examples、版本导航及 AI 发布规则。请阅读 [发布与迁移说明](RELEASE_NOTES.md)，尤其注意事件精简、V3 默认行为和 Go /v9 模块迁移。

以下保留旧版 README 的发布与集成记录，其中接口、依赖和验证结果描述当时版本，不代表当前行为。当前用法见 [README](README_CN.md) 和 [发布与迁移说明](RELEASE_NOTES.md)。

## v0.6.0

详见 [发布与迁移说明](RELEASE_NOTES.md)。

## v0.5.17 — Shred 元数据透传与顺序范围

保留并验证 `tx_index`、接收时间从解析入口传到原始指令回调。明确 Entries 消息内序号与 Yellowstone 节点 index 的区别：本地序号不是全 slot 索引；跨消息去重使用交易签名。此版本完善文档和回归保护，不改变现有字段或编号规则。

## v0.5.17 — RawInstruction 协议过滤与 DEX 枚举

修复原始指令绕过 DEX 订阅范围的问题。按单条指令引用的官方程序识别 DEX，使用 `RawInstructionDexType` 枚举；JSON 字符串保持不变。未知 router 不推断池子地址、买卖方向或方法版本。

## v0.5.17 — Pump 执行事件与最新回放示例

修复官方 Anchor CPI TradeEvent 前缀识别，保留历史编码兼容。RPC/gRPC 的实际 `sol_amount`、`quote_amount`、`token_amount` 只从执行事件合并；预算和最小输出保留在指令参数字段。如果日志及事件 CPI 均缺失，实际金额字段为 0（表示无执行数据，不代表已确认零成交）。非 SOL 报价金额不能当作 SOL。

在本仓库运行 `go run ./examples/pump_execution_replay`，离线回放 2026-10-19 重新抓取的真实交易。示例输出完整日志、无日志、截断日志三种结果；命令末尾追加最新签名可在线验证。通过 `SOLANA_RPC_URL` 选择 RPC，支持 transaction v1。

新 Delta 卖出样本：`4ARos3Yq1SMNE3izDpy9YdKrRgMmD7UCDxvnVrBRaGCJHiifPHSmQWc4DPfjUaEUcxS7puGmEN64K7EZdeXLJaj2`，slot 458542546。日期样本文件还包含 Rapid 创建后买入、BuyV3 和 PumpSwap 买入的完整签名及 slot。旧的五种双跳组合仅作为回归样本，未冒充本次新抓取交易。

## v0.5.17 — Shred 原始指令透传

默认 Shred 入口将未适配的第三方程序（包括 Rapid/Delta）作为 `RawInstruction` 返回，不再推断买卖方向、金额或内部 buy 版本。System、Token、ATA、ComputeBudget 等基础程序默认不透传；已支持的官方 DEX 指令继续正常解析，官方 Pump/PumpSwap 多跳仍返回 `PumpRouteIntent`。

事件包含 `program_id`、`instruction_index`、`accounts`、`account_indices`、`instruction_data` 和交易 metadata。账户保留原顺序与重复项；未解析 ALT 地址为 null/None，原始索引仍保留。使用缓存 loaded-addresses/resolver 补齐，热路径不请求 RPC。data 是完整原始字节，不要求已知布局。

运行 `go run ./examples/raw_instruction_filter` 对比同一组已签名交易的 PumpFun、Raydium CPMM 与全量原始指令订阅；输出包含 `dexes`。

原始事件保留调用方传入的 `metadata.tx_index` 和 `metadata.grpc_recv_us`（streamer 对应 `metadata.recv_us`）。默认 Entries 客户端的交易序号仅在单条接收消息内递增，下一条消息从 0 重新计数，不等于已确认的整个 slot 交易序号。`instruction_index` 表示交易内外层指令顺序；接收时间是本地微秒时间戳，不保证链上执行顺序。独立回放 API 可能默认填 0，调用时应传入上游已有的序号与接收时间。

无过滤器时返回未知第三方指令。宽泛交易/Swap 订阅（如 `PumpFunTrade`、`RaydiumCpmmSwap`）也返回关联该 DEX 的原始指令；买/卖方向订阅不推测方向。显式 `RawInstruction` 仅在未限制协议时返回全部外层指令，streamer 的协议限制始终生效；显式排除 `RawInstruction` 优先。`dex_type` 使用公开枚举 `RawInstructionDexType`（Go 为命名字符串类型与常量）；JSON 仍为 `"pumpfun"`、`"raydium_cpmm"` 等原有字符串。`dexes` 数组包含各关联 DEX 的 `{ dex_type, program_id, pool }`，只识别本条指令直接调用或账户中引用的 DEX 程序，不借用同笔其他指令的账户；多 DEX 包装保留所有关联。未知或未解析 ALT 中的程序不能匹配 DEX 订阅。`pool` 当前为 null：未知包装布局无法确定池子账户。program 引用不证明执行了 CPI，原始回调不是成交。

已有 wrapper 适配函数仅作为显式调用的兼容接口保留，不参与默认 Shred 解码。以下已发布版本的历史说明描述旧行为。

## v0.5.14 — Pump 路由方向与指令名称

路由事件沿用 `is_buy`、`ix_name` 命名。RapidBuy 为买入意图，指令名为官方 IDL 的 `buy_exact_quote_in_v2`；BuyPumpDelta 仅确定买入方向。多跳路由方向和无法确定的方法保持为空。币种读取 `hops[].base_mint` / `quote_mint`。这些字段不代表已成交；过滤器不会把路由事件转换为交易事件。

旧 JSON 缺少新字段时保持未知。Rust 结构体字面量须补齐新字段；Go 新字段为指针，Node.js/Python 未知值为 null/None。Streamer 新增 `DexEvent::PumpRouteIntentEvent` 并保留路由字段，支持协议与事件过滤。

## v0.5.13 — Pump routes and offline replay

Adds native Pump route-intent parsing for official multi-hop instructions and the exact RapidBuy/BuyPumpDelta wrapper layouts captured in reported transactions. Adds locally cached ALT resolution and rejects unresolved account identities by default, with explicit best-effort compatibility APIs where supported. Route intents expose instruction constraints and hop identities; confirmed execution amounts require execution metadata.

Includes the latest complete implementation, signed transaction fixtures, native offline-bank regression examples and reproducible local hot-path benchmarks. Unknown wrapper layouts are not inferred from embedded bytes. Local tests do not measure production network latency.

## v0.5.12 — Pump depth-2 compatibility

Adds signed Pump depth-2 create-event replay coverage, including ALT-loaded accounts and executed logs. Existing event layouts and public parsing interfaces remain unchanged.

Validation includes official Pump SDK 4.1.0 instruction comparisons, signed local LiteSVM simulations and SDK regressions. Local simulation state/config is synthetic; no funded transactions were broadcast and production latency was not measured.

## v0.5.11 — Signed transaction and hot-path hardening

Hardens signed transaction sanitization, loaded-address boundaries, ordered stream filtering and parser lifecycle. Aligns CLMM and DEX instruction/event layouts, nested CPI route attribution and liquidity/reward accounting. Adds independently signed wire fixtures, ALT failure cases and offline bank regressions.

Validation includes local CPU benchmarks and offline signed-bank scenarios. Measured hot paths use cached inputs without RPC. Benchmarks do not establish production network or transaction-landing latency. No funded mainnet transactions were broadcast.

在仓库根目录运行 `go run ./examples/hotpath_latency` 可重复执行 Go 原生热路径测试。
修改 `examples/hotpath_latency/config.json` 配置预热次数、样本数和批次数量，
也可传入其它配置路径。输出六笔已签名包装调用/ALT 交易一组的 p50/p95/p99
纳秒耗时及分配次数、字节数。读取样例不计入测量，不调用 RPC、不发送交易；
结果仅表示本地 CPU 工作量，不能代表网络上链延迟。

### v0.5.10

- 按实际指令 discriminator 和事件 mint 匹配 PumpFun create/create_v2 账户，支持 CPI，并拒绝猜测有歧义的匹配。
- 停止从 create_v2 的任意尾部账户推断 quote mint、vault、token program，保留权威解码字段。
- 恢复历史 CreateEvent 布局，并通过保存的真实主网成功、失败交易验证。
- 过滤失败交易中已经回滚的 Create/Buy 等 DEX 事件，并在 RPC/gRPC 转换中保留失败状态。

### v0.5.9

- Adds native gRPC lifecycle and block metadata support, instruction-level route analysis and StonkFun registry/snapshot examples.
- Improves exact metadata serialization, account matching and wire transaction bounds validation.
- Includes native alignment evidence and simulation route fixtures. These language releases do not yet implement the new CPMM creator-fee collection event API.

### v0.5.8

- 与 Rust sol-parser-sdk **0.7.3** 对齐，覆盖当前 PumpFun、PumpSwap 与 Pump Fees 协议定义。
- 新增 Meteora DAMM v2 事件：UpdateDelegatePermission、WithdrawDeadLiquidityReward、CreateConfig、CreateDynamicConfig。
- 对齐 EvtSwap2 180 字节布局、slot 门控 claiming/compounding fee 与 swap mode；EvtLiquidityChange 路由到 Add/Remove。
- 补充 DLMM swap 账户上下文字段、PumpFun pre/post 余额快照、LaunchLab trade quote 上下文。
- 在 PumpFun/PumpSwap instruction、log、account、gRPC 与 JSON 事件链路中补齐 creator fee 和 holder reward 字段。
- 拒绝 PumpSwap trade 与 CreatePool 已知字段的截断尾部，同时兼容完整历史布局。

### v0.5.6

- 新增 Meteora DBC 日志解析，并按当前 program context 做路由与过滤。
- 新增 Raydium CLMM/CPMM、Orca account parser。
- RPC 与 Geyser-to-RPC 转换保留区块内交易序号。
- ShredStream 对 account-only 或空 include-only filter 直接跳过指令解析。
- 收紧 RPC 日志 active-program 跟踪，保证低延迟解析准确性。

### v0.5.5

- ShredStream 线格式交易解析与 Rust/Node.js/Python 对齐。
- V0 ALT-loaded 指令账户不再整条跳过，而是用默认 pubkey 占位继续 best-effort 解析。
- 当 ShredStream 外层 program id 来自 ALT 时，按 discriminator 做候选 program id fallback。
- 改进 Pump.fun v2 短账户解析、同交易 post-merge 补全和多协议路由一致性。
- 新增 ShredStream ALT 账户与 ALT program id fallback 回归测试。
