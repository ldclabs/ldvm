// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
)

const (
	// 0: mint tx fee, issued by validators
	TypeMintFee TxType = iota

	// 1: send given amount of Wei to address, throws on failure
	TypeTransfer

	// 2. update account's Guardians and Threshold
	TypeUpdateAccountGuardians

	// 3. create a data model
	TypeCreateModel

	// 4. update data model's Keepers and Threshold
	TypeUpdateModelKeepers

	// 5. create a data from the model
	TypeCreateData

	// 6. update the data's Data
	TypeUpdateData

	// 7. update data's Owners and Threshold
	TypeUpdateDataOwners

	// 8. update data's Owners and Threshold by authorization
	TypeUpdateDataOwnersByAuth

	// 9. delete the data
	TypeDeleteData
)

// TxType is an uint8 representing the type of the tx

type TxType uint8

type Transaction struct {
	Type         TxType      `json:"type"`
	ChainID      uint64      `json:"chainID"`
	AccountNonce uint64      `json:"accountNonce"`
	Gas          uint64      `json:"gas"`
	GasTipCap    uint64      `json:"gasTipCap"`
	GasFeeCap    uint64      `json:"gasFeeCap"`
	Amount       *big.Int    `json:"amount"`
	From         ids.ShortID `json:"from"`
	To           ids.ShortID `json:"to"`
	Data         []byte      `json:"data"`
	Signatures   []Signature `json:"signatures"`
	ExSignatures []Signature `json:"exSignatures"`
	gas          uint64
	id           ids.ID
	raw          []byte // the transaction's raw bytes
}

func (t *Transaction) ID() ids.ID {
	if t.id == ids.Empty {
		t.gas = t.Gas
		t.Gas = 0 // compute id without gas
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
		t.id = hashing.ComputeHash256Array(t.Bytes())
		// update raw again
		t.Gas = t.gas
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}
	return t.id
}

func (t *Transaction) Copy() *Transaction {
	x := new(Transaction)
	*x = *t
	x.Amount = new(big.Int).Set(t.Amount)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = make([]byte, len(t.raw))
	copy(x.raw, t.raw)
	x.Signatures = make([]Signature, len(t.Signatures))
	copy(x.Signatures, t.Signatures)
	x.ExSignatures = make([]Signature, len(t.ExSignatures))
	copy(x.ExSignatures, t.ExSignatures)
	return x
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	if t.Type > TypeDeleteData {
		return fmt.Errorf("invalid transaction Type")
	}
	if t.Amount != nil && t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid transaction Amount")
	}
	if len(t.Signatures) > 16 {
		return fmt.Errorf("too many transaction signatures")
	}
	if len(t.ExSignatures) > 16 {
		return fmt.Errorf("too many transaction signatures")
	}

	if t.id == ids.Empty {
		t.gas = t.Gas
		t.Gas = 0 // compute id without gas
		if _, err := t.Marshal(); err != nil {
			t.Gas = t.gas
			return err
		}
		t.id = hashing.ComputeHash256Array(t.Bytes())
		t.Gas = t.gas
	}

	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("transaction marshal error: %v", err)
	}
	return nil
}

func (t *Transaction) Equal(o *Transaction) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Type != t.Type {
		return false
	}
	if o.ChainID != t.ChainID {
		return false
	}
	if o.AccountNonce != t.AccountNonce {
		return false
	}
	if o.Gas != t.Gas {
		return false
	}
	if o.GasTipCap != t.GasTipCap {
		return false
	}
	if o.GasFeeCap != t.GasFeeCap {
		return false
	}
	if o.Amount == nil || t.Amount == nil {
		if o.Amount != t.Amount {
			return false
		}
	}
	if o.Amount.Cmp(t.Amount) != 0 {
		return false
	}
	if o.From != t.From {
		return false
	}
	if o.To != t.To {
		return false
	}
	if len(o.Signatures) != len(t.Signatures) {
		return false
	}
	for i := range t.Signatures {
		if o.Signatures[i] != t.Signatures[i] {
			return false
		}
	}
	if len(o.ExSignatures) != len(t.ExSignatures) {
		return false
	}
	for i := range t.ExSignatures {
		if o.ExSignatures[i] != t.ExSignatures[i] {
			return false
		}
	}
	if o.id != t.id {
		return false
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *Transaction) Bytes() []byte {
	if len(t.raw) == 0 || t.gas != t.Gas {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *Transaction) Unmarshal(data []byte) error {
	p, err := transactionLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTransaction); ok {
		t.Type = TxType(v.Type.Value())
		t.ChainID = v.ChainID.Value()
		t.AccountNonce = v.AccountNonce.Value()
		t.GasTipCap = v.GasTipCap.Value()
		t.GasFeeCap = v.GasFeeCap.Value()
		t.Gas = v.Gas.Value()
		t.Amount = PtrToBigInt(v.Amount)
		t.Data = PtrToBytes(v.Data)
		if t.From, err = PtrToShortID(v.From); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.Signatures, err = PtrToSignatures(v.Signatures); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.ExSignatures, err = PtrToSignatures(v.ExSignatures); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.id, err = PtrToID(v.ID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.gas = t.Gas
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTransaction")
}

func (t *Transaction) Marshal() ([]byte, error) {
	v := &bindTransaction{
		Type:         FromUint8(uint8(t.Type)),
		ChainID:      FromUint64(t.ChainID),
		AccountNonce: PtrFromUint64(t.AccountNonce),
		Gas:          PtrFromUint64(t.Gas),
		GasTipCap:    PtrFromUint64(t.GasTipCap),
		GasFeeCap:    PtrFromUint64(t.GasFeeCap),
		Amount:       PtrFromBigInt(t.Amount),
		From:         PtrFromShortID(t.From),
		To:           PtrFromShortID(t.To),
		Data:         PtrFromBytes(t.Data),
		Signatures:   PtrFromSignatures(t.Signatures),
		ExSignatures: PtrFromSignatures(t.ExSignatures),
		ID:           PtrFromID(t.id),
	}
	data, err := transactionLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.gas = t.Gas
	t.raw = data
	return data, nil
}

type bindTransaction struct {
	Type         Uint8
	ChainID      Uint64
	AccountNonce *Uint64
	Gas          *Uint64
	GasTipCap    *Uint64
	GasFeeCap    *Uint64
	Amount       *[]byte
	From         *[]byte
	To           *[]byte
	Data         *[]byte
	Signatures   *[][]byte
	ExSignatures *[][]byte
	ID           *[]byte
}

var transactionLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type ID32 bytes
	type Sig65 bytes
	type BigInt bytes
	type Transaction struct {
		Type         Uint8            (rename "t")
		ChainID      Uint64           (rename "c")
		AccountNonce nullable Uint64  (rename "n")
		Gas          nullable Uint64  (rename "g")
		GasTipCap    nullable Uint64  (rename "gt")
		GasFeeCap    nullable Uint64  (rename "gf")
		Amount       nullable BigInt  (rename "a")
		From         nullable ID20    (rename "fr")
		To           nullable ID20    (rename "to")
		Data         nullable Bytes   (rename "d")
		Signatures   nullable [Sig65] (rename "ss")
		ExSignatures nullable [Sig65] (rename "es")
		ID           nullable ID32    (rename "id")
	}
`
	builder, err := NewLDBuilder("Transaction", []byte(sch), (*bindTransaction)(nil))
	if err != nil {
		panic(err)
	}
	transactionLDBuilder = builder
}
