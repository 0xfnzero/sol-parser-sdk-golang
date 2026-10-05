# PumpFun 原生 SOL 成交量归属（2026-10-04）

三语言 parser 对 WSOL-sentinel PumpFun V2 统一报告钱包作为 native 支付/收款账户，不再误报指令中的 WSOL ATA 为实际结算账户。mint 仍按 Solana 的 WSOL sentinel 表示该资产；外层 WSOL 转换动作单独保留。

买入实际输入来自成功调用内的直属 System transfer：曲线本金、protocol fee、buyback fee、creator fee。必须观察到曲线收款；收款人不属于该指令明确账户、付款调用深度缺失或有嵌套歧义、总额溢出时返回未知。不会拿 specified amount 补实际金额。交易手续费、ATA rent、外层 WSOL 转出及临时账户关闭不会重复计入 swap。

真实旧/新买入模拟均得到实际 SOL 输入 10000。新 funded WSOL 模拟可同时看到外层 10000 WSOL 转出、临时账户关闭和 PumpFun 10000 SOL 支付；这不是两次用户支付。卖出直接修改 lamports 的实际净输出仍未归属，Token-2022 净到账缺少费用证据仍为未知。失败交易始终不报告实际成交量。

新 fixtures/pumpfun_settlement_mainnet_simulations_20261004.json 保留实际 mainnet 模拟 wire/response 和三语言相同解析预期。并补缺失 pool 支付、缺失深度、陌生收款人、嵌套及溢出负向回归。

本轮从 Yellowstone gRPC 捕获 5 组当前交易线索，冷核对 canonical ATA 均无可卖余额。还检查了事件中的非 ATA source：该买入源账户在核验时已不再是有效 initialized 账户；另一买入账户余额为零。未构建余额不足的卖出并当作成功。对应 trade examples/fixtures/pumpfun_sell_discovery_unverified_20261004.json 保存尝试记录。卖出真实模拟仍未验收。

前一份 settlement 记录中“当前 V2 实际 SOL 输入未知”的状态已被本轮买入 CPI 证据补齐；legacy 指令和卖出输出不能据此宣称已验收。

整体 Rust 公共 API/字段、gRPC fork/恢复、新数组及 SWQOS 验证仍按计划继续。没有广播、发布或创建 Actions。


后续账户级 gRPC 发现有余额的钱包，8 个钱包共 16 次独立 SOL/WSOL 卖出模拟成功。新 funded_sell corpus 三语言验证卖出 source debit 100000000，原生收款账户为钱包，net output 仍保留未知；WSOL minimum 7542 和实际 SOL 残余 397 有交易 SDK 独立账户余额证据。先前无可用余额的尝试仍作为历史记录保存。
