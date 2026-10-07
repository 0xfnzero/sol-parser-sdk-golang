# Sol Parser SDK golang usage

[English](../README.md) · [中文](../README_CN.md)

- [API compatibility](#api-compatibility)
- [gRPC account cache](#grpc-cache)
- [Simulation routes](#simulation-routes)

## API compatibility

以下说明当前公开行为和仍需调用方处理的边界；完整跨语言功能对齐仍在审核。

### 元数据精度

Node 的 `EventMetadata.slot`、`tx_index`、`block_time_us`、`grpc_recv_us` 现在是 `bigint`。输入支持 bigint、十进制字符串及安全整数 Number；超出安全整数范围的 Number、浮点数、越界 u64/i64 拒绝。根入口导出 `exactU64`、`exactI64`、`exactInteger`。使用 `dexEventToJsonString` 或 `bigintToJsonReplacer` 输出 JSON；直接 JSON.stringify(bigint) 会失败。

```ts
const slot = exactU64("18446744073709551615");
const latency = Number(BigInt(nowUs()) - event.metadata.grpc_recv_us);
console.log(JSON.stringify(event, bigintToJsonReplacer));
```

只在计算显示用延迟时将差值转为 Number。slot、index、时间及金额参与解析和排序时保留精确整数。

Go `EventMetadata` 保持 uint64/int64，JSON 输出上述四个字段为十进制字符串；Unmarshal 接受整数文本或十进制字符串，并检查范围，不接受 null/float/bool。Python `EventMetadata` 将十进制字符串归一为 int，拒绝 bool/float/越界；公开 JSON 工具输出上述元数据字段为字符串。各语言省略元数据整数时仍默认 0。未知实际净到账继续使用明确未知值，不能用 0 代替。

其余全部事件整数/JSON 字段的行为审核尚未完成。

### 事件别名

三语言提供 `LaunchLabPoolCreateEvent`、`StonkFunPoolCreateEvent`、`LaunchLabTradeEvent`、`StonkFunTradeEvent`。它们复用已有 RaydiumLaunchlab 结构。平台归属仍需池配置或迁移记录，不能仅凭共享 program ID 断言 StonkFun。

### gRPC 生命周期

一个 client 同时管理一个 DEX 订阅。再次订阅会停止并等待原订阅；动态更新和停止串行进行。Node `sub.join()` / Go `sub.Join()` 等待全部 worker；Python `await client.stop()` 等待任务。Node 使用 grpc-js + proto-loader，安装 SDK 不再包含 Yellowstone Rust N-API 客户端依赖。

Node `sub.states` / `sub.status()`、Go `sub.States` / `sub.Status()`、Python `queue.states` / `queue.status()` 提供连接、重连、连续性中断、丢弃及停止状态；保留原事件和错误消费方式。状态通知有界，状态快照可读取当前累计信息。重连成功不会自动清除连续性中断。调用方需中止交易准备，在选定 fork 下重新验证缓存，再恢复 readiness。

节点停止可能使事件消费结束或队列停止产出；Python 队列消费者同时观察状态以终止 get 等待。完整实时缓存恢复、新 array 发现和 commitment/fork 自动关联仍在实现。

示例见 [gRPC 缓存接入](#grpc-cache)；PublicNode 接入读取 GRPC_URL/GRPC_TOKEN，不打印凭据。


### DAMM route 解析

新增 legacy swap/swap2 route；mode 0/1 exact-in，mode 2 第一个金额为期望输出、第二个为最大输入，其它 mode unknown。失败保留意图、实际成交未知；Token-2022 缺净到账证据不填零。历史模拟 expected 新增 DAMM leg，调用方不可依赖旧的空 legs。

<a id="grpc-cache"></a>

## gRPC account cache

实时交易数据以 Rust sol-parser-sdk 的 Yellowstone gRPC 接口为准。三个原生 parser 使用相同来源：交易事件提供 pool/mint/方向线索，原始账户事件提供当前字节、owner、slot、write_version，Clock 提供 epoch/时间，BlockMeta 或显式 gRPC unary 提供 blockhash。

此流程不使用 WebSocket。parser 没有 WebSocket 订阅接口；parser 的 SubscriptionHandle/Subscription 管理 gRPC 流。trade-sdk 中另有通用 WebSocket 兼容模块，其实现与测试不是 parser gRPC 能力的证明。

| 语言 | parser 实时入口 | 原始账户写入 trade cache |
| --- | --- | --- |
| Node.js | YellowstoneGrpc.subscribeDexEvents | SubscriptionAccountCache.updateFromParser(raw) |
| Python | YellowstoneGrpc.subscribe_dex_events | SubscriptionAccountCache.update_from_parser(raw) |
| Go | YellowstoneGrpc.SubscribeDexEvents | SubscriptionAccountCache.UpdateRaw(pubkey, owner, data, lamports, slot, writeVersion) |

Node/Python 传原始账户事件的 payload，而非最外层 DexEvent wrapper。Go 将事件公钥解析为 solana.PublicKey；解析失败必须返回错误。账户关闭（lamports=0）保留为 tombstone，不复用旧字节。不要从 swap 事件伪造账户完整状态、当前手续费或 slot。

推荐接入顺序：

1. 在 parser 订阅交易与 AccountRawSnapshot；raw snapshot 需要显式开启事件过滤。按已知 pool/config/vault/mint/tick/bin/Clock 地址订阅。路由候选是线索，用户可选择 SOL、WSOL、USDC 或股票作为支付/收款资产。
2. 原始账户更新写入选定 fork 的本地 cache，保留真实版本。缺失配置/新 array 必须在交易前通过订阅或冷启动补齐；不会用 WebSocket 或热路径 RPC 自动兜底。
3. 冻结 cache.snapshot()，替换为用户的钱包、金额与资产，使用 cached prepare/route。本地校验 owner、pool/mint 连通性、状态、费率、Token-2022 费用及 freshness 后报价和构建。本次买入或卖出是独立交易。
4. 模拟可显式调用 simulateTransaction；这是验证步骤。读取模拟结果用 parser 的 simulation_routes 示例，不等于真实发送。

gRPC 流不保证首次提供所有静态账户。pool 已更新不等于 mint/config 已新鲜；不得把其它账户 slot 改成 Clock slot。重连后须确认缺失更新和 fork，缓存就绪前暂停构建。缺失数据或过期状态明确失败。

### 完整示例

两个 SDK 安装在各自环境中即可运行，也可分别在对应仓库运行。快照文件是两个示例间的明确接口，安装和运行不需要 Rust。

| 语言 | parser 仓库：gRPC 刷新 | trade 仓库：本地构建 |
| --- | --- | --- |
| Node.js | npx tsx examples/stonkfun_snapshot_refresh.ts /absolute/snapshot.json --require-pool-update | npx tsx examples/cached_trade.ts /absolute/snapshot.json |
| Python | python examples/stonkfun_snapshot_refresh.py /absolute/snapshot.json --require-pool-update | python examples/cached_trade.py /absolute/snapshot.json |
| Go | go run ./examples/stonkfun_snapshot_refresh /absolute/snapshot.json --require-pool-update | go run ./examples/cached_trade /absolute/snapshot.json |

使用与 cached_trade 匹配的 accounts[]、legs[] 完整快照，例如现有 cached_tip_route_* 样本；更换自己的钱包、金额和本次方向。CPMM 的 named snapshot 应交给 cached_cpmm 示例。不要直接用历史快照实时交易。

parser 刷新读取 GRPC_URL / GRPC_TOKEN，保留没更新账户的原 slot，可能因 provider 没有发送 pool 更新而超时；trade 默认不联网。需要模拟时为 trade 示例增加 --simulate，并配置 RPC_URL。冷启动与显式模拟允许 RPC，交易报价和构建热路径禁止 RPC。

本文件描述已实现的接入接口；完整自动重连恢复、fork 一致性策略、新 array 发现和全部协议缓存仍须逐项验证。之前十轮报告中的 WebSocket 修复仅属于 trade-sdk 通用模块，不计入这些 gRPC 对齐能力。

<a id="simulation-routes"></a>

## Simulation routes

运行trade侧cached_trade示例时，--simulate返回包含innerInstructions的模拟结果。内部SPL转账可能是jsonParsed，原来的compiled route入口仍会拒绝它。专用入口需要**原始交易wire和对应完整simulateTransaction响应**，仅离线处理证据，不查询RPC或重新构建签名交易。

```sh
# Go
go run ./examples/simulation_routes solparser/testdata/cached_tip_routes_20261002.json route-buy
# Go
go run ./examples/simulation_routes solparser/testdata/cached_tip_routes_20261002.json route-buy
# Node
npx tsx examples/simulation_routes.ts examples/fixtures/cached_tip_routes_20261002.json route-buy
# Python
PYTHONPATH=. python examples/simulation_routes.py examples/fixtures/cached_tip_routes_20261002.json route-buy
```

```ts
import { analyzeSimulationRoutes } from "sol-parser-sdk";
const route = analyzeSimulationRoutes(wire, simulationResponse);
```

```python
from sol_parser import analyze_simulation_routes
route = analyze_simulation_routes(wire, simulation_response)
```

SPL普通／checked／明确fee转账，以及首次ATA的getAccountDataSize／createAccount／initializeImmutableOwner／initializeAccount3的parsed表示被还原为分析证据；raw/compiled CPI原样保留位置和stack height。其它parsed指令明确报错。金额使用最小单位amount字段，不用uiAmount浮点值。缺失Token-2022手续费的净到账保持未知；失败交易实际成交字段为空。tip不会被当成WSOL充值。含ALT的V0需要原有compiled route入口与显式loadedAddresses；此入口不查询lookup table。调用方必须提供与wire对应的可信模拟响应。

保存快照与模拟仅用于复现，不是实时状态。Go parser 已补齐 AnalyzeSimulationRoutes，接受 wire、完整模拟响应的 []byte 和可选 graduatedPools；安装和使用不依赖 Rust。

```go
route, err := solparser.AnalyzeSimulationRoutes(wire, simulationResponseJSON, nil)
if err != nil { return err }
// route.Legs 的 ActualOutputAmount == nil 表示净到账未知，不能当作零。
```

```go
route, err := solparser.AnalyzeSimulationRoutes(wire, simulationResponseJSON, nil)
if err != nil { return err }
// route.Legs 的 ActualOutputAmount == nil 表示净到账未知，不能当作零。
```

当前银行模拟复现：`examples/fixtures/simulation_routes_live_20261002.json`。包括因旧保护价格触发滑点错误的买入和成功卖出；失败时只保留意图，actual amount为空。Go可以用相同文件路径运行上述示例。保存wire用replaceRecentBlockhash重新模拟并不等于当前状态重新报价。

首次 ATA 模拟：`examples/fixtures/simulation_ata_20261002.json`，含 classic SPL、Token-2022 股票和带建账户的单池／三跳卖出。setup指令用于mint推断，不计入成交转账；getAccountDataSize仅支持空列表或immutableOwner扩展，其它扩展明确报错。compiled route入口现在也要求meta.err显式存在，缺失状态不再默认为成功。


2026-10-03: Prefunded ATA creation is also supported: System transfer, allocate and assign CPI are reconstructed and validated. getAccountDataSize accepts empty/immutableOwner extension lists only; unknown parsed instructions remain errors. Wire and compiled-route indexes, duplicate CPI groups and token-balance indexes are validated before event reconstruction. No RPC is used by the parser. Mainnet prefunded-ATA evidence is covered by the review10 fixture tests.


#### 第二批审查证据（2026-10-03）

`fixtures/batch2_cpmm_simulations_20261003.json` 含独立买入/卖出的真实主网模拟响应与原 wire，两个模拟均成功，三语言解析结果一致。买入实际输入 1000000、输出 13486034；卖出实际输入 1000000，但响应未明确 Token-2022 转账费，净到账保持未知。不要用快照报价填充实际成交字段。根目录第二批十轮审查证据同时保留触发滑点保护的延迟快照与历史三跳重放。


### CPMM LP 模拟中的铸造／销毁 CPI（2026-10-06）

三语言支持 parsed SPL／Token-2022 的 mintTo、burn、mintToChecked、burnChecked，校验精确 u64 金额、u8 decimals、多签账户与原始指令位置。它们保留为供分析的指令，不计入转账或 swap 成交；LP 顶层操作仍可能出现在 unknown_invocations，不意味着已新增 LP 事件 API。首次 ATA、资金准备 swap、LP 操作可在同笔模拟中解析。未知 Token-2022 净到账仍为 null。

真实主网模拟样例：examples/fixtures/cpmm_lp_simulations_20261006.json，包含 LP 存入与取出。在各语言仓库运行：

```sh
# Go
go run ./examples/simulation_routes examples/fixtures/cpmm_lp_simulations_20261006.json
# Node
npx tsx examples/simulation_routes.ts examples/fixtures/cpmm_lp_simulations_20261006.json
# Python
PYTHONPATH=. python examples/simulation_routes.py examples/fixtures/cpmm_lp_simulations_20261006.json
```

该样例包含 sigVerify=false 的资金准备前缀，仅作执行验证，不可广播。新交易的完整原始响应可由 Rust CPMM_SIMULATION_DIR 保存后，用相同示例解析。
