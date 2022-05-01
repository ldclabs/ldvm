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

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

// TxMinter
type TxMinter struct {
	Threshold     uint8
	Keepers       []ids.ShortID
	LockTime      uint64   // only used with StakeAccount
	DelegationFee uint64   // only used with StakeAccount, 1_000 == 100%, should be in [1, 500]
	Amount        *big.Int // only used with TokenAccount and StakeAccount
	Name          string
	Message       string

	// external assignment
	raw []byte
}

type jsonTxMinter struct {
	Threshold     uint8    `json:"threshold,omitempty"`
	Keepers       []string `json:"keepers,omitempty"`
	LockTime      uint64   `json:"lockTime,omitempty"`
	DelegationFee uint64   `json:"delegationFee,omitempty"`
	Amount        *big.Int `json:"amount,omitempty"`
	Name          string   `json:"name,omitempty"`
	Message       string   `json:"message,omitempty"`
}

func (t *TxMinter) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxMinter{
		Threshold:     t.Threshold,
		LockTime:      t.LockTime,
		DelegationFee: t.DelegationFee,
		Amount:        t.Amount,
		Name:          t.Name,
		Message:       t.Message,
	}
	if len(t.Keepers) > 0 {
		v.Keepers = make([]string, len(t.Keepers))
		for i := range t.Keepers {
			v.Keepers[i] = util.EthID(t.Keepers[i]).String()
		}
	}
	return json.Marshal(v)
}

func (t *TxMinter) Copy() *TxMinter {
	x := new(TxMinter)
	*x = *t
	if t.Amount != nil {
		x.Amount = new(big.Int).Set(t.Amount)
	}
	x.Keepers = make([]ids.ShortID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxMinter is well-formed.
func (t *TxMinter) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxMinter")
	}
	if t.Name != "" && !util.ValidName(t.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(t.Name))
	}
	if t.Message != "" && !util.ValidMessage(t.Message) {
		return fmt.Errorf("invalid message string %s", strconv.Quote(t.Message))
	}

	if t.Amount != nil && t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid keeper")
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxMinter marshal error: %v", err)
	}
	return nil
}

func (t *TxMinter) Equal(o *TxMinter) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Threshold != t.Threshold {
		return false
	}
	if o.LockTime != t.LockTime {
		return false
	}
	if o.DelegationFee != t.DelegationFee {
		return false
	}
	if o.Name != t.Name {
		return false
	}
	if o.Message != t.Message {
		return false
	}
	if o.Amount == nil || t.Amount == nil {
		if o.Amount != t.Amount {
			return false
		}
	} else if o.Amount.Cmp(t.Amount) != 0 {
		return false
	}
	if len(o.Keepers) != len(t.Keepers) {
		return false
	}
	for i := range t.Keepers {
		if o.Keepers[i] != t.Keepers[i] {
			return false
		}
	}
	return true
}

func (t *TxMinter) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxMinter) Unmarshal(data []byte) error {
	p, err := txMinterLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxMinter); ok {
		if !v.Threshold.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}
		if !v.LockTime.Valid() ||
			!v.DelegationFee.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}

		t.Threshold = v.Threshold.Value()
		t.LockTime = v.LockTime.Value()
		t.DelegationFee = v.DelegationFee.Value()
		t.Amount = v.Amount.PtrValue()
		if t.Keepers, err = PtrToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if v.Name != nil {
			t.Name = *v.Name
		}
		if v.Message != nil {
			t.Message = *v.Message
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxMinter")
}

func (t *TxMinter) Marshal() ([]byte, error) {
	v := &bindTxMinter{
		Threshold:     PtrFromUint8(t.Threshold),
		Keepers:       PtrFromShortIDs(t.Keepers),
		LockTime:      PtrFromUint64(t.LockTime),
		DelegationFee: PtrFromUint64(t.DelegationFee),
		Amount:        PtrFromUint(t.Amount),
	}
	if t.Name != "" {
		v.Name = &t.Name
	}
	if t.Message != "" {
		v.Message = &t.Message
	}
	data, err := txMinterLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindTxMinter struct {
	Threshold     *Uint8
	Keepers       *[][]byte
	LockTime      *Uint64
	DelegationFee *Uint64
	Amount        *BigUint
	Name          *string
	Message       *string
}

var txMinterLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type TxMinter struct {
		Threshold     nullable Uint8   (rename "th")
		Keepers       nullable [ID20]  (rename "kp")
		LockTime      nullable Uint64  (rename "lt")
		DelegationFee nullable Uint64  (rename "df")
		Amount        nullable BigUint (rename "a")
		Name          nullable String  (rename "n")
		Message       nullable String  (rename "m")
	}
`

	builder, err := NewLDBuilder("TxMinter", []byte(sch), (*bindTxMinter)(nil))
	if err != nil {
		panic(err)
	}
	txMinterLDBuilder = builder
}
