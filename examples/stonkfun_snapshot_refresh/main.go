// go run ./examples/stonkfun_snapshot_refresh snapshots.json [--require-pool-update]
// Reads GRPC_URL/GRPC_TOKEN, refreshes saved LaunchLab/CPMM snapshots, never sends.
// Unchanged cold-bootstrap accounts retain their age and may be rejected by trade.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
	"strconv"
	"time"
)

const clockKey = "SysvarC1ock11111111111111111111111111111111"

type savedAccount struct {
	Pubkey       string `json:"pubkey"`
	Owner        string `json:"owner"`
	Data         string `json:"data"`
	Slot         string `json:"slot"`
	WriteVersion string `json:"write_version"`
}

func applyRawSnapshot(old savedAccount, raw *solparser.RawAccountSnapshotEvent) (savedAccount, bool, error) {
	if old.Pubkey != raw.Account.Pubkey {
		return old, false, errors.New("snapshot identity mismatch")
	}
	slot, e := strconv.ParseUint(old.Slot, 10, 64)
	if e != nil {
		return old, false, e
	}
	version, e := strconv.ParseUint(old.WriteVersion, 10, 64)
	if e != nil {
		return old, false, e
	}
	newSlot, newVersion := raw.Metadata.Slot, raw.WriteVersion
	if newSlot < slot || newSlot == slot && newVersion < version {
		return old, false, nil
	}
	data := raw.Account.Data
	if raw.Account.Lamports == 0 {
		data = nil
	}
	if newSlot == slot && newVersion == version {
		previous, e := base64.StdEncoding.DecodeString(old.Data)
		if e != nil {
			return old, false, e
		}
		if old.Owner != raw.Account.Owner || !bytes.Equal(previous, data) {
			return old, false, errors.New("conflicting account version; select a fork explicitly")
		}
		return old, false, nil
	}
	return savedAccount{raw.Account.Pubkey, raw.Account.Owner, base64.StdEncoding.EncodeToString(data), strconv.FormatUint(newSlot, 10), strconv.FormatUint(newVersion, 10)}, true, nil
}
func run() error {
	if len(os.Args) < 2 || os.Getenv("GRPC_URL") == "" {
		return errors.New("provide snapshot path and GRPC_URL")
	}
	path := os.Args[1]
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	var snapshot map[string]json.RawMessage
	if e = json.Unmarshal(b, &snapshot); e != nil {
		return e
	}
	names := map[string]string{}
	saved := map[string]savedAccount{}
	keys := []string{clockKey}
	var listed []savedAccount
	if raw, ok := snapshot["accounts"]; ok {
		if e = json.Unmarshal(raw, &listed); e != nil {
			return e
		}
		for i, a := range listed {
			if a.Pubkey == clockKey {
				continue
			}
			n := strconv.Itoa(i)
			names[a.Pubkey] = n
			saved[n] = a
			keys = append(keys, a.Pubkey)
		}
	} else {
		for _, n := range []string{"pool", "global", "platform", "config", "base_mint", "quote_mint", "base_vault", "quote_vault"} {
			if raw, ok := snapshot[n]; ok && len(raw) > 0 && raw[0] == '{' {
				var a savedAccount
				if e = json.Unmarshal(raw, &a); e != nil {
					return e
				}
				names[a.Pubkey] = n
				saved[n] = a
				keys = append(keys, a.Pubkey)
			}
		}
	}
	pools := map[string]bool{}
	var legs []struct{ Pool string }
	if raw, ok := snapshot["legs"]; ok {
		if e = json.Unmarshal(raw, &legs); e != nil {
			return e
		}
		for _, h := range legs {
			pools[h.Pool] = true
		}
	}
	if raw, ok := snapshot["pool"]; ok {
		var pool string
		if json.Unmarshal(raw, &pool) != nil {
			var a savedAccount
			if e = json.Unmarshal(raw, &a); e != nil {
				return e
			}
			pool = a.Pubkey
		}
		pools[pool] = true
	}
	updatedPools := map[string]bool{}
	client := solparser.NewYellowstoneGrpc(os.Getenv("GRPC_URL"))
	client.SetXToken(os.Getenv("GRPC_TOKEN"))
	if e = client.Connect(); e != nil {
		return e
	}
	defer client.Disconnect()
	sub, e := client.SubscribeDexEvents(nil, []solparser.AccountFilter{{Account: keys}}, solparser.EventTypeFilterIncludeOnly([]solparser.EventType{solparser.EventTypeAccountRawSnapshot}))
	if e != nil {
		return e
	}
	defer sub.Cancel()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	var clock *savedAccount
	poolUpdated := false
	requirePool := len(os.Args) > 2 && os.Args[2] == "--require-pool-update"
	errorsChannel := sub.Errors
	put := func(name string, value any) { snapshot[name], _ = json.Marshal(value) }
	for {
		select {
		case <-timer.C:
			return errors.New("timed out waiting for required snapshots")
		case err, ok := <-errorsChannel:
			if !ok {
				errorsChannel = nil
			}
			if ok && err != nil {
				return err
			}
		case ev, ok := <-sub.Events:
			if !ok {
				return errors.New("subscription closed before required snapshots")
			}
			raw, ok := ev.Data.(*solparser.RawAccountSnapshotEvent)
			if !ok {
				continue
			}
			if name, ok := names[raw.Account.Pubkey]; ok {
				updated, changed, e := applyRawSnapshot(saved[name], raw)
				if e != nil {
					return e
				}
				saved[name] = updated
				if listed != nil {
					i, err := strconv.Atoi(name)
					if err != nil {
						return err
					}
					listed[i] = updated
					put("accounts", listed)
				} else {
					put(name, updated)
				}
				if pools[raw.Account.Pubkey] && changed {
					updatedPools[raw.Account.Pubkey] = true
				}
				poolUpdated = len(pools) > 0 && len(updatedPools) == len(pools)
			} else if raw.Account.Pubkey == clockKey {
				if raw.Account.Lamports == 0 || len(raw.Account.Data) != 40 {
					return errors.New("invalid Clock account")
				}
				previous := savedAccount{Pubkey: clockKey, Owner: raw.Account.Owner, Slot: "0", WriteVersion: "0"}
				if clock != nil {
					previous = *clock
				}
				updated, changed, e := applyRawSnapshot(previous, raw)
				if e != nil {
					return e
				}
				if !changed {
					continue
				}
				clock = &updated
				if listed != nil {
					for i, a := range listed {
						if a.Pubkey == clockKey {
							listed[i] = updated
						}
					}
					put("accounts", listed)
				}
				data := raw.Account.Data
				put("read_slot", strconv.FormatUint(binary.LittleEndian.Uint64(data), 10))
				put("epoch", strconv.FormatUint(binary.LittleEndian.Uint64(data[16:]), 10))
				put("unix_timestamp", strconv.FormatInt(int64(binary.LittleEndian.Uint64(data[32:])), 10))
			}
			if clock != nil && (poolUpdated || !requirePool) {
				commitment := solparser.CommitmentLevelConfirmed
				hash, e := client.GetLatestBlockhash(&commitment)
				if e != nil {
					return e
				}
				put("recent_blockhash", hash.Blockhash)
				output, e := json.MarshalIndent(snapshot, "", "  ")
				if e != nil {
					return e
				}
				if e = os.WriteFile(path, append(output, '\n'), 0600); e != nil {
					return e
				}
				fmt.Printf("{\"slot\":%s,\"epoch\":%s,\"blockhash_slot\":%d,\"pool_updated\":%t}\n", snapshot["read_slot"], snapshot["epoch"], hash.Slot, poolUpdated)
				return nil
			}
		}
	}
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
