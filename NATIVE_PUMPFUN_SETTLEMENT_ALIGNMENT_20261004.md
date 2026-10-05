# PumpFun 交易形态解析对齐（2026-10-04）

三语言 transaction route 增加 PumpFun 旧版 buy/sell/exact-SOL-in 与 V2 exact-quote-in、buy exact-out、sell exact-in 的指令意图识别。V2 从账户提取 curve、钱包、base/quote mint 和账户；失败交易保留意图。

真实 mainnet V2 模拟 corpus：测试 fixtures/pumpfun_current_mainnet_simulations_20261004.json（三语言 fixtures 位置按既有规范）。三语言复放验证单个 PumpFun leg，实际 SOL 输入及 Token-2022 净输出缺证据仍为未知，不以 specified amount 或零代替。该 corpus 的旧 funded WSOL 案例实际走钱包 SOL；不能宣称 WSOL 支付成功。

trade-sdk 现已另存专用 WSOL 结算模拟：原 ATA 实际扣减 10000，V1 wire 1238 字节。parser 对 System CPI 实际 SOL debit 的逐调用归属及所有旧版账户净余额行为仍需补全。仅识别 route intent 不代表所有归一化事件字段已完全对齐 Rust 0.7.7。

实时仍使用 Yellowstone gRPC；本次 RPC 仅为显式冷准备和模拟，无发送。整体公开 API/字段矩阵、生命周期恢复、fork、动态数组等其余验收继续保留未完成状态。

验证摘要见 [机器可读记录](NATIVE_PUMPFUN_SETTLEMENT_EVIDENCE_20261004.json)：六仓全量测试通过；最终新增回归单独通过；Go 两仓全量 race/vet、Node 类型/构建/示例类型、独立 npm/wheel 安装后的公共入口 wire 对照通过。跳过的在线测试不计为验收。
