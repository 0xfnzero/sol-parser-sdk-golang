package solparser

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/mr-tron/base58"
	"math"
)

const StonkFunStandardPlatformConfig = "4E876qZTE9FJMrBzgVtBrSrzz2TLivB5Y5QXPjB4gZL7"
const StonkFunRewardPlatformConfig = "6BwHHDg3u1854jC8PDLXvR4spTcLNaoBxLJNGC4nTESt"
const routeToken = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
const routeToken2022 = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
const routeWSOL = "So11111111111111111111111111111111111111112"

func StonkFunModeFromPlatformConfig(key string) *string {
	m := ""
	if key == StonkFunStandardPlatformConfig {
		m = "Standard"
	} else if key == StonkFunRewardPlatformConfig {
		m = "Reward"
	} else {
		return nil
	}
	return &m
}

type InstructionPosition struct {
	OuterIndex  int  `json:"outer_index"`
	InnerIndex  *int `json:"inner_index"`
	StackHeight *int `json:"stack_height"`
}
type RouteSwapLeg struct {
	Position               InstructionPosition `json:"position"`
	Program                string              `json:"program"`
	Protocol               string              `json:"protocol"`
	Pool                   string              `json:"pool"`
	Trader                 string              `json:"trader"`
	InputAccount           string              `json:"input_account"`
	OutputAccount          string              `json:"output_account"`
	InputMint              *string             `json:"input_mint"`
	OutputMint             *string             `json:"output_mint"`
	AmountSpecifiedIsInput bool                `json:"amount_specified_is_input"`
	SpecifiedAmount        uint64              `json:"specified_amount,string"`
	OtherAmountThreshold   uint64              `json:"other_amount_threshold,string"`
	ActualInputAmount      *uint64             `json:"actual_input_amount,string"`
	ActualOutputAmount     *uint64             `json:"actual_output_amount,string"`
	StonkFunMode           *string             `json:"stonkfun_mode"`
	StonkFunGraduated      bool                `json:"stonkfun_graduated"`
}
type RouteTokenTransfer struct {
	Position    InstructionPosition `json:"position"`
	Program     string              `json:"program"`
	Source      string              `json:"source"`
	Destination string              `json:"destination"`
	Mint        *string             `json:"mint"`
	Amount      uint64              `json:"amount,string"`
	WithheldFee *uint64             `json:"withheld_fee,string"`
}
type RouteUnknownInvocation struct {
	Position                InstructionPosition `json:"position"`
	Program                 string              `json:"program"`
	HasTokenTransfers       bool                `json:"has_token_transfers"`
	HasKnownSwapDescendants bool                `json:"has_known_swap_descendants"`
}
type RouteNativeTokenAction struct {
	Position InstructionPosition `json:"position"`
	Account  string              `json:"account"`
	Action   any                 `json:"action"`
}
type TransactionRoute struct {
	Signature          string                   `json:"signature"`
	Succeeded          bool                     `json:"succeeded"`
	Legs               []RouteSwapLeg           `json:"legs"`
	Transfers          []RouteTokenTransfer     `json:"transfers"`
	NativeTokenActions []RouteNativeTokenAction `json:"native_token_actions"`
	UnknownInvocations []RouteUnknownInvocation `json:"unknown_invocations"`
}
type routeIx struct {
	Position InstructionPosition
	Program  string
	Accounts []string
	Data     []byte
}

func pumpfunNativeDebit(ix routeIx, children []routeIx, leg *RouteSwapLeg) *uint64 {
	if leg.Protocol != "PumpFun" || len(ix.Accounts) < 17 || ix.a(2) != routeWSOL || leg.InputAccount != leg.Trader || ix.Position.StackHeight == nil {
		return nil
	}
	recipients := map[string]bool{ix.a(6): true, ix.a(8): true, ix.a(10): true, ix.a(16): true}
	var total uint64
	paidPool := false
	for _, child := range children {
		if child.Program != zeroPubkey || len(child.Data) != 12 || binary.LittleEndian.Uint32(child.Data) != 2 || len(child.Accounts) < 2 || child.a(0) != leg.Trader {
			continue
		}
		if child.Position.StackHeight == nil || *child.Position.StackHeight != *ix.Position.StackHeight+1 {
			return nil
		}
		if !recipients[child.a(1)] {
			return nil
		}
		paidPool = paidPool || child.a(1) == leg.Pool
		amount := binary.LittleEndian.Uint64(child.Data[4:])
		if math.MaxUint64-total < amount {
			return nil
		}
		total += amount
	}
	if !paidPool {
		return nil
	}
	return &total
}

