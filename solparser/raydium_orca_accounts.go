package solparser

import (
	"encoding/binary"
	"math/big"
)

var (
	clmmAmmConfigDisc      = []byte{218, 244, 33, 104, 203, 203, 43, 111}
	clmmPoolStateDisc      = []byte{247, 237, 227, 245, 215, 195, 222, 70}
	clmmTickArrayStateDisc = []byte{192, 155, 85, 205, 49, 249, 129, 42}
	orcaWhirlpoolDisc      = []byte{63, 149, 209, 12, 225, 128, 99, 9}
	orcaPositionDisc       = []byte{170, 188, 143, 228, 122, 64, 247, 208}
	orcaTickArrayDisc      = []byte{69, 97, 189, 190, 110, 7, 66, 187}
	orcaFeeTierDisc        = []byte{56, 75, 159, 76, 142, 68, 190, 105}
	orcaConfigDisc         = []byte{157, 20, 49, 224, 217, 87, 193, 254}
)

const (
	clmmAmmConfigBody      = 109
	clmmPoolStateBody      = 1536
	clmmTickArrayStateBody = 10232
	clmmTickArrayLen       = 60
	cpmmAmmConfigBody      = 228
	cpmmPoolStateBody      = 629
	orcaWhirlpoolBody      = 645
	orcaPositionBody       = 208
	orcaTickArrayBody      = 9980
	orcaFeeTierBody        = 36
	orcaConfigBody         = 98
	orcaTickArrayLen       = 88
)

type accountReader struct {
	data []byte
	off  int
	ok   bool
}

func newAccountReader(data []byte) *accountReader {
	return &accountReader{data: data, ok: true}
}

func (r *accountReader) take(n int) []byte {
	if !r.ok || r.off+n > len(r.data) {
		r.ok = false
		return nil
	}
	out := r.data[r.off : r.off+n]
	r.off += n
	return out
}

func (r *accountReader) u8() uint8 {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}

func (r *accountReader) bool() bool { return r.u8() != 0 }

func (r *accountReader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (r *accountReader) u32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func (r *accountReader) i32() int32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(b))
}

func (r *accountReader) u64() uint64 {
	b := r.take(8)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(b)
}

func (r *accountReader) u128() string {
	b := r.take(16)
	if b == nil {
		return "0"
	}
	var raw [16]byte
	copy(raw[:], b)
	return u128LEDecimalString(raw)
}

func (r *accountReader) i128() string {
	b := r.take(16)
	if b == nil {
		return "0"
	}
	return i128LEDecimalString(b)
}

func (r *accountReader) pubkey() string {
	b := r.take(32)
	if b == nil {
		return zeroPubkey
	}
	return Base58Encode(b)
}

func (r *accountReader) bytes(n int) []uint8 {
	b := r.take(n)
	if b == nil {
		return nil
	}
	out := make([]uint8, n)
	copy(out, b)
	return out
}

func i128LEDecimalString(le []byte) string {
	be := make([]byte, len(le))
	for i := range le {
		be[len(le)-1-i] = le[i]
	}
	v := new(big.Int).SetBytes(be)
	if len(le) > 0 && le[len(le)-1]&0x80 != 0 {
		mod := new(big.Int).Lsh(big.NewInt(1), uint(len(le)*8))
		v.Sub(v, mod)
	}
	return v.String()
}

func accountBody(data []byte, disc []byte, bodySize int) []byte {
	if len(data) < 8+bodySize || !HasDiscriminator(data, disc) {
		return nil
	}
	return data[8:]
}

func ParseRaydiumClmmAccount(account *AccountData, metadata EventMetadata) DexEvent {
	if account.Owner != RAYDIUM_CLMM_PROGRAM_ID {
		return DexEvent{}
	}
	if ev := ParseRaydiumClmmAmmConfig(account, metadata); ev.Type != "" {
		return ev
	}
	if ev := ParseRaydiumClmmPoolState(account, metadata); ev.Type != "" {
		return ev
	}
	return ParseRaydiumClmmTickArrayState(account, metadata)
}

func ParseRaydiumCpmmAccount(account *AccountData, metadata EventMetadata) DexEvent {
	if account.Owner != RAYDIUM_CPMM_PROGRAM_ID {
		return DexEvent{}
	}
	if ev := ParseRaydiumCpmmAmmConfig(account, metadata); ev.Type != "" {
		return ev
	}
	return ParseRaydiumCpmmPoolState(account, metadata)
}

