package solparser

import (
	"bytes"
	"encoding/binary"
	"github.com/mr-tron/base58"
	"math/big"
)

const EventTypeAccountRawSnapshot EventType = "AccountRawSnapshot"
const EventTypeAccountLiquiditySnapshot EventType = "AccountLiquiditySnapshot"

type RawAccountSnapshotEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	Account      AccountData   `json:"account"`
	WriteVersion uint64        `json:"write_version,string"`
	IsStartup    bool          `json:"is_startup"`
}

func (e *RawAccountSnapshotEvent) GetMetadata() EventMetadata { return e.Metadata }
func (e *RawAccountSnapshotEvent) EventType() EventType       { return EventTypeAccountRawSnapshot }

type LiquidityAccountSnapshotEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Pubkey   string        `json:"pubkey"`
	Owner    string        `json:"owner"`
	Kind     any           `json:"kind"`
	Data     []byte        `json:"data"`
}

func (e *LiquidityAccountSnapshotEvent) GetMetadata() EventMetadata { return e.Metadata }
func (e *LiquidityAccountSnapshotEvent) EventType() EventType {
	return EventTypeAccountLiquiditySnapshot
}

// ParseLiquidityAccount validates owner, discriminator and variable tick payloads.
func ParseLiquidityAccount(account *AccountData, metadata EventMetadata) DexEvent {
	if account == nil || account.Lamports == 0 || account.Executable || len(account.Data) < 8 {
		return DexEvent{}
	}
	d := account.Data
	disc := func(b []byte) bool { return bytes.Equal(d[:8], b) }
	key := func(o int) string { return base58.Encode(d[o : o+32]) }
	var kind any
	switch account.Owner {
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		switch {
		case disc([]byte{247, 237, 227, 245, 215, 195, 222, 70}) && len(d) >= 429:
			kind = map[string]any{"LaunchLabPool": map[string]string{"base_mint": key(205), "quote_mint": key(237), "global_config": key(141), "platform_config": key(173)}}
		case disc([]byte{149, 8, 156, 202, 160, 252, 176, 217}) && len(d) >= 35:
			kind = "LaunchLabGlobalConfig"
		case disc([]byte{160, 78, 128, 0, 248, 83, 230, 160}) && len(d) >= 728:
			kind = "LaunchLabPlatformConfig"
		}
	case METEORA_DLMM_PROGRAM_ID:
		switch {
		case disc([]byte{33, 11, 49, 98, 181, 101, 177, 13}) && len(d) >= 904:
			kind = map[string]any{"DlmmPool": map[string]any{"token_x_mint": key(88), "token_y_mint": key(120), "active_id": int32(binary.LittleEndian.Uint32(d[76:])), "bin_step": binary.LittleEndian.Uint16(d[80:])}}
		case disc([]byte{92, 142, 92, 220, 5, 148, 70, 181}) && len(d) >= 10136:
			kind = map[string]any{"DlmmBinArray": struct {
				Pool  string `json:"pool"`
				Index int64  `json:"index,string"`
			}{key(24), int64(binary.LittleEndian.Uint64(d[8:]))}}
		case disc([]byte{80, 111, 124, 113, 55, 237, 18, 5}) && len(d) >= 1576:
			kind = map[string]any{"DlmmBitmap": map[string]string{"pool": key(8)}}
		}
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		switch {
		case disc([]byte{17, 216, 246, 142, 225, 199, 218, 56}) && len(d) >= 148:
			raw := append([]byte{}, d[44:60]...)
			for i, j := 0, 15; i < j; i, j = i+1, j-1 {
				raw[i], raw[j] = raw[j], raw[i]
			}
			bitmap := new(big.Int).SetBytes(raw)
			if bitmap.BitLen() > 88 {
				return DexEvent{}
			}
			o := 60
			for i := 0; i < 88; i++ {
				if o >= len(d) || d[o] > 1 || uint(d[o]) != bitmap.Bit(i) {
					return DexEvent{}
				}
				tag := d[o]
				o++
				if tag != 0 {
					o += 112
				}
				if o > len(d) {
					return DexEvent{}
				}
			}
			kind = map[string]any{"OrcaDynamicTickArray": map[string]any{"pool": key(12), "start_tick_index": int32(binary.LittleEndian.Uint32(d[8:])), "tick_bitmap": bitmap.String()}}
		case disc([]byte{139, 194, 131, 179, 140, 179, 229, 244}) && len(d) >= 254:
			kind = map[string]any{"OrcaAdaptiveOracle": map[string]string{"pool": key(8)}}
		}
	case RAYDIUM_CLMM_PROGRAM_ID:
		if disc([]byte{60, 150, 36, 219, 97, 128, 139, 153}) && len(d) >= 1832 {
			kind = map[string]any{"ClmmBitmap": map[string]string{"pool": key(8)}}
		}
	}
	if kind == nil {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountLiquiditySnapshot, Data: &LiquidityAccountSnapshotEvent{metadata, account.Pubkey, account.Owner, kind, append([]byte{}, d...)}}
}
