# 本轮 API 迁移（实施中，尚未发布）

基线：Rust sol-parser-sdk 0.7.7。完整功能对照仍在审核，本文件描述已修改的公开行为。

## 元数据精度

Node 的 `EventMetadata.slot`、`tx_index`、`block_time_us`、`grpc_recv_us` 现在是 `bigint`。输入支持 bigint、十进制字符串及安全整数 Number；超出安全整数范围的 Number、浮点数、越界 u64/i64 拒绝。根入口导出 `exactU64`、`exactI64`、`exactInteger`。使用 `dexEventToJsonString` 或 `bigintToJsonReplacer` 输出 JSON；直接 JSON.stringify(bigint) 会失败。

```ts
const slot = exactU64("18446744073709551615");
const latency = Number(BigInt(nowUs()) - event.metadata.grpc_recv_us);
console.log(JSON.stringify(event, bigintToJsonReplacer));
```

只在计算显示用延迟时将差值转为 Number。slot、index、时间及金额参与解析和排序时保留精确整数。

Go `EventMetadata` 保持 uint64/int64，JSON 输出上述四个字段为十进制字符串；Unmarshal 接受整数文本或十进制字符串，并检查范围，不接受 null/float/bool。Python `EventMetadata` 将十进制字符串归一为 int，拒绝 bool/float/越界；公开 JSON 工具输出上述元数据字段为字符串。各语言省略元数据整数时仍默认 0。未知实际净到账继续使用明确未知值，不能用 0 代替。

其余全部事件整数/JSON 字段的行为审核尚未完成。

## 事件别名

三语言提供 `LaunchLabPoolCreateEvent`、`StonkFunPoolCreateEvent`、`LaunchLabTradeEvent`、`StonkFunTradeEvent`。它们复用已有 RaydiumLaunchlab 结构。平台归属仍需池配置或迁移记录，不能仅凭共享 program ID 断言 StonkFun。

## gRPC 生命周期

一个 client 同时管理一个 DEX 订阅。再次订阅会停止并等待原订阅；动态更新和停止串行进行。Node `sub.join()` / Go `sub.Join()` 等待全部 worker；Python `await client.stop()` 等待任务。Node 使用 grpc-js + proto-loader，安装 SDK 不再包含 Yellowstone Rust N-API 客户端依赖。

Node `sub.states` / `sub.status()`、Go `sub.States` / `sub.Status()`、Python `queue.states` / `queue.status()` 提供连接、重连、连续性中断、丢弃及停止状态；保留原事件和错误消费方式。状态通知有界，状态快照可读取当前累计信息。重连成功不会自动清除连续性中断。调用方需中止交易准备，在选定 fork 下重新验证缓存，再恢复 readiness。

节点停止可能使事件消费结束或队列停止产出；Python 队列消费者同时观察状态以终止 get 等待。完整实时缓存恢复、新 array 发现和 commitment/fork 自动关联仍在实现。

示例见 `examples/GRPC_CACHE.md`；PublicNode 接入读取 GRPC_URL/GRPC_TOKEN，不打印凭据。


## DAMM route 解析（2026-10-04）

新增 legacy swap/swap2 route；mode 0/1 exact-in，mode 2 第一个金额为期望输出、第二个为最大输入，其它 mode unknown。失败保留意图、实际成交未知；Token-2022 缺净到账证据不填零。历史模拟 expected 新增 DAMM leg，调用方不可依赖旧的空 legs。
