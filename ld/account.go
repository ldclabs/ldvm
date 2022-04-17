// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
)

type Account struct {
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint8
	// keepers who can use this account, no more than 255
	// the account id must be one of them.
	Keepers []ids.ShortID

	// external assignment
	raw []byte
	ID  ids.ShortID
}

type jsonAccount struct {
	Address   string   `json:"address"`
	Nonce     uint64   `json:"nonce"`
	Balance   *big.Int `json:"balance"`
	Threshold uint8    `json:"threshold"`
	Keepers   []string `json:"keepers"`
}

func (a *Account) MarshalJSON() ([]byte, error) {
	if a == nil {
		return Null, nil
	}
	v := &jsonAccount{
		Address:   EthID(a.ID).String(),
		Nonce:     a.Nonce,
		Balance:   a.Balance,
		Threshold: a.Threshold,
		Keepers:   make([]string, len(a.Keepers)),
	}
	for i := range a.Keepers {
		v.Keepers[i] = EthID(a.Keepers[i]).String()
	}
	return json.Marshal(v)
}

func (a *Account) Copy() *Account {
	x := new(Account)
	*x = *a
	x.Balance = new(big.Int).Set(a.Balance)
	x.Keepers = make([]ids.ShortID, len(a.Keepers))
	copy(x.Keepers, a.Keepers)
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
	if a.Threshold < 1 || int(a.Threshold) > len(a.Keepers) {
		return fmt.Errorf("invalid threshold")
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
	if o.Nonce != a.Nonce {
		return false
	}
	if o.Balance == nil || a.Balance == nil {
		if o.Balance != a.Balance {
			return false
		}
	}
	if o.Balance.Cmp(a.Balance) != 0 {
		return false
	}
	if o.Threshold != a.Threshold {
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
		a.Nonce = v.Nonce.Value()
		a.Balance = ToBigInt(v.Balance)
		a.Threshold = v.Threshold.Value()
		if a.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		a.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindAccount")
}

func (a *Account) Marshal() ([]byte, error) {
	v := &bindAccount{
		Nonce:     FromUint64(a.Nonce),
		Balance:   FromBigInt(a.Balance),
		Threshold: FromUint8(a.Threshold),
		Keepers:   FromShortIDs(a.Keepers),
	}
	data, err := accountLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	a.raw = data
	return data, nil
}

type bindAccount struct {
	Nonce     Uint64
	Balance   []byte
	Threshold Uint8
	Keepers   [][]byte
}

var accountLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigInt bytes
	type Account struct {
		Nonce     Uint64 (rename "n")
		Balance   BigInt (rename "b")
		Threshold Uint8  (rename "th")
		Keepers   [ID20] (rename "ks")
	}
`
	builder, err := NewLDBuilder("Account", []byte(sch), (*bindAccount)(nil))
	if err != nil {
		panic(err)
	}
	accountLDBuilder = builder
}
