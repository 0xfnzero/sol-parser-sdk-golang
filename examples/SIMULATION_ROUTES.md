# Saved simulation execution evidence

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


### 第二批审查证据（2026-10-03）

`fixtures/batch2_cpmm_simulations_20261003.json` 含独立买入/卖出的真实主网模拟响应与原 wire，两个模拟均成功，三语言解析结果一致。买入实际输入 1000000、输出 13486034；卖出实际输入 1000000，但响应未明确 Token-2022 转账费，净到账保持未知。不要用快照报价填充实际成交字段。根目录第二批十轮审查证据同时保留触发滑点保护的延迟快照与历史三跳重放。


## CPMM LP 模拟中的铸造／销毁 CPI（2026-10-06）

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
