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
	Purchaser *util.EthID      `cbor:"to,omitempty" json:"purchaser,omitempty"` // optional designated purchaser

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxExchanger is well-formed.
func (t *TxExchanger) SyntacticVerify() error {
	errPrefix := "TxExchanger.SyntacticVerify failed:"

	switch {
	case t == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case t.Nonce == 0:
		return fmt.Errorf("%s invalid nonce", errPrefix)

	case !t.Sell.Valid():
		return fmt.Errorf("%s invalid sell token symbol %s", errPrefix, strconv.Quote(t.Sell.GoString()))

	case !t.Receive.Valid():
		return fmt.Errorf("%s invalid receive token symbol %s", errPrefix, strconv.Quote(t.Receive.GoString()))

	case t.Sell == t.Receive:
		return fmt.Errorf("%s sell and receive token should not equal", errPrefix)

	case t.Minimum == nil || t.Minimum.Sign() < 1:
		return fmt.Errorf("%s invalid minimum", errPrefix)

	case t.Quota == nil || t.Quota.Cmp(t.Minimum) < 0:
		return fmt.Errorf("%s invalid quota", errPrefix)

	case t.Price == nil || t.Price.Sign() < 1:
		return fmt.Errorf("%s invalid price", errPrefix)

	case t.Payee == util.EthIDEmpty:
		return fmt.Errorf("%s invalid payee", errPrefix)
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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
	return UnmarshalCBOR(data, t)
}

func (t *TxExchanger) Marshal() ([]byte, error) {
	return MarshalCBOR(t)
}
