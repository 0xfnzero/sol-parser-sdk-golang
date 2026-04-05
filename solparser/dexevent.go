package solparser

// DexEvent DEX 事件结构体（外部标签格式，与 TS / Rust 对齐）
// 使用 map[string]any 承载内部字段，与 Go matcher 系列函数保持一致。
type DexEvent map[string]any
