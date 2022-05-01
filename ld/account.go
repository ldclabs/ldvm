// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

// AccountType is an uint8 representing the type of account
type AccountType uint8

const (
	NativeAccount AccountType = iota
	TokenAccount              // The first 10 bytes of account address must be 0
	StakeAccount              // The first byte of account address must be $
)

func AccountTypeString(t AccountType) string {
	switch t {
	case NativeAccount:
		return "NativeAccount"
	case TokenAccount:
		return "TokenAccount"
	case StakeAccount:
		return "StakeAccount"
	default:
		return "TypeUnknown"
	}
}

type Account struct {
	Type AccountType
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int
	// M of N threshold signatures, aka MultiSig: threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint8
	// keepers who can use this account, no more than 255
	// the account id must be one of them.
	Keepers []ids.ShortID

	LockTime       uint64              // only used with StakeAccount
	DelegationFee  uint64              // only used with StakeAccount, 1_000 == 100%, should be in [1, 500]
	MaxTotalSupply *big.Int            // only used with TokenAccount
	NonceTable     map[uint64][]uint64 // map[expire][]nonce
	Ledger         map[ids.ShortID]*big.Int

	// external assignment
	raw []byte
	ID  ids.ShortID
}

type jsonAccount struct {
	Type           string              `json:"type"`
	Address        string              `json:"address"`
	Nonce          uint64              `json:"nonce"`
	Balance        *big.Int            `json:"balance"`
	Threshold      uint8               `json:"threshold"`
	Keepers        []string            `json:"keepers"`
	LockTime       uint64              `json:"lockTime,omitempty"`
	DelegationFee  uint64              `json:"delegationFee,omitempty"`
	MaxTotalSupply *big.Int            `json:"maxTotalSupply,omitempty"`
	NonceTable     map[string][]uint64 `json:"nonceTable"`
	Ledger         map[string]*big.Int `json:"ledger"`
}

func (a *Account) MarshalJSON() ([]byte, error) {
	if a == nil {
		return util.Null, nil
	}
	v := &jsonAccount{
		Type:           AccountTypeString(a.Type),
		Address:        util.EthID(a.ID).String(),
		Nonce:          a.Nonce,
		Balance:        a.Balance,
		Threshold:      a.Threshold,
		Keepers:        make([]string, len(a.Keepers)),
		LockTime:       a.LockTime,
		DelegationFee:  a.DelegationFee,
		MaxTotalSupply: a.MaxTotalSupply,
		NonceTable:     make(map[string][]uint64, len(a.NonceTable)),
		Ledger:         make(map[string]*big.Int, len(a.Ledger)),
	}
	for i := range a.Keepers {
		v.Keepers[i] = util.EthID(a.Keepers[i]).String()
	}
	for k, arr := range a.NonceTable {
		v.NonceTable[strconv.FormatUint(k, 10)] = arr
	}
	for k := range a.Ledger {
		str := util.TokenSymbol(k).String()
		if str == "" {
			str = util.EthID(k).String()
		}
		v.Ledger[str] = a.Ledger[k]
	}
	return json.Marshal(v)
}

