package shredstream

// ShredStreamConfig 对齐 Rust `shredstream::config::ShredStreamConfig`。
type ShredStreamConfig struct {
	ConnectionTimeoutMs    uint64
	RequestTimeoutMs       uint64
	MaxDecodingMessageSize int
	ReconnectDelayMs       uint64
	MaxReconnectAttempts   uint32
}

// DefaultShredStreamConfig 与 Rust `Default` 一致。
func DefaultShredStreamConfig() ShredStreamConfig {
	return ShredStreamConfig{
		ConnectionTimeoutMs:    8000,
		RequestTimeoutMs:       15000,
		MaxDecodingMessageSize: 1024 * 1024 * 100,
		ReconnectDelayMs:       1000,
		MaxReconnectAttempts:     3,
	}
}