func (ix routeIx) a(i int) string {
	if i >= 0 && i < len(ix.Accounts) {
		return ix.Accounts[i]
	}
	return zeroPubkey
}
func routeString(m map[string]string, k string) *string {
	s, ok := m[k]
	if !ok || s == zeroPubkey {
		return nil
	}
	return &s
}
func tokenIx(ix routeIx) bool { return ix.Program == routeToken || ix.Program == routeToken2022 }
func checkedIx(ix routeIx) bool {
	return tokenIx(ix) && len(ix.Data) >= 10 && ix.Data[0] == 12 && len(ix.Accounts) >= 4
}
func feeIx(ix routeIx) bool {
	return ix.Program == routeToken2022 && len(ix.Data) >= 19 && ix.Data[0] == 26 && ix.Data[1] == 1 && len(ix.Accounts) >= 4
}
func routeTransfer(ix routeIx, mints map[string]string) *RouteTokenTransfer {
	if !tokenIx(ix) || len(ix.Data) == 0 {
		return nil
	}
	dest, offset := 0, 0
	var fee *uint64
	if checkedIx(ix) {
		dest = 2
		offset = 1
		if ix.Program == routeToken {
			v := uint64(0)
			fee = &v
		}
	} else if feeIx(ix) {
		dest = 2
		offset = 2
		v := binary.LittleEndian.Uint64(ix.Data[11:])
		fee = &v
	} else if ix.Data[0] == 3 && len(ix.Data) >= 9 && len(ix.Accounts) >= 3 {
		dest = 1
		offset = 1
		v := uint64(0)
		fee = &v
	} else {
		return nil
	}
	return &RouteTokenTransfer{ix.Position, ix.Program, ix.a(0), ix.a(dest), routeString(mints, ix.a(0)), binary.LittleEndian.Uint64(ix.Data[offset:]), fee}
}
func routeSwap(ix routeIx, mints map[string]string, graduated map[string]bool) *RouteSwapLeg {
	d, n := ix.Data, len(ix.Accounts)
	disc := ""
	if len(d) >= 8 {
		disc = hex.EncodeToString(d[:8])
	}
	a := ix.a
	protocol := ""
	switch ix.Program {
	case "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P":
		protocol = "PumpFun"
	case RAYDIUM_CLMM_PROGRAM_ID:
		protocol = "RaydiumClmm"
	case RAYDIUM_CPMM_PROGRAM_ID:
		protocol = "RaydiumCpmm"
	case RAYDIUM_AMM_V4_PROGRAM_ID:
		protocol = "RaydiumAmmV4"
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		protocol = "LaunchLab"
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		protocol = "OrcaWhirlpool"
	case METEORA_DAMM_V2_PROGRAM_ID:
		protocol = "MeteoraDammV2"
	case METEORA_DLMM_PROGRAM_ID:
		protocol = "MeteoraDlmm"
	case PUMPSWAP_PROGRAM_ID:
		protocol = "PumpSwap"
	}
	leg := &RouteSwapLeg{Position: ix.Position, Program: ix.Program, Protocol: protocol, AmountSpecifiedIsInput: true}
	offset, invert := 8, false
	pair := [2]string{}
	exact := true
	switch {
	case protocol == "PumpFun" && n == 17 && len(d) >= 24 && (disc == "07051dc4f5176550" || disc == "e1f7501ed5b38488" || disc == "1c92de7726c469d5"):
		buy := disc != "1c92de7726c469d5"
		exact = disc != "07051dc4f5176550"
		leg.Pool = a(5)
		leg.Trader = a(8)
		if buy {
			leg.InputAccount = a(10)
			leg.OutputAccount = a(9)
			pair = [2]string{a(2), a(1)}
		} else {
			leg.InputAccount = a(9)
			leg.OutputAccount = a(10)
			pair = [2]string{a(1), a(2)}
		}
		if a(2) == routeWSOL {
			if buy {
				leg.InputAccount = leg.Trader
			} else {
				leg.OutputAccount = leg.Trader
			}
		}
	case protocol == "PumpFun" && n >= 16 && (disc == "c2ab1c46684d5b2f" || disc == "b817ee6167c5d33d" || disc == "5df6823ce7e940b2"):
		buy := disc != "5df6823ce7e940b2"
		exact = disc != "b817ee6167c5d33d"
		leg.Pool = a(10)
		leg.Trader = a(13)
		if buy {
			leg.InputAccount = a(15)
			leg.OutputAccount = a(14)
			pair = [2]string{a(2), a(1)}
		} else {
			leg.InputAccount = a(14)
			leg.OutputAccount = a(15)
			pair = [2]string{a(1), a(2)}
		}
		if a(2) == routeWSOL {
			if buy {
				leg.InputAccount = leg.Trader
			} else {
				leg.OutputAccount = leg.Trader
			}
		}
	case protocol == "PumpFun" && n >= 12 && (disc == "38fc74089edfcd5f" || disc == "66063d1201daebea" || disc == "33e685a4017f83ad"):
		buy := disc != "33e685a4017f83ad"
		exact = disc != "66063d1201daebea"
		leg.Pool = a(3)
		leg.Trader = a(6)
		if buy {
			leg.InputAccount = a(6)
			leg.OutputAccount = a(5)
			pair = [2]string{routeWSOL, a(2)}
		} else {
			leg.InputAccount = a(5)
			leg.OutputAccount = a(6)
			pair = [2]string{a(2), routeWSOL}
		}
	case protocol == "RaydiumClmm" && (disc == "f8c69e91e17587c8" || disc == "2b04ed0b1ac91e62") && n >= 10 && len(d) >= 41:
		leg.Pool = a(2)
		leg.Trader = a(0)
		leg.InputAccount = a(3)
		leg.OutputAccount = a(4)
		exact = d[40] != 0
		if disc == "2b04ed0b1ac91e62" && n >= 13 {
			pair = [2]string{a(11), a(12)}
		}
	case protocol == "OrcaWhirlpool" && (disc == "f8c69e91e17587c8" || disc == "2b04ed0b1ac91e62") && len(d) >= 42:
		v2, dir := disc == "2b04ed0b1ac91e62", d[41] != 0
		x, y := 3, 5
		if v2 {
			if n < 15 {
				return nil
			}
			leg.Pool = a(4)
			leg.Trader = a(3)
			x, y = 7, 9
			pair = [2]string{a(5), a(6)}
		} else {
			if n < 11 {
				return nil
			}
			leg.Pool = a(2)
			leg.Trader = a(1)
		}
		if dir {
			leg.InputAccount = a(x)
			leg.OutputAccount = a(y)
		} else {
			leg.InputAccount = a(y)
			leg.OutputAccount = a(x)
			pair = [2]string{pair[1], pair[0]}
		}
		exact = d[40] != 0
	case protocol == "RaydiumCpmm" && (disc == "8fbe5adac41e33de" || disc == "37d96256a34ab4ad") && n >= 13:
		leg.Pool = a(3)
		leg.Trader = a(0)
		leg.InputAccount = a(4)
		leg.OutputAccount = a(5)
		pair = [2]string{a(10), a(11)}
		exact = disc == "8fbe5adac41e33de"
		invert = !exact
	case protocol == "MeteoraDammV2" && (disc == "f8c69e91e17587c8" || disc == "414b3f4ceb5b5b88") && n >= 11:
		if len(d) < 24 {
			return nil
		}
		if disc != "f8c69e91e17587c8" {
			if len(d) < 25 || d[24] > 2 {
				return nil
			}
			exact = d[24] != 2
		}
		leg.Pool = a(1)
		leg.Trader = a(8)
		leg.InputAccount = a(2)
		leg.OutputAccount = a(3)
	case protocol == "MeteoraDlmm" && (disc == "f8c69e91e17587c8" || disc == "414b3f4ceb5b5b88" || disc == "fa49652126cf4bb8" || disc == "2bd7f784893cf351") && n >= 11:
		leg.Pool = a(0)
		leg.Trader = a(10)
		leg.InputAccount = a(4)
		leg.OutputAccount = a(5)
		exact = disc == "f8c69e91e17587c8" || disc == "414b3f4ceb5b5b88"
		invert = !exact
	case protocol == "LaunchLab" || protocol == "PumpSwap":
		buy, valid := true, true
		if protocol == "LaunchLab" {
			switch disc {
			case "faea0d7bd59c13ec":
				exact = true
			case "18d3742869039938":
				exact = false
			case "9527de9bd37c981a":
				buy = false
				exact = true
			case "5fc8472208090ba6":
				buy = false
				exact = false
			default:
				valid = false
			}
			if n < 18 {
				return nil
			}
			leg.Pool = a(4)
			leg.Trader = a(0)
			leg.StonkFunMode = StonkFunModeFromPlatformConfig(a(3))
			pair = [2]string{a(10), a(9)}
		} else {
			switch disc {
			case "c62e1552b4d9e870", "c2ab1c46684d5b2f":
				exact = true
			case "66063d1201daebea", "b817ee6167c5d33d":
				exact = false
			case "33e685a4017f83ad", "5df6823ce7e940b2":
				buy = false
				exact = true
			default:
				valid = false
			}
			if n < 21 && !(n == 17 && (disc == "c2ab1c46684d5b2f" || disc == "b817ee6167c5d33d" || disc == "5df6823ce7e940b2")) {
				return nil
			}
			leg.Pool = a(0)
			leg.Trader = a(1)
			pair = [2]string{a(4), a(3)}
		}
		if !valid {
			return nil
		}
		if buy {
			leg.InputAccount = a(6)
			leg.OutputAccount = a(5)
		} else {
			leg.InputAccount = a(5)
			leg.OutputAccount = a(6)
			pair = [2]string{pair[1], pair[0]}
		}
	case protocol == "RaydiumAmmV4" && len(d) >= 17 && (d[0] == 9 || d[0] == 11 || d[0] == 16 || d[0] == 17):
		modern := d[0] == 16 || d[0] == 17
		leg.Pool = a(1)
		if modern {
			if n < 8 {
				return nil
			}
			leg.Trader = a(7)
			leg.InputAccount = a(5)
			leg.OutputAccount = a(6)
		} else {
			if n < 17 {
				return nil
			}
			leg.Trader = a(n - 1)
			leg.InputAccount = a(n - 3)
			leg.OutputAccount = a(n - 2)
		}
		offset = 1
		exact = d[0] == 9 || d[0] == 16
		invert = !exact
	default:
		return nil
	}
	if len(d) < offset+16 {
		return nil
	}
	first, second := binary.LittleEndian.Uint64(d[offset:]), binary.LittleEndian.Uint64(d[offset+8:])
	leg.SpecifiedAmount, leg.OtherAmountThreshold = first, second
	if invert {
		leg.SpecifiedAmount, leg.OtherAmountThreshold = second, first
	}
	leg.AmountSpecifiedIsInput = exact
	leg.InputMint = routeString(mints, leg.InputAccount)
	leg.OutputMint = routeString(mints, leg.OutputAccount)
	if leg.InputMint == nil && pair[0] != "" && pair[0] != zeroPubkey {
		v := pair[0]
		leg.InputMint = &v
	}
	if leg.OutputMint == nil && pair[1] != "" && pair[1] != zeroPubkey {
		v := pair[1]
		leg.OutputMint = &v
	}
	leg.StonkFunGraduated = protocol == "RaydiumCpmm" && graduated[leg.Pool]
	return leg
}
func routeDescendant(a, b InstructionPosition) bool {
	if a.OuterIndex != b.OuterIndex {
		return false
	}
	if a.InnerIndex == nil {
		return b.InnerIndex != nil
	}
	return b.InnerIndex != nil && *b.InnerIndex > *a.InnerIndex && a.StackHeight != nil && b.StackHeight != nil && *b.StackHeight > *a.StackHeight
}