func (a *Account) Copy() *Account {
	x := new(Account)
	*x = *a
	x.Balance = new(big.Int).Set(a.Balance)
	x.Keepers = make([]ids.ShortID, len(a.Keepers))
	copy(x.Keepers, a.Keepers)
	x.NonceTable = make(map[uint64][]uint64, len(a.NonceTable))
	for k, v := range a.NonceTable {
		x.NonceTable[k] = make([]uint64, len(v))
		for i := range v {
			x.NonceTable[k][i] = v[i]
		}
	}
	x.Ledger = make(map[ids.ShortID]*big.Int, len(a.Ledger))
	for k, v := range a.Ledger {
		x.Ledger[k] = new(big.Int).Set(v)
	}
	if a.MaxTotalSupply != nil {
		x.MaxTotalSupply = new(big.Int).Set(a.MaxTotalSupply)
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	if a == nil {
		return fmt.Errorf("invalid Account")
	}

	if a.Balance == nil || a.Balance.Sign() < 0 {
		return fmt.Errorf("invalid balance")
	}
	if len(a.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(a.Threshold) > len(a.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	if a.NonceTable == nil {
		return fmt.Errorf("invalid nonceTable")
	}
	if a.Ledger == nil {
		return fmt.Errorf("invalid ledger")
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.LockTime != 0 {
			return fmt.Errorf("invalid lockTime, should be 0")
		}
		if a.DelegationFee != 0 {
			return fmt.Errorf("invalid delegationFee, should be 0")
		}
	case TokenAccount:
		if a.LockTime != 0 {
			return fmt.Errorf("invalid lockTime, should be 0")
		}
		if a.DelegationFee != 0 {
			return fmt.Errorf("invalid delegationFee, should be 0")
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return fmt.Errorf("invalid maxTotalSupply")
		}
	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.LockTime == 0 {
			return fmt.Errorf("invalid lockTime, should not be 0")
		}
		if a.DelegationFee == 0 || a.DelegationFee > 500 {
			return fmt.Errorf("invalid delegationFee, should be in [1, 500]")
		}
	default:
		return fmt.Errorf("invalid type")
	}

	if _, err := a.Marshal(); err != nil {
		return fmt.Errorf("Account marshal error: %v", err)
	}
	return nil
}

func (a *Account) Equal(o *Account) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(a.raw) > 0 {
		return bytes.Equal(o.raw, a.raw)
	}
	if o.Type != a.Type {
		return false
	}
	if o.Nonce != a.Nonce {
		return false
	}
	if o.Balance == nil || a.Balance == nil || o.Balance.Cmp(a.Balance) != 0 {
		return false
	}
	if o.MaxTotalSupply == nil || a.MaxTotalSupply == nil {
		if o.MaxTotalSupply != a.MaxTotalSupply {
			return false
		}
	} else if o.MaxTotalSupply.Cmp(a.MaxTotalSupply) != 0 {
		return false
	}
	if o.Threshold != a.Threshold {
		return false
	}
	if o.LockTime != a.LockTime {
		return false
	}
	if o.DelegationFee != a.DelegationFee {
		return false
	}
	if len(o.Keepers) != len(a.Keepers) {
		return false
	}
	for i := range a.Keepers {
		if o.Keepers[i] != a.Keepers[i] {
			return false
		}
	}
	if len(o.NonceTable) != len(a.NonceTable) {
		return false
	}
	for k, v := range a.NonceTable {
		if len(o.NonceTable[k]) != len(v) {
			return false
		}
		for i := range v {
			if o.NonceTable[k][i] != v[i] {
				return false
			}
		}
	}
	if len(o.Ledger) != len(a.Ledger) {
		return false
	}
	for id := range a.Ledger {
		if o.Ledger[id] == nil || a.Ledger[id] == nil || o.Ledger[id].Cmp(a.Ledger[id]) != 0 {
			return false
		}
	}
	return true
}

func (a *Account) Bytes() []byte {
	if len(a.raw) == 0 {
		if _, err := a.Marshal(); err != nil {
			panic(err)
		}
	}

	return a.raw
}

func (a *Account) Unmarshal(data []byte) error {
	p, err := accountLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindAccount); ok {
		if !v.Type.Valid() || !v.Threshold.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}
		if !v.Nonce.Valid() ||
			!v.LockTime.Valid() ||
			!v.DelegationFee.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}

		a.Type = AccountType(v.Type.Value())
		a.Nonce = v.Nonce.Value()
		a.Balance = v.Balance.Value()
		a.Threshold = v.Threshold.Value()
		a.MaxTotalSupply = v.MaxTotalSupply.Value()
		a.LockTime = v.LockTime.Value()
		a.DelegationFee = v.DelegationFee.Value()
		a.NonceTable = make(map[uint64][]uint64, len(a.NonceTable))

		if a.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		now := uint64(time.Now().Unix())
		for _, data := range v.NonceTable {
			uu, err := ReadUint64s(data)
			if err != nil {
				return err
			}
			if len(uu) < 1 {
				return fmt.Errorf("unmarshal NonceTable error")
			}
			if exp := uu[0]; exp >= now && len(uu) > 1 {
				a.NonceTable[exp] = uu[1:]
			}
		}

		a.Ledger = make(map[ids.ShortID]*big.Int, len(v.Ledger))
		for _, data := range v.Ledger {
			if len(data) < 20 {
				return fmt.Errorf("unmarshal error: invalid ledger data")
			}
			id := ids.ShortID{}
			copy(id[:], data[:20])
			a.Ledger[id] = new(big.Int).SetBytes(data[20:])
		}
		a.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindAccount")
}

func (a *Account) Marshal() ([]byte, error) {
	ba := &bindAccount{
		Type:           FromUint8(uint8(a.Type)),
		Nonce:          FromUint64(a.Nonce),
		Balance:        FromUint(a.Balance),
		Threshold:      FromUint8(a.Threshold),
		Keepers:        FromShortIDs(a.Keepers),
		LockTime:       FromUint64(a.LockTime),
		DelegationFee:  FromUint64(a.DelegationFee),
		MaxTotalSupply: FromUint(a.MaxTotalSupply),
		NonceTable:     make([][]byte, 0, len(a.NonceTable)),
		Ledger:         make([][]byte, 0, len(a.Ledger)),
	}
	now := uint64(time.Now().Unix())
	buf := new(bytes.Buffer)
	var err error
	for k, uu := range a.NonceTable {
		if k < now || len(uu) == 0 {
			continue
		}
		if err = WriteUint64s(buf, k, uu...); err != nil {
			return nil, err
		}
		ba.NonceTable = append(ba.NonceTable, buf.Bytes())
		buf.Reset()
	}

	for k, v := range a.Ledger {
		b := FromUint(v)
		data := make([]byte, 20+len(b))
		copy(data, k[:])
		copy(data[20:], b)
		ba.Ledger = append(ba.Ledger, data)
	}
	data, err := accountLDBuilder.Marshal(ba)
	if err != nil {
		return nil, err
	}
	a.raw = data
	return data, nil
}

type bindAccount struct {
	Type           Uint8
	Nonce          Uint64
	Balance        BigUint
	Threshold      Uint8
	Keepers        [][]byte
	NonceTable     [][]byte
	Ledger         [][]byte
	LockTime       Uint64
	DelegationFee  Uint64
	MaxTotalSupply BigUint
}

var accountLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type Account struct {
		Type           Uint8   (rename "t")
		Nonce          Uint64  (rename "n")
		Balance        BigUint (rename "b")
		Threshold      Uint8   (rename "th")
		Keepers        [ID20]  (rename "kp")
		NonceTable     [Bytes] (rename "nt")
		Ledger         [Bytes] (rename "lg")
		LockTime       Uint64  (rename "lt")
		DelegationFee  Uint64  (rename "df")
		MaxTotalSupply BigUint (rename "mts")
	}
`
	builder, err := NewLDBuilder("Account", []byte(sch), (*bindAccount)(nil))
	if err != nil {
		panic(err)
	}
	accountLDBuilder = builder
}
