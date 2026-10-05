package solparser

import (
	"bytes"
	"errors"
	"github.com/mr-tron/base58"
	"sort"
	"sync"
)

type StonkFunGraduatedPool struct {
	CurvePool          string `json:"curve_pool"`
	Pool               string `json:"pool"`
	BaseMint           string `json:"base_mint"`
	QuoteMint          string `json:"quote_mint"`
	PlatformConfig     string `json:"platform_config"`
	MigrationSignature string `json:"migration_signature"`
	MigrationSlot      uint64 `json:"migration_slot,string"`
}

func (p StonkFunGraduatedPool) Mode() *string {
	return StonkFunModeFromPlatformConfig(p.PlatformConfig)
}
func (p StonkFunGraduatedPool) Validate() error {
	if p.Mode() == nil || p.CurvePool == p.Pool || p.BaseMint == p.QuoteMint {
		return errors.New("invalid registry identity")
	}
	for _, key := range []string{p.CurvePool, p.Pool, p.BaseMint, p.QuoteMint} {
		b, e := base58.Decode(key)
		if e != nil || len(b) != 32 || key == zeroPubkey {
			return errors.New("invalid registry key")
		}
	}
	b, e := base58.Decode(p.MigrationSignature)
	if e != nil || len(b) != 64 {
		return errors.New("invalid migration signature")
	}
	return nil
}

type StonkFunPoolRegistry struct {
	mu    sync.RWMutex
	pools map[string]StonkFunGraduatedPool
}

func NewStonkFunPoolRegistry(entries []StonkFunGraduatedPool) (*StonkFunPoolRegistry, error) {
	r := &StonkFunPoolRegistry{pools: map[string]StonkFunGraduatedPool{}}
	for _, p := range entries {
		if e := p.Validate(); e != nil {
			return nil, e
		}
		if _, ok := r.pools[p.Pool]; ok {
			return nil, errors.New("duplicate registry entry")
		}
		r.pools[p.Pool] = p
	}
	return r, nil
}
func (r *StonkFunPoolRegistry) Get(pool string) (StonkFunGraduatedPool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.pools[pool]
	return p, ok
}
func (r *StonkFunPoolRegistry) Pools() []StonkFunGraduatedPool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := []string{}
	for k := range r.pools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]StonkFunGraduatedPool, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.pools[k])
	}
	return out
}
func (r *StonkFunPoolRegistry) VerifiedCpmmPools() []string {
	out := []string{}
	for _, p := range r.Pools() {
		out = append(out, p.Pool)
	}
	return out
}
func (r *StonkFunPoolRegistry) PoolsForBaseMint(mint string) []StonkFunGraduatedPool {
	out := []StonkFunGraduatedPool{}
	for _, p := range r.Pools() {
		if p.BaseMint == mint {
			out = append(out, p)
		}
	}
	return out
}

// ObserveMigrations applies an entire batch atomically; replay is idempotent.
func (r *StonkFunPoolRegistry) ObserveMigrations(entries []StonkFunGraduatedPool, succeeded bool) (int, error) {
	if !succeeded {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	pending := map[string]StonkFunGraduatedPool{}
	for _, p := range entries {
		if e := p.Validate(); e != nil {
			return 0, e
		}
		prior, ok := pending[p.Pool]
		if !ok {
			prior, ok = r.pools[p.Pool]
		}
		if ok && prior != p {
			return 0, errors.New("conflicting migration provenance")
		}
		pending[p.Pool] = p
	}
	if r.pools == nil {
		r.pools = map[string]StonkFunGraduatedPool{}
	}
	n := 0
	for k, p := range pending {
		if _, ok := r.pools[k]; !ok {
			n++
		}
		r.pools[k] = p
	}
	return n, nil
}
func (r *StonkFunPoolRegistry) ObserveRPCTransaction(raw []byte) (int, error) {
	body, meta, ixs, slot, err := decodeRouteRPC(raw)
	if err != nil {
		return 0, err
	}
	if len(meta.Err) > 0 && string(meta.Err) != "null" {
		return 0, nil
	}
	entries := []StonkFunGraduatedPool{}
	for _, ix := range ixs {
		if ix.Program != RAYDIUM_LAUNCHLAB_PROGRAM_ID || len(ix.Data) < 8 || !bytes.Equal(ix.Data[:8], []byte{136, 92, 200, 103, 28, 218, 144, 140}) || len(ix.Accounts) < 28 {
			continue
		}
		a := ix.a
		if a(4) != RAYDIUM_CPMM_PROGRAM_ID || StonkFunModeFromPlatformConfig(a(3)) == nil {
			continue
		}
		if a(17) == zeroPubkey || a(5) == zeroPubkey || a(1) == zeroPubkey || a(2) == zeroPubkey || a(17) == a(5) || a(1) == a(2) {
			continue
		}
		if len(body.Signatures) == 0 {
			return 0, errors.New("migration signature missing")
		}
		entries = append(entries, StonkFunGraduatedPool{a(17), a(5), a(1), a(2), a(3), body.Signatures[0], slot})
	}
	return r.ObserveMigrations(entries, true)
}