func ParseOrcaWhirlpoolAccount(account *AccountData, metadata EventMetadata) DexEvent {
	if account.Owner != ORCA_WHIRLPOOL_PROGRAM_ID {
		return DexEvent{}
	}
	if ev := ParseOrcaWhirlpool(account, metadata); ev.Type != "" {
		return ev
	}
	if ev := ParseOrcaPosition(account, metadata); ev.Type != "" {
		return ev
	}
	if ev := ParseOrcaTickArray(account, metadata); ev.Type != "" {
		return ev
	}
	if ev := ParseOrcaFeeTier(account, metadata); ev.Type != "" {
		return ev
	}
	return ParseOrcaWhirlpoolsConfig(account, metadata)
}

func ParseRaydiumClmmAmmConfig(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, clmmAmmConfigDisc, clmmAmmConfigBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	cfg := RaydiumClmmAmmConfig{
		Bump:            r.u8(),
		Index:           r.u16(),
		Owner:           r.pubkey(),
		ProtocolFeeRate: r.u32(),
		TradeFeeRate:    r.u32(),
		TickSpacing:     r.u16(),
		FundFeeRate:     r.u32(),
		PaddingU32:      r.u32(),
		FundOwner:       r.pubkey(),
		Padding:         readU64Array(r, 3),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountRaydiumClmmAmmConfig, Data: &RaydiumClmmAmmConfigAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, AmmConfig: cfg,
	}}
}

func ParseRaydiumClmmPoolState(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, clmmPoolStateDisc, clmmPoolStateBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	state := RaydiumClmmPoolState{
		Bump:                []uint8{r.u8()},
		AmmConfig:           r.pubkey(),
		Owner:               r.pubkey(),
		TokenMint0:          r.pubkey(),
		TokenMint1:          r.pubkey(),
		TokenVault0:         r.pubkey(),
		TokenVault1:         r.pubkey(),
		ObservationKey:      r.pubkey(),
		MintDecimals0:       r.u8(),
		MintDecimals1:       r.u8(),
		TickSpacing:         r.u16(),
		Liquidity:           r.u128(),
		SqrtPriceX64:        r.u128(),
		TickCurrent:         r.i32(),
		Padding3:            r.u16(),
		Padding4:            r.u16(),
		FeeGrowthGlobal0X64: r.u128(),
		FeeGrowthGlobal1X64: r.u128(),
		ProtocolFeesToken0:  r.u64(),
		ProtocolFeesToken1:  r.u64(),
		Padding5:            readU128Array(r, 4),
		Status:              r.u8(),
		FeeOn:               r.u8(),
		Padding:             r.bytes(6),
		RewardInfos:         readClmmRewardInfos(r),
		TickArrayBitmap:     readU64Array(r, 16),
		Padding6:            readU64Array(r, 4),
		FundFeesToken0:      r.u64(),
		FundFeesToken1:      r.u64(),
		OpenTime:            r.u64(),
		RecentEpoch:         r.u64(),
		DynamicFeeInfo:      readClmmDynamicFeeInfo(r),
		Padding1:            readU64Array(r, 14),
		Padding2:            readU64Array(r, 32),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountRaydiumClmmPoolState, Data: &RaydiumClmmPoolStateAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, PoolState: state,
	}}
}

func ParseRaydiumClmmTickArrayState(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, clmmTickArrayStateDisc, clmmTickArrayStateBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	state := RaydiumClmmTickArrayState{
		PoolID:               r.pubkey(),
		StartTickIndex:       r.i32(),
		Ticks:                readClmmTicks(r),
		InitializedTickCount: r.u8(),
		RecentEpoch:          r.u64(),
		Padding:              r.bytes(107),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountRaydiumClmmTickArrayState, Data: &RaydiumClmmTickArrayStateAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, TickArrayState: state,
	}}
}

func readClmmRewardInfos(r *accountReader) []RaydiumClmmRewardInfo {
	out := make([]RaydiumClmmRewardInfo, 3)
	for i := range out {
		out[i] = RaydiumClmmRewardInfo{
			RewardState:           r.u8(),
			OpenTime:              r.u64(),
			EndTime:               r.u64(),
			LastUpdateTime:        r.u64(),
			EmissionsPerSecondX64: r.u128(),
			RewardTotalEmitted:    r.u64(),
			RewardClaimed:         r.u64(),
			TokenMint:             r.pubkey(),
			TokenVault:            r.pubkey(),
			Authority:             r.pubkey(),
			RewardGrowthGlobalX64: r.u128(),
		}
	}
	return out
}

