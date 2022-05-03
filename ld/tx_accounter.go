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

// TxAccounter
type TxAccounter struct {
	Threshold uint8
	Keepers   []ids.ShortID
	Amount    *big.Int
	Name      string
	Message   string
	Data      []byte

	// external assignment
	DataObject LDObject
	raw        []byte
}

type jsonTxAccounter struct {
	Threshold uint8           `json:"threshold,omitempty"`
	Keepers   []string        `json:"keepers,omitempty"`
	Amount    *big.Int        `json:"amount,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Name      string          `json:"name,omitempty"`
	Message   string          `json:"message,omitempty"`
}

func (t *TxAccounter) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxAccounter{
		Threshold: t.Threshold,
		Amount:    t.Amount,
		Name:      t.Name,
		Message:   t.Message,
		Data:      util.JSONMarshalData(t.Data),
	}
	if len(t.Keepers) > 0 {
		v.Keepers = make([]string, len(t.Keepers))
		for i := range t.Keepers {
			v.Keepers[i] = util.EthID(t.Keepers[i]).String()
		}
	}
	if t.DataObject != nil {
		data, err := t.DataObject.MarshalJSON()
		if err != nil {
			return nil, err
		}
		v.Data = data
	}
	return json.Marshal(v)
}

func (t *TxAccounter) Copy() *TxAccounter {
	x := new(TxAccounter)
	*x = *t
	if t.Amount != nil {
		x.Amount = new(big.Int).Set(t.Amount)
	}
	x.Keepers = make([]ids.ShortID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxAccounter is well-formed.
func (t *TxAccounter) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxAccounter")
	}
	if t.Name != "" && !util.ValidName(t.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(t.Name))
	}
	if t.Message != "" && !util.ValidMessage(t.Message) {
		return fmt.Errorf("invalid message string %s", strconv.Quote(t.Message))
	}

	if t.Amount == nil || t.Amount.Sign() < 0 {
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
	if t.DataObject != nil {
		if err := t.DataObject.SyntacticVerify(); err != nil {
			return err
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxAccounter marshal error: %v", err)
	}
	return nil
}

func (t *TxAccounter) Equal(o *TxAccounter) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Threshold != t.Threshold {
		return false
	}
	if o.Name != t.Name {
		return false
	}
	if o.Message != t.Message {
		return false
	}
	if o.Amount.Cmp(t.Amount) != 0 {
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
	return bytes.Equal(o.Data, t.Data)
}

func (t *TxAccounter) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxAccounter) Unmarshal(data []byte) error {
	p, err := txMinterLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxAccounter); ok {
		if !v.Threshold.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}

		t.Threshold = v.Threshold.Value()
		t.Amount = v.Amount.Value()
		t.Data = PtrToBytes(v.Data)
		if t.Keepers, err = PtrToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if v.Name != nil {
			t.Name = *v.Name
		}
		if v.Message != nil {
			t.Message = *v.Message
		}
		if len(t.Data) > 0 && t.DataObject != nil {
			if err = t.DataObject.Unmarshal(t.Data); err != nil {
				return err
			}
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxAccounter")
}

func (t *TxAccounter) Marshal() ([]byte, error) {
	if len(t.Data) == 0 && t.DataObject != nil {
		data, err := t.DataObject.Marshal()
		if err != nil {
			return nil, err
		}
		t.Data = data
	}

	v := &bindTxAccounter{
		Threshold: PtrFromUint8(t.Threshold),
		Keepers:   PtrFromShortIDs(t.Keepers),
		Amount:    PtrFromUint(t.Amount),
		Data:      PtrFromBytes(t.Data),
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

type bindTxAccounter struct {
	Threshold *Uint8
	Keepers   *[][]byte
	Amount    *BigUint
	Data      *[]byte
	Name      *string
	Message   *string
}

var txMinterLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type TxAccounter struct {
		Threshold nullable Uint8   (rename "th")
		Keepers   nullable [ID20]  (rename "kp")
		Amount    nullable BigUint (rename "a")
		Data      nullable Bytes   (rename "d")
		Name      nullable String  (rename "n")
		Message   nullable String  (rename "m")
	}
`

	builder, err := NewLDBuilder("TxAccounter", []byte(sch), (*bindTxAccounter)(nil))
	if err != nil {
		panic(err)
	}
	txMinterLDBuilder = builder
}
