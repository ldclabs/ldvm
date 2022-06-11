// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
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
	Data   RawData           `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxTransfer is well-formed.
func (t *TxTransfer) SyntacticVerify() error {
	errPrefix := "TxTransfer.SyntacticVerify failed:"
	switch {
	case t == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case t.Token != nil && !t.Token.Valid():
		return fmt.Errorf("%s invalid token symbol %s", errPrefix, strconv.Quote(t.Token.GoString()))

	case t.Amount != nil && t.Amount.Sign() < 1:
		return fmt.Errorf("%s invalid amount", errPrefix)
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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
	return DecMode.Unmarshal(data, t)
}

func (t *TxTransfer) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
