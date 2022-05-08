// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

// TxExchanger
type TxExchanger struct {
	Nonce   uint64           `cbor:"n" json:"nonce"`    // saler' account nonce
	Sell    util.TokenSymbol `cbor:"st" json:"sell"`    // token symbol to sell
	Receive util.TokenSymbol `cbor:"rt" json:"receive"` // token symbol to receive
	Quota   *big.Int         `cbor:"q" json:"quota"`    // token sales quota per a tx
	Minimum *big.Int         `cbor:"m" json:"minimum"`  // minimum amount to buy
	Price   *big.Int         `cbor:"p" json:"price"`    // receive token amount = Quota * Price
	Expire  uint64           `cbor:"e" json:"expire"`
	Seller  util.EthID       `cbor:"sl" json:"payee"`
	To      *util.EthID      `cbor:"to,omitempty" json:"to,omitempty"` // optional designated purchaser
}

// SyntacticVerify verifies that a *TxExchanger is well-formed.
func (t *TxExchanger) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxExchanger")
	}

	if t.Nonce == 0 {
		return fmt.Errorf("invalid nonce")
	}
	if t.Sell != constants.NativeToken && t.Sell.String() == "" {
		return fmt.Errorf("invalid token symbol to sell")
	}
	if t.Receive != constants.NativeToken && t.Receive.String() == "" {
		return fmt.Errorf("invalid token symbol to receive")
	}
	if t.Sell == t.Receive {
		return fmt.Errorf("invalid token symbol to receive")
	}
	if t.Quota == nil || t.Quota.Sign() < 1 {
		return fmt.Errorf("invalid quota")
	}
	if t.Minimum == nil || t.Minimum.Sign() < 1 {
		return fmt.Errorf("invalid minimum")
	}
	if t.Price == nil || t.Price.Sign() < 1 {
		return fmt.Errorf("invalid price")
	}
	if t.Seller == util.EthIDEmpty {
		return fmt.Errorf("invalid payee")
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxExchanger marshal error: %v", err)
	}
	return nil
}

func (t *TxExchanger) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxExchanger) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