func readClmmDynamicFeeInfo(r *accountReader) RaydiumClmmDynamicFeeInfo {
	return RaydiumClmmDynamicFeeInfo{
		FilterPeriod:              r.u16(),
		DecayPeriod:               r.u16(),
		ReductionFactor:           r.u16(),
		DynamicFeeControl:         r.u32(),
		MaxVolatilityAccumulator:  r.u32(),
		TickSpacingIndexReference: r.i32(),
		VolatilityReference:       r.u32(),
		VolatilityAccumulator:     r.u32(),
		LastUpdateTimestamp:       r.u64(),
		Padding:                   r.bytes(46),
	}
}

func readClmmTicks(r *accountReader) []RaydiumClmmTick {
	out := make([]RaydiumClmmTick, clmmTickArrayLen)
	for i := range out {
		out[i] = RaydiumClmmTick{
			Tick:                      r.i32(),
			LiquidityNet:              r.i128(),
			LiquidityGross:            r.u128(),
			FeeGrowthOutside0X64:      r.u128(),
			FeeGrowthOutside1X64:      r.u128(),
			RewardGrowthsOutsideX64:   readU128Array(r, 3),
			OrderPhase:                r.u64(),
			OrdersAmount:              r.u64(),
			PartFilledOrdersRemaining: r.u64(),
			UnfilledRatioX64:          r.u128(),
			Padding:                   readU32Array(r, 3),
		}
	}
	return out
}

func ParseRaydiumCpmmAmmConfig(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, clmmAmmConfigDisc, cpmmAmmConfigBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	cfg := RaydiumCpmmAmmConfig{
		Bump:              r.u8(),
		DisableCreatePool: r.bool(),
		Index:             r.u16(),
		TradeFeeRate:      r.u64(),
		ProtocolFeeRate:   r.u64(),
		FundFeeRate:       r.u64(),
		CreatePoolFee:     r.u64(),
		ProtocolOwner:     r.pubkey(),
		FundOwner:         r.pubkey(),
		CreatorFeeRate:    r.u64(),
		Padding:           readU64Array(r, 15),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountRaydiumCpmmAmmConfig, Data: &RaydiumCpmmAmmConfigAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, AmmConfig: cfg,
	}}
}

func ParseRaydiumCpmmPoolState(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, clmmPoolStateDisc, cpmmPoolStateBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	state := RaydiumCpmmPoolState{
		AmmConfig:          r.pubkey(),
		PoolCreator:        r.pubkey(),
		Token0Vault:        r.pubkey(),
		Token1Vault:        r.pubkey(),
		LpMint:             r.pubkey(),
		Token0Mint:         r.pubkey(),
		Token1Mint:         r.pubkey(),
		Token0Program:      r.pubkey(),
		Token1Program:      r.pubkey(),
		ObservationKey:     r.pubkey(),
		AuthBump:           r.u8(),
		Status:             r.u8(),
		LpMintDecimals:     r.u8(),
		Mint0Decimals:      r.u8(),
		Mint1Decimals:      r.u8(),
		LpSupply:           r.u64(),
		ProtocolFeesToken0: r.u64(),
		ProtocolFeesToken1: r.u64(),
		FundFeesToken0:     r.u64(),
		FundFeesToken1:     r.u64(),
		OpenTime:           r.u64(),
		RecentEpoch:        r.u64(),
		CreatorFeeOn:       r.u8(),
		EnableCreatorFee:   r.bool(),
		Padding1:           r.bytes(6),
		CreatorFeesToken0:  r.u64(),
		CreatorFeesToken1:  r.u64(),
		Padding:            readU64Array(r, 28),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountRaydiumCpmmPoolState, Data: &RaydiumCpmmPoolStateAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, PoolState: state,
	}}
}

func ParseOrcaWhirlpool(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, orcaWhirlpoolDisc, orcaWhirlpoolBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	whirlpool := OrcaWhirlpoolAccount{
		WhirlpoolsConfig:           r.pubkey(),
		WhirlpoolBump:              r.u8(),
		TickSpacing:                r.u16(),
		TickSpacingSeed:            r.bytes(2),
		FeeRate:                    r.u16(),
		ProtocolFeeRate:            r.u16(),
		Liquidity:                  r.u128(),
		SqrtPrice:                  r.u128(),
		TickCurrentIndex:           r.i32(),
		ProtocolFeeOwedA:           r.u64(),
		ProtocolFeeOwedB:           r.u64(),
		TokenMintA:                 r.pubkey(),
		TokenVaultA:                r.pubkey(),
		FeeGrowthGlobalA:           r.u128(),
		TokenMintB:                 r.pubkey(),
		TokenVaultB:                r.pubkey(),
		FeeGrowthGlobalB:           r.u128(),
		RewardLastUpdatedTimestamp: r.u64(),
		RewardInfos:                readOrcaWhirlpoolRewardInfos(r),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountOrcaWhirlpool, Data: &OrcaWhirlpoolAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, Whirlpool: whirlpool,
	}}
}