type routeRPCInstruction struct {
	ProgramIDIndex *int   `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	StackHeight    *int   `json:"stackHeight"`
}
type routeRPCMessage struct {
	AccountKeys  []json.RawMessage     `json:"accountKeys"`
	Instructions []routeRPCInstruction `json:"instructions"`
}
type routeRPCBody struct {
	Signatures []string        `json:"signatures"`
	Message    routeRPCMessage `json:"message"`
}
type routeRPCMeta struct {
	Err    json.RawMessage `json:"err"`
	Loaded struct {
		Writable []string `json:"writable"`
		Readonly []string `json:"readonly"`
	} `json:"loadedAddresses"`
	Pre []struct {
		AccountIndex int    `json:"accountIndex"`
		Mint         string `json:"mint"`
	} `json:"preTokenBalances"`
	Post []struct {
		AccountIndex int    `json:"accountIndex"`
		Mint         string `json:"mint"`
	} `json:"postTokenBalances"`
	Inner []struct {
		Index        int                   `json:"index"`
		Instructions []routeRPCInstruction `json:"instructions"`
	} `json:"innerInstructions"`
}

func decodeRouteRPC(raw []byte) (routeRPCBody, *routeRPCMeta, []routeIx, uint64, error) {
	var root struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		return routeRPCBody{}, nil, nil, 0, err
	}
	if root.Result != nil {
		raw = root.Result
	}
	var tx struct {
		Transaction json.RawMessage `json:"transaction"`
		Meta        *routeRPCMeta   `json:"meta"`
		Slot        uint64          `json:"slot"`
	}
	if err := json.Unmarshal(raw, &tx); err != nil {
		return routeRPCBody{}, nil, nil, 0, err
	}
	if tx.Meta == nil {
		return routeRPCBody{}, nil, nil, 0, errors.New("transaction metadata missing")
	}
	if len(tx.Meta.Err) == 0 {
		return routeRPCBody{}, nil, nil, 0, errors.New("transaction execution status missing")
	}
	var body routeRPCBody
	bodyJSON := tx.Transaction
	if len(bodyJSON) > 0 && bodyJSON[0] == '[' {
		var encoded []string
		if err := json.Unmarshal(bodyJSON, &encoded); err != nil {
			return body, nil, nil, 0, err
		}
		if len(encoded) != 2 || encoded[1] != "base64" {
			return body, nil, nil, 0, errors.New("expected base64 encoding")
		}
		bytes, err := base64.StdEncoding.DecodeString(encoded[0])
		if err != nil {
			return body, nil, nil, 0, err
		}
		decoded, _, err := DecodeWireTransaction(bytes, 0, true)
		if err != nil {
			return body, nil, nil, 0, err
		}
		bodyJSON, err = json.Marshal(decoded)
		if err != nil {
			return body, nil, nil, 0, err
		}
	}
	if err := json.Unmarshal(bodyJSON, &body); err != nil {
		return body, nil, nil, 0, err
	}
	keys := []string{}
	for _, rawKey := range body.Message.AccountKeys {
		var k string
		if err := json.Unmarshal(rawKey, &k); err != nil {
			var v struct {
				Pubkey string `json:"pubkey"`
			}
			if err := json.Unmarshal(rawKey, &v); err != nil {
				return body, nil, nil, 0, err
			}
			k = v.Pubkey
		}
		keys = append(keys, k)
	}
	keys = append(keys, tx.Meta.Loaded.Writable...)
	keys = append(keys, tx.Meta.Loaded.Readonly...)
	seen := map[int]bool{}
	for _, group := range tx.Meta.Inner {
		if group.Index < 0 || group.Index >= len(body.Message.Instructions) || seen[group.Index] {
			return body, nil, nil, 0, errors.New("invalid compiled instruction group")
		}
		seen[group.Index] = true
		for _, ix := range group.Instructions {
			if ix.StackHeight != nil && *ix.StackHeight < 2 {
				return body, nil, nil, 0, errors.New("invalid compiled stack height")
			}
		}
	}
	for _, b := range append(tx.Meta.Pre, tx.Meta.Post...) {
		if b.AccountIndex < 0 || b.AccountIndex >= len(keys) {
			return body, nil, nil, 0, errors.New("invalid token balance account index")
		}
	}
	key := func(i int) string {
		if i >= 0 && i < len(keys) {
			return keys[i]
		}
		return zeroPubkey
	}
	out := []routeIx{}
	push := func(ix routeRPCInstruction, pos InstructionPosition) error {
		if ix.ProgramIDIndex == nil {
			return errors.New("expected compiled instructions")
		}
		if *ix.ProgramIDIndex < 0 || *ix.ProgramIDIndex >= len(keys) {
			return errors.New("invalid compiled program index")
		}
		d, err := base58.Decode(ix.Data)
		if err != nil {
			return err
		}
		a := make([]string, len(ix.Accounts))
		for i, k := range ix.Accounts {
			if k < 0 || k >= len(keys) {
				return errors.New("invalid compiled account index")
			}
			a[i] = key(k)
		}
		out = append(out, routeIx{pos, key(*ix.ProgramIDIndex), a, d})
		return nil
	}
	for i, ix := range body.Message.Instructions {
		height := 1
		if err := push(ix, InstructionPosition{i, nil, &height}); err != nil {
			return body, nil, nil, 0, err
		}
		for _, group := range tx.Meta.Inner {
			if group.Index == i {
				for j, child := range group.Instructions {
					index := j
					if err := push(child, InstructionPosition{i, &index, child.StackHeight}); err != nil {
						return body, nil, nil, 0, err
					}
				}
			}
		}
	}
	return body, tx.Meta, out, tx.Slot, nil
}

// AnalyzeRPCTransactionRoutes accepts compiled JSON or base64 RPC responses.
// The graduated pool list must come from independently verified migration evidence.
func AnalyzeRPCTransactionRoutes(raw []byte, graduatedPools []string) (*TransactionRoute, error) {
	body, meta, invocations, _, err := decodeRouteRPC(raw)
	if err != nil {
		return nil, err
	}
	mints := map[string]string{}
	keys := []string{}
	for _, rawKey := range body.Message.AccountKeys {
		var k string
		if json.Unmarshal(rawKey, &k) != nil {
			var v struct {
				Pubkey string `json:"pubkey"`
			}
			json.Unmarshal(rawKey, &v)
			k = v.Pubkey
		}
		keys = append(keys, k)
	}
	keys = append(keys, meta.Loaded.Writable...)
	keys = append(keys, meta.Loaded.Readonly...)
	for _, b := range append(meta.Pre, meta.Post...) {
		k, err := base58.Decode(b.Mint)
		if err != nil {
			return nil, err
		}
		if len(k) == 32 && b.AccountIndex >= 0 && b.AccountIndex < len(keys) {
			mints[keys[b.AccountIndex]] = b.Mint
		}
	}
	for _, ix := range invocations {
		if checkedIx(ix) || feeIx(ix) {
			mints[ix.a(0)] = ix.a(1)
			mints[ix.a(2)] = ix.a(1)
		}
		if tokenIx(ix) && len(ix.Accounts) >= 2 && ((len(ix.Data) == 1 && ix.Data[0] == 1) || (len(ix.Data) == 33 && (ix.Data[0] == 16 || ix.Data[0] == 18))) {
			mints[ix.a(0)] = ix.a(1)
		}
	}
	for {
		before := len(mints)
		for _, ix := range invocations {
			if tokenIx(ix) && len(ix.Data) >= 9 && ix.Data[0] == 3 && len(ix.Accounts) >= 3 {
				mint := mints[ix.a(0)]
				if mint == "" {
					mint = mints[ix.a(1)]
				}
				if mint != "" {
					if mints[ix.a(0)] == "" {
						mints[ix.a(0)] = mint
					}
					if mints[ix.a(1)] == "" {
						mints[ix.a(1)] = mint
					}
				}
			}
		}
		if len(mints) == before {
			break
		}
	}
	delete(mints, zeroPubkey)
	route := &TransactionRoute{Signature: base58.Encode(make([]byte, 64)), Succeeded: len(meta.Err) == 0 || string(meta.Err) == "null", Legs: []RouteSwapLeg{}, Transfers: []RouteTokenTransfer{}, NativeTokenActions: []RouteNativeTokenAction{}, UnknownInvocations: []RouteUnknownInvocation{}}
	if len(body.Signatures) > 0 {
		route.Signature = body.Signatures[0]
	}
	graduated := map[string]bool{}
	for _, p := range graduatedPools {
		graduated[p] = true
	}
	known := map[string]bool{routeToken: true, routeToken2022: true, zeroPubkey: true, "ComputeBudget111111111111111111111111111111": true, "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL": true, "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr": true}
	for i, ix := range invocations {
		if t := routeTransfer(ix, mints); t != nil {
			route.Transfers = append(route.Transfers, *t)
		}
		nested := []RouteTokenTransfer{}
		children := []routeIx{}
		knownSwap := false
		for _, child := range invocations[i+1:] {
			if !routeDescendant(ix.Position, child.Position) {
				break
			}
			if t := routeTransfer(child, mints); t != nil {
				// Keep all descendants separately for native payment attribution.
				nested = append(nested, *t)
			}
			children = append(children, child)
			if routeSwap(child, mints, graduated) != nil {
				knownSwap = true
			}
		}
		leg := routeSwap(ix, mints, graduated)
		if leg != nil {
			if route.Succeeded {
				var in, out uint64
				inCount, outCount := 0, 0
				inOK, outOK := true, true
				for _, t := range nested {
					if t.Source == leg.InputAccount {
						inCount++
						if math.MaxUint64-in < t.Amount {
							inOK = false
						} else {
							in += t.Amount
						}
					}
					if t.Destination == leg.OutputAccount {
						outCount++
						if t.WithheldFee == nil || *t.WithheldFee > t.Amount {
							outOK = false
						} else {
							net := t.Amount - *t.WithheldFee
							if math.MaxUint64-out < net {
								outOK = false
							} else {
								out += net
							}
						}
					}
				}
				if inCount > 0 && inOK {
					leg.ActualInputAmount = &in
				}
				if outCount > 0 && outOK {
					leg.ActualOutputAmount = &out
				}
			}
			if route.Succeeded && leg.Protocol == "PumpFun" && leg.InputMint != nil && *leg.InputMint == routeWSOL && leg.InputAccount == leg.Trader {
				leg.ActualInputAmount = pumpfunNativeDebit(ix, children, leg)
			}
			route.Legs = append(route.Legs, *leg)
		} else if !known[ix.Program] {
			route.UnknownInvocations = append(route.UnknownInvocations, RouteUnknownInvocation{ix.Position, ix.Program, len(nested) > 0, knownSwap})
		}
		target := ""
		var action any
		if ix.Program == zeroPubkey && len(ix.Data) == 12 && binary.LittleEndian.Uint32(ix.Data) == 2 && len(ix.Accounts) >= 2 {
			target = ix.a(1)
			action = map[string]any{"Fund": struct {
				Source   string `json:"source"`
				Lamports uint64 `json:"lamports,string"`
			}{ix.a(0), binary.LittleEndian.Uint64(ix.Data[4:])}}
		} else if tokenIx(ix) && len(ix.Data) == 1 {
			if ix.Data[0] == 17 && len(ix.Accounts) > 0 {
				target = ix.a(0)
				action = "SyncNative"
			} else if ix.Data[0] == 9 && len(ix.Accounts) >= 3 {
				target = ix.a(0)
				action = map[string]any{"Close": map[string]string{"destination": ix.a(1), "authority": ix.a(2)}}
			}
		}
		if target != "" && mints[target] == routeWSOL {
			route.NativeTokenActions = append(route.NativeTokenActions, RouteNativeTokenAction{ix.Position, target, action})
		}
	}
	return route, nil
}
