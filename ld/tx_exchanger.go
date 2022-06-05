// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

// TxExchanger
type TxExchanger struct {
	Nonce     uint64           `cbor:"n" json:"nonce"`    // saler' account nonce
	Sell      util.TokenSymbol `cbor:"st" json:"sell"`    // token symbol to sell
	Receive   util.TokenSymbol `cbor:"rt" json:"receive"` // token symbol to receive
	Quota     *big.Int         `cbor:"q" json:"quota"`    // token sales quota per a tx
	Minimum   *big.Int         `cbor:"m" json:"minimum"`  // minimum amount to buy
	Price     *big.Int         `cbor:"p" json:"price"`    // receive token amount = Quota * Price
	Expire    uint64           `cbor:"e" json:"expire"`
	Payee     util.EthID       `cbor:"py" json:"payee"`
	Purchaser *util.EthID      `cbor:"to,omitempty" json:"to,omitempty"` // optional designated purchaser

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxExchanger is well-formed.
func (t *TxExchanger) SyntacticVerify() error {
	switch {
	case t == nil:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: nil pointer")
	case t.Nonce == 0:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid nonce")
	case !t.Sell.Valid():
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid sell token symbol %s",
			strconv.Quote(t.Sell.GoString()))
	case !t.Receive.Valid():
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid receive token symbol %s",
			strconv.Quote(t.Receive.GoString()))
	case t.Sell == t.Receive:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: sell and receive token should not equal")
	case t.Minimum == nil || t.Minimum.Sign() < 1:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid minimum")
	case t.Quota == nil || t.Quota.Cmp(t.Minimum) < 0:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid quota")
	case t.Price == nil || t.Price.Sign() < 1:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid price")
	case t.Payee == util.EthIDEmpty:
		return fmt.Errorf("TxExchanger.SyntacticVerify failed: invalid payee")
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("TxExchanger.SyntacticVerify marshal failed: %v", err)
	}
	return nil
}

func (t *TxExchanger) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxExchanger) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxExchanger) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
