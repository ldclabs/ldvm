// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

// TxTransfer is a hybrid data model for:
//
// TxTransferPay{To[, Token, Amount, Expire, Data]}
// TxTransferCash{Nonce, From, Amount, Expire[, Token, To, Data]}
// TxTakeStake{Nonce, From, To, Amount, Expire[, Data]}
type TxTransfer struct {
	Nonce  uint64      // sender's nonce
	Token  ids.ShortID // token symbol, default is NativeToken
	From   ids.ShortID // amount sender
	To     ids.ShortID // amount recipient
	Amount *big.Int    // transfer amount
	Expire uint64
	Data   []byte

	// external assignment
	raw []byte
}

type jsonTxTransfer struct {
	Nonce  uint64          `json:"nonce,omitempty"`
	Token  string          `json:"token,omitempty"`
	From   string          `json:"from,omitempty"`
	To     string          `json:"to,omitempty"`
	Amount *big.Int        `json:"amount,omitempty"`
	Expire uint64          `json:"expire,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

func (t *TxTransfer) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxTransfer{
		Nonce:  t.Nonce,
		Amount: t.Amount,
		Expire: t.Expire,
		Token:  util.TokenSymbol(t.Token).String(),
		Data:   util.JSONMarshalData(t.Data),
	}

	if t.From != ids.ShortEmpty {
		v.From = util.EthID(t.From).String()
	}
	if t.To != ids.ShortEmpty {
		v.To = util.EthID(t.To).String()
	}
	return json.Marshal(v)
}

func (t *TxTransfer) Copy() *TxTransfer {
	x := new(TxTransfer)
	*x = *t
	if t.Amount != nil {
		x.Amount = new(big.Int).Set(t.Amount)
	}
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxTransfer is well-formed.
func (t *TxTransfer) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxTransfer")
	}

	if t.Nonce == 0 {
		return fmt.Errorf("invalid nonce")
	}
	if t.Token != constants.LDCAccount && util.TokenSymbol(t.Token).String() == "" {
		return fmt.Errorf("invalid token symbol")
	}
	if t.From == ids.ShortEmpty {
		return fmt.Errorf("invalid from")
	}
	if t.Amount == nil || t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxTransfer marshal error: %v", err)
	}
	return nil
}

func (t *TxTransfer) Equal(o *TxTransfer) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Nonce != t.Nonce {
		return false
	}
	if o.Token != t.Token {
		return false
	}
	if o.From != t.From {
		return false
	}
	if o.To != t.To {
		return false
	}
	if o.Amount.Cmp(t.Amount) != 0 {
		return false
	}
	if o.Expire != t.Expire {
		return false
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *TxTransfer) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxTransfer) Unmarshal(data []byte) error {
	p, err := txTransferLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxTransfer); ok {
		if !v.Nonce.Valid() ||
			!v.Expire.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}

		t.Nonce = v.Nonce.Value()
		t.Expire = v.Expire.Value()
		t.Amount = v.Amount.Value()
		t.Data = PtrToBytes(v.Data)
		if t.Token, err = PtrToShortID(v.Token); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.From, err = PtrToShortID(v.From); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxTransfer")
}

func (t *TxTransfer) Marshal() ([]byte, error) {
	v := &bindTxTransfer{
		Nonce:  PtrFromUint64(t.Nonce),
		Token:  PtrFromShortID(ids.ShortID(t.Token)),
		To:     PtrFromShortID(t.To),
		From:   PtrFromShortID(t.From),
		Amount: PtrFromUint(t.Amount),
		Expire: PtrFromUint64(t.Expire),
		Data:   PtrFromBytes(t.Data),
	}
	data, err := txTransferLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindTxTransfer struct {
	Nonce  *Uint64
	Token  *[]byte
	From   *[]byte
	To     *[]byte
	Amount *BigUint
	Expire *Uint64
	Data   *[]byte
}

var txTransferLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type TxTransfer struct {
		Nonce  nullable Uint64  (rename "n")
		Token  nullable ID20    (rename "tk")
		From   nullable ID20    (rename "fr")
		To     nullable ID20    (rename "to")
		Amount nullable BigUint (rename "a")
		Expire nullable Uint64  (rename "e")
		Data   nullable Bytes   (rename "d")
	}
`

	builder, err := NewLDBuilder("TxTransfer", []byte(sch), (*bindTxTransfer)(nil))
	if err != nil {
		panic(err)
	}
	txTransferLDBuilder = builder
}
