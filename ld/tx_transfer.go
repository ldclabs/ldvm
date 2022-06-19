// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

// TxTransfer is a hybrid data model for:
//
// TxTransferPay{To[, Token, Amount, Expire, Data]}
// TxTransferCash{Nonce, From, Amount, Expire[, Token, To, Data]}
// TxTakeStake{Nonce, From, To, Amount, Expire[, Data]}
type TxTransfer struct {
	Nonce  uint64            `cbor:"n,omitempty" json:"nonce,omitempty"`  // sender's nonce
	From   *util.EthID       `cbor:"fr,omitempty" json:"from,omitempty"`  // amount sender
	To     *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`    // amount recipient
	Token  *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"` // token symbol, default is NativeToken
	Amount *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"` // transfer amount
	Expire uint64            `cbor:"e,omitempty" json:"expire,omitempty"`
	Data   util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxTransfer is well-formed.
func (t *TxTransfer) SyntacticVerify() error {
	errp := util.ErrPrefix("TxTransfer.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.Token != nil && !t.Token.Valid():
		return errp.Errorf("invalid token symbol %s", strconv.Quote(t.Token.GoString()))

	case t.Amount != nil && t.Amount.Sign() < 0:
		return errp.Errorf("invalid amount")
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxTransfer) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxTransfer) Unmarshal(data []byte) error {
	if err := util.UnmarshalCBOR(data, t); err != nil {
		return util.ErrPrefix("TxTransfer.Unmarshal error: ").ErrorIf(err)
	}
	return nil
}

func (t *TxTransfer) Marshal() ([]byte, error) {
	data, err := util.MarshalCBOR(t)
	if err != nil {
		return nil, util.ErrPrefix("TxTransfer.Marshal error: ").ErrorIf(err)
	}
	return data, nil
}
