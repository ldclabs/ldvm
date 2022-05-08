// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/util"
)

// TxTransfer is a hybrid data model for:
//
// TxTransferPay{To[, Token, Amount, Expire, Data]}
// TxTransferCash{Nonce, From, Amount, Expire[, Token, To, Data]}
// TxTakeStake{Nonce, From, To, Amount, Expire[, Data]}
type TxTransfer struct {
	Nonce  uint64            `cbor:"n,omitempty" json:"nonce,omitempty"`  // sender's nonce
	Token  *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"` // token symbol, default is NativeToken
	From   *util.EthID       `cbor:"fr,omitempty" json:"from,omitempty"`  // amount sender
	To     *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`    // amount recipient
	Amount *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"` // transfer amount
	Expire uint64            `cbor:"e,omitempty" json:"expire,omitempty"`
	Data   RawData           `cbor:"d,omitempty" json:"data,omitempty"`
}

// SyntacticVerify verifies that a *TxTransfer is well-formed.
func (t *TxTransfer) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxTransfer")
	}
	if t.Token != nil && !t.Token.Valid() {
		return fmt.Errorf("invalid token symbol")
	}
	if t.Amount != nil && t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxTransfer marshal error: %v", err)
	}
	return nil
}

func (t *TxTransfer) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxTransfer) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
