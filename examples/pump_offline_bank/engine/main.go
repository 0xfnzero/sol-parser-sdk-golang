// Go-native signed SBF execution; no RPC and no other-language subprocesses.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	litesvm "github.com/LiteSVM/litesvm-go"
	sol "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/sysvar"
	"github.com/mr-tron/base58"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
)

type wireInstruction struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}
type wireMessage struct {
	AccountKeys []string `json:"accountKeys"`
	Header      struct {
		NumRequiredSignatures       int
		NumReadonlySignedAccounts   int
		NumReadonlyUnsignedAccounts int
	}
	Instructions []wireInstruction `json:"instructions"`
}
type execution struct {
	Name     string   `json:"name"`
	Alt      bool     `json:"alt"`
	Reject   bool     `json:"reject"`
	Wire     string   `json:"wire"`
	Logs     []string `json:"logs"`
	Writable []string `json:"writable"`
	Readonly []string `json:"readonly"`
	Input    string   `json:"input"`
	Output   string   `json:"output"`
	Amount   uint64   `json:"amount"`
	Minimum  uint64   `json:"minimum"`
}
type rpcAccount struct {
	Lamports   uint64   `json:"lamports"`
	Data       []string `json:"data"`
	Owner      string   `json:"owner"`
	Executable bool     `json:"executable"`
	RentEpoch  uint64   `json:"rentEpoch"`
}
type bankCase struct {
	Name     string                `json:"name"`
	Accounts map[string]rpcAccount `json:"accounts"`
	Raw      struct {
		Transaction struct {
			Message wireMessage `json:"message"`
		} `json:"transaction"`
	} `json:"raw"`
}
type bundle struct {
	Schema   int               `json:"schema"`
	Clock    []int64           `json:"clock"`
	Programs map[string]string `json:"programs"`
	Cases    []bankCase        `json:"cases"`
}

