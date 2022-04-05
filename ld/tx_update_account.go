// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
)

type TxUpdateAccountGuardians struct {
	Threshold uint8         `json:"threshold"`
	Guardians []ids.ShortID `json:"guardians"`
	raw       []byte
}

func (tx *TxUpdateAccountGuardians) Copy() *TxUpdateAccountGuardians {
	x := new(TxUpdateAccountGuardians)
	*x = *tx
	x.Guardians = make([]ids.ShortID, len(tx.Guardians))
	copy(x.Guardians, tx.Guardians)
	x.raw = make([]byte, len(tx.raw))
	copy(x.raw, tx.raw)
	return x
}

// SyntacticVerify verifies that a *Account is well-formed.
func (tx *TxUpdateAccountGuardians) SyntacticVerify() error {
	if len(tx.Guardians) > 15 {
		return fmt.Errorf("too many account Guardians")
	}
	if tx.Threshold < 1 || int(tx.Threshold) > len(tx.Guardians)+1 {
		return fmt.Errorf("invalid account Threshold")
	}

	set := ids.NewShortSet(len(tx.Guardians))
	for _, id := range tx.Guardians {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid account Guardian")
		}
		if set.Contains(id) {
			return fmt.Errorf("address %s exists", id)
		}
		set.Add(id)
	}
	if _, err := tx.Marshal(); err != nil {
		return fmt.Errorf("account marshal error: %v", err)
	}
	return nil
}

func (tx *TxUpdateAccountGuardians) Equal(o *TxUpdateAccountGuardians) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(tx.raw) > 0 {
		return bytes.Equal(o.raw, tx.raw)
	}
	if o.Threshold != tx.Threshold {
		return false
	}
	if len(o.Guardians) != len(tx.Guardians) {
		return false
	}
	for i := range tx.Guardians {
		if o.Guardians[i] != tx.Guardians[i] {
			return false
		}
	}
	return true
}

func (tx *TxUpdateAccountGuardians) Bytes() []byte {
	if len(tx.raw) == 0 {
		if _, err := tx.Marshal(); err != nil {
			panic(err)
		}
	}

	return tx.raw
}

func (tx *TxUpdateAccountGuardians) Unmarshal(data []byte) error {
	p, err := txUpdateAccountGuardiansLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxUpdateAccountGuardians); ok {
		tx.Threshold = v.Threshold.Value()
		if tx.Guardians, err = ToShortIDs(v.Guardians); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		tx.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *txUpdateAccountGuardiansLDBuilder")
}

func (tx *TxUpdateAccountGuardians) Marshal() ([]byte, error) {
	v := &bindTxUpdateAccountGuardians{
		Threshold: FromUint8(tx.Threshold),
		Guardians: FromShortIDs(tx.Guardians),
	}
	data, err := txUpdateAccountGuardiansLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	tx.raw = data
	return data, nil
}

type bindTxUpdateAccountGuardians struct {
	Threshold Uint8
	Guardians [][]byte
}

var txUpdateAccountGuardiansLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type ID20 bytes
	type TxUpdateAccountGuardians struct {
		Threshold Uint8  (rename "th")
		Guardians [ID20] (rename "gs")
	}
`
	builder, err := NewLDBuilder("TxUpdateAccountGuardians", []byte(sch), (*bindTxUpdateAccountGuardians)(nil))
	if err != nil {
		panic(err)
	}
	txUpdateAccountGuardiansLDBuilder = builder
}
