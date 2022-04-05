// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
)

var (
	LDC  = big.NewInt(1_000_000_000)
	Gwei = big.NewInt(1)
)

type Account struct {
	Nonce     uint64
	Balance   *big.Int      // 最小单位 gwei, decimals 9
	Threshold uint8         // account 消费时要求的签名阈值，account 自身必须签名，最小值为 1，最大值为 len(guardians) + 1
	Guardians []ids.ShortID // 本账号之外的其它监护账号，不能大于 15
	raw       []byte
}

func (a *Account) Copy() *Account {
	x := new(Account)
	*x = *a
	x.Balance = new(big.Int).Set(a.Balance)
	x.Guardians = make([]ids.ShortID, len(a.Guardians))
	copy(x.Guardians, a.Guardians)
	x.raw = make([]byte, len(a.raw))
	copy(x.raw, a.raw)
	return x
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	if a.Balance == nil || a.Balance.Sign() < 0 {
		return fmt.Errorf("invalid account Balance")
	}
	if len(a.Guardians) > 15 {
		return fmt.Errorf("too many account Guardians")
	}
	if a.Threshold < 1 || int(a.Threshold) > len(a.Guardians)+1 {
		return fmt.Errorf("invalid account Threshold")
	}
	for _, id := range a.Guardians {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid account Guardian")
		}
	}
	if _, err := a.Marshal(); err != nil {
		return fmt.Errorf("account marshal error: %v", err)
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
	if len(o.Guardians) != len(a.Guardians) {
		return false
	}
	for i := range a.Guardians {
		if o.Guardians[i] != a.Guardians[i] {
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
		if a.Guardians, err = ToShortIDs(v.Guardians); err != nil {
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
		Guardians: FromShortIDs(a.Guardians),
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
	Guardians [][]byte
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
		Guardians [ID20] (rename "gs")
	}
`
	builder, err := NewLDBuilder("Account", []byte(sch), (*bindAccount)(nil))
	if err != nil {
		panic(err)
	}
	accountLDBuilder = builder
}