func ParseOrcaPosition(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, orcaPositionDisc, orcaPositionBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	position := OrcaPositionAccount{
		Whirlpool:            r.pubkey(),
		PositionMint:         r.pubkey(),
		Liquidity:            r.u128(),
		TickLowerIndex:       r.i32(),
		TickUpperIndex:       r.i32(),
		FeeGrowthCheckpointA: r.u128(),
		FeeOwedA:             r.u64(),
		FeeGrowthCheckpointB: r.u128(),
		FeeOwedB:             r.u64(),
		RewardInfos:          readOrcaPositionRewardInfos(r),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountOrcaPosition, Data: &OrcaPositionAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, Position: position,
	}}
}

func ParseOrcaTickArray(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, orcaTickArrayDisc, orcaTickArrayBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	tickArray := OrcaTickArrayAccount{
		StartTickIndex: r.i32(),
		Ticks:          readOrcaTicks(r),
		Whirlpool:      r.pubkey(),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountOrcaTickArray, Data: &OrcaTickArrayAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, TickArray: tickArray,
	}}
}

func ParseOrcaFeeTier(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, orcaFeeTierDisc, orcaFeeTierBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	feeTier := OrcaFeeTierAccount{
		WhirlpoolsConfig: r.pubkey(),
		TickSpacing:      r.u16(),
		DefaultFeeRate:   r.u16(),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountOrcaFeeTier, Data: &OrcaFeeTierAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, FeeTier: feeTier,
	}}
}

func ParseOrcaWhirlpoolsConfig(account *AccountData, metadata EventMetadata) DexEvent {
	body := accountBody(account.Data, orcaConfigDisc, orcaConfigBody)
	if body == nil {
		return DexEvent{}
	}
	r := newAccountReader(body)
	config := OrcaWhirlpoolsConfigAccount{
		FeeAuthority:                  r.pubkey(),
		CollectProtocolFeesAuthority:  r.pubkey(),
		RewardEmissionsSuperAuthority: r.pubkey(),
		DefaultProtocolFeeRate:        r.u16(),
	}
	if !r.ok {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypeAccountOrcaWhirlpoolsConfig, Data: &OrcaWhirlpoolsConfigAccountEvent{
		Metadata: metadata, Pubkey: account.Pubkey, Config: config,
	}}
}

func readOrcaWhirlpoolRewardInfos(r *accountReader) []OrcaWhirlpoolRewardInfo {
	out := make([]OrcaWhirlpoolRewardInfo, 3)
	for i := range out {
		out[i] = OrcaWhirlpoolRewardInfo{
			Mint:                  r.pubkey(),
			Vault:                 r.pubkey(),
			Authority:             r.pubkey(),
			EmissionsPerSecondX64: r.u128(),
			GrowthGlobalX64:       r.u128(),
		}
	}
	return out
}

func readOrcaPositionRewardInfos(r *accountReader) []OrcaPositionRewardInfo {
	out := make([]OrcaPositionRewardInfo, 3)
	for i := range out {
		out[i] = OrcaPositionRewardInfo{
			GrowthInsideCheckpoint: r.u128(),
			AmountOwed:             r.u64(),
		}
	}
	return out
}

func readOrcaTicks(r *accountReader) []OrcaTick {
	out := make([]OrcaTick, orcaTickArrayLen)
	for i := range out {
		out[i] = OrcaTick{
			Initialized:          r.bool(),
			LiquidityNet:         r.i128(),
			LiquidityGross:       r.u128(),
			FeeGrowthOutsideA:    r.u128(),
			FeeGrowthOutsideB:    r.u128(),
			RewardGrowthsOutside: readU128Array(r, 3),
		}
	}
	return out
}

func readU64Array(r *accountReader, count int) []uint64 {
	out := make([]uint64, count)
	for i := range out {
		out[i] = r.u64()
	}
	return out
}

func readU32Array(r *accountReader, count int) []uint32 {
	out := make([]uint32, count)
	for i := range out {
		out[i] = r.u32()
	}
	return out
}

func readU128Array(r *accountReader, count int) []string {
	out := make([]string, count)
	for i := range out {
		out[i] = r.u128()
	}
	return out
}