var amm = sol.MustPublicKeyFromBase58("pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA")
var pump = sol.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
var ataProgram = sol.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")
var token = sol.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
var token22 = sol.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
var wallet = sol.PrivateKey(ed25519.NewKeyFromSeed(bytes.Repeat([]byte{117}, 32)))

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func check(ok bool, s string) {
	if !ok {
		panic(s)
	}
}
func b64(s string) []byte        { b, e := base64.StdEncoding.DecodeString(s); must(e); return b }
func key(s string) sol.PublicKey { return sol.MustPublicKeyFromBase58(s) }
func pda(seeds [][]byte, program sol.PublicKey) sol.PublicKey {
	p, _, e := sol.FindProgramAddress(seeds, program)
	must(e)
	return p
}
func ata(user, owner, mint sol.PublicKey) sol.PublicKey {
	return pda([][]byte{user[:], owner[:], mint[:]}, ataProgram)
}
func initialize(b bundle, c bankCase, alt bool) (*litesvm.LiteSVM, *sol.GenericInstruction, map[sol.PublicKey]sol.PublicKeySlice, map[sol.PublicKey]bool) {
	msg := c.Raw.Transaction.Message
	keys := make([]sol.PublicKey, len(msg.AccountKeys))
	for i, k := range msg.AccountKeys {
		keys[i] = key(k)
	}
	var original wireInstruction
	for _, ix := range msg.Instructions {
		if keys[ix.ProgramIDIndex] == amm {
			original = ix
			break
		}
	}
	old := keys[original.Accounts[0]]
	user := wallet.PublicKey()
	mapping := map[sol.PublicKey]sol.PublicKey{old: user}
	for _, program := range []sol.PublicKey{amm, pump} {
		mapping[pda([][]byte{[]byte("user_volume_accumulator"), old[:]}, program)] = pda([][]byte{[]byte("user_volume_accumulator"), user[:]}, program)
	}
	for address, a := range c.Accounts {
		data := b64(a.Data[0])
		owner := key(a.Owner)
		if (owner == token || owner == token22) && len(data) >= 165 && bytes.Equal(data[32:64], old[:]) {
			mint := sol.PublicKeyFromBytes(data[:32])
			if ata(old, owner, mint) == key(address) {
				mapping[key(address)] = ata(user, owner, mint)
			}
		}
	}
	mapped := func(k sol.PublicKey) sol.PublicKey {
		if m, ok := mapping[k]; ok {
			return m
		}
		return k
	}
	bank, e := litesvm.New()
	must(e)
	must(bank.SetSigverify(true))
	must(bank.SetBlockhashCheck(true))
	must(bank.SetLogBytesLimit(1000000))
	must(bank.SetClock(&sysvar.Clock{Slot: uint64(b.Clock[0]), EpochStartTimestamp: b.Clock[1], Epoch: uint64(b.Clock[2]), LeaderScheduleEpoch: uint64(b.Clock[3]), UnixTimestamp: b.Clock[4]}))
	// Stable insertion order makes execution traces independent of Go map iteration.
	programs := make([]string, 0, len(b.Programs))
	for p := range b.Programs {
		programs = append(programs, p)
	}
	sort.Strings(programs)
	for _, p := range programs {
		must(bank.AddProgram(key(p), b64(b.Programs[p])))
	}
	tracked := map[sol.PublicKey]bool{}
	for address, a := range c.Accounts {
		source := b64(a.Data[0])
		data := append([]byte{}, source...)
		for from, to := range mapping {
			for i := 0; i+32 <= len(source); i++ {
				if bytes.Equal(source[i:i+32], from[:]) {
					copy(data[i:i+32], to[:])
				}
			}
		}
		k := mapped(key(address))
		account, e := litesvm.NewAccount(a.Lamports, data, key(a.Owner), a.Executable, a.RentEpoch)
		must(e)
		must(bank.SetAccount(k, account))
		tracked[k] = true
	}
	must(bank.Airdrop(user, 500_000_000_000))
	h := msg.Header
	metas := make(sol.AccountMetaSlice, len(keys))
	for i, k := range keys {
		w := i < len(keys)-h.NumReadonlyUnsignedAccounts
		if i < h.NumRequiredSignatures {
			w = i < h.NumRequiredSignatures-h.NumReadonlySignedAccounts
		}
		metas[i] = &sol.AccountMeta{PublicKey: mapped(k), IsSigner: i < h.NumRequiredSignatures, IsWritable: w}
	}
	accounts := make(sol.AccountMetaSlice, len(original.Accounts))
	for i, k := range original.Accounts {
		accounts[i] = metas[k]
		tracked[accounts[i].PublicKey] = true
	}
	data, e := base58.Decode(original.Data)
	must(e)
	ix := sol.NewInstruction(amm, accounts, data)
	tracked[user] = true
	tables := map[sol.PublicKey]sol.PublicKeySlice{}
	if alt {
		addresses := sol.PublicKeySlice{}
		seen := map[sol.PublicKey]bool{}
		for _, a := range accounts {
			if !a.IsSigner && !seen[a.PublicKey] {
				addresses = append(addresses, a.PublicKey)
				seen[a.PublicKey] = true
			}
		}
		tableKey := sol.PublicKeyFromBytes(bytes.Repeat([]byte{118}, 32))
		data := make([]byte, 56)
		binary.LittleEndian.PutUint32(data, 1)
		binary.LittleEndian.PutUint64(data[4:], ^uint64(0))
		for _, k := range addresses {
			data = append(data, k[:]...)
		}
		rent, e := bank.MinimumBalanceForRentExemption(len(data))
		must(e)
		account, e := litesvm.NewAccount(rent, data, key("AddressLookupTab1e1111111111111111111111111"), false, 0)
		must(e)
		must(bank.SetAccount(tableKey, account))
		tables[tableKey] = addresses
		tracked[tableKey] = true
	}
	return bank, ix, tables, tracked
}
func state(bank *litesvm.LiteSVM, tracked map[sol.PublicKey]bool) map[sol.PublicKey]*sol.Account {
	m := map[sol.PublicKey]*sol.Account{}
	for k := range tracked {
		m[k] = bank.GetAccount(k)
	}
	return m
}
func execute(b bundle, c bankCase, alt, reject bool, oracle []map[string]string) execution {
	bank, ix, tables, tracked := initialize(b, c, alt)
	defer bank.Close()
	if reject {
		for i := 16; i < 24; i++ {
			ix.DataBytes[i] = 255
		}
	}
	hash, e := bank.LatestBlockhash()
	must(e)
	budget := []byte{2, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(budget[1:], 1_400_000)
	tx, e := sol.NewTransaction([]sol.Instruction{sol.NewInstruction(key("ComputeBudget111111111111111111111111111111"), nil, budget), ix}, hash, sol.TransactionPayer(wallet.PublicKey()), sol.TransactionAddressTables(tables))
	must(e)
	_, e = tx.Sign(func(k sol.PublicKey) *sol.PrivateKey {
		if k == wallet.PublicKey() {
			return &wallet
		}
		return nil
	})
	must(e)
	must(tx.VerifySignatures())
	raw, e := tx.MarshalBinary()
	must(e)
	for _, k := range tx.Message.AccountKeys {
		tracked[k] = true
	}
	before := state(bank, tracked)
	sim, e := bank.SimulateVersionedTransaction(raw)
	must(e)
	check(sim.IsOk() != reject, fmt.Sprintf("%s simulation: %s %v", c.Name, sim.Error(), sim.Logs()))
	check(reflect.DeepEqual(state(bank, tracked), before), "simulation mutated bank")
	commit, e := bank.SendVersionedTransaction(raw)
	must(e)
	check(commit.IsOk() != reject, "commit differs")
	after := state(bank, tracked)
	if reject {
		// The pinned Go VM reports index 0 for this parent error (Rust reports 1).
		// This harness checks its exact result; SDK wire parsing still checks outer index 1.
		check(sim.Error() == `{"InstructionError":[0,{"Custom":6004}]}`, sim.Error())
		for k, a := range before {
			if k != wallet.PublicKey() {
				check(reflect.DeepEqual(after[k], a), "parent rollback leaked writes")
			}
		}
		check(before[wallet.PublicKey()].Lamports-after[wallet.PublicKey()].Lamports == sim.Fee(), "failed transaction fee")
	} else {
		changed := false
		for k, a := range before {
			if k != wallet.PublicKey() && !reflect.DeepEqual(a, after[k]) {
				changed = true
			}
		}
		check(changed, "successful route changed no accounts")
	}
	var w, r []string
	for _, l := range tx.Message.GetAddressTableLookups() {
		table := tables[l.AccountKey]
		for _, i := range l.WritableIndexes {
			w = append(w, table[i].String())
		}
		for _, i := range l.ReadonlyIndexes {
			r = append(r, table[i].String())
		}
	}
	bad := append([]byte{}, raw...)
	bad[1] ^= 1
	failed, e := bank.SimulateVersionedTransaction(bad)
	must(e)
	check(failed.Error() == `"SignatureFailure"` && len(failed.Logs()) == 0, "signature rejection")
	if alt {
		for side := 0; side < 2; side++ {
			copyTx, e := sol.TransactionFromBytes(raw)
			must(e)
			lookups := copyTx.Message.GetAddressTableLookups()
			if side == 0 {
				lookups[0].WritableIndexes[0] = 255
			} else {
				lookups[0].ReadonlyIndexes[0] = 255
			}
			copyTx.Message.SetAddressTableLookups(lookups)
			_, e = copyTx.Sign(func(k sol.PublicKey) *sol.PrivateKey { return &wallet })
			must(e)
			invalid, e := copyTx.MarshalBinary()
			must(e)
			failed, e := bank.SimulateVersionedTransaction(invalid)
			must(e)
			check(failed.Error() == `"InvalidAddressLookupTableIndex"` && len(failed.Logs()) == 0, "ALT rejection")
		}
	}
	check(reflect.DeepEqual(state(bank, tracked), after), "pre-execution rejection mutated state")
	return execution{c.Name, alt, reject, base64.StdEncoding.EncodeToString(raw), sim.Logs(), w, r, ix.AccountValues[1].PublicKey.String(), ix.AccountValues[2].PublicKey.String(), binary.LittleEndian.Uint64(ix.DataBytes[8:]), binary.LittleEndian.Uint64(ix.DataBytes[16:])}
}
func runBank(configPath string) []execution {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "../../offline_bank")
	if configPath == "" {
		configPath = filepath.Join(root, "config.json")
	}
	data, e := os.ReadFile(configPath)
	must(e)
	var cfg struct {
		Snapshot    string `json:"snapshot"`
		SHA         string `json:"sha256"`
		Repetitions int    `json:"repetitions"`
	}
	must(json.Unmarshal(data, &cfg))
	payload, e := os.ReadFile(filepath.Join(root, cfg.Snapshot))
	must(e)
	check(fmt.Sprintf("%x", sha256.Sum256(payload)) == cfg.SHA, "snapshot checksum")
	gz, e := gzip.NewReader(bytes.NewReader(payload))
	must(e)
	data, e = io.ReadAll(gz)
	must(e)
	must(gz.Close())
	var b bundle
	must(json.Unmarshal(data, &b))
	check(b.Schema == 1 && cfg.Repetitions >= 2, "config/schema")
	data, e = os.ReadFile(filepath.Join(root, "expected.json"))
	must(e)
	var oracle map[string][]map[string]string
	must(json.Unmarshal(data, &oracle))
	var baseline []execution
	for pass := 0; pass < cfg.Repetitions; pass++ {
		var trace []execution
		for _, c := range b.Cases {
			for _, alt := range []bool{false, true} {
				for _, reject := range []bool{false, true} {
					trace = append(trace, execute(b, c, alt, reject, oracle[c.Name]))
				}
			}
		}
		check(len(trace) == 20, "matrix size")
		if pass > 0 {
			check(reflect.DeepEqual(trace, baseline), "non-repeatable bank")
		}
		baseline = trace
		fmt.Fprintf(os.Stderr, "Go bank engine pass %d: 20 signed routes executed\n", pass+1)
	}
	return baseline
}
func main() {
	config := ""
	if len(os.Args) > 1 {
		config = os.Args[1]
	}
	must(json.NewEncoder(os.Stdout).Encode(runBank(config)))
}
