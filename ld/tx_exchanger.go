// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

// TxExchanger
type TxExchanger struct {
	Nonce     uint64          `cbor:"n" json:"nonce"`    // saler' account nonce
	Sell      ids.TokenSymbol `cbor:"st" json:"sell"`    // token symbol to sell
	Receive   ids.TokenSymbol `cbor:"rt" json:"receive"` // token symbol to receive
	Quota     *big.Int        `cbor:"q" json:"quota"`    // token sales quota per a tx
	Minimum   *big.Int        `cbor:"m" json:"minimum"`  // minimum amount to buy
	Price     *big.Int        `cbor:"p" json:"price"`    // receive token amount = Quota * Price
	Expire    uint64          `cbor:"e" json:"expire"`
	Payee     ids.Address     `cbor:"py" json:"payee"`
	Purchaser *ids.Address    `cbor:"to,omitempty" json:"purchaser,omitempty"` // optional designated purchaser

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxExchanger is well-formed.
func (t *TxExchanger) SyntacticVerify() error {
	errp := erring.ErrPrefix("ld.TxExchanger.SyntacticVerify: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.Nonce == 0:
		return errp.Errorf("invalid nonce")

	case !t.Sell.Valid():
		return errp.Errorf("invalid sell token symbol %q", t.Sell.GoString())

	case !t.Receive.Valid():
		return errp.Errorf("invalid receive token symbol %q", t.Receive.GoString())

	case t.Sell == t.Receive:
		return errp.Errorf("sell and receive token should not equal")

	case t.Minimum == nil || t.Minimum.Sign() < 1:
		return errp.Errorf("invalid minimum")

	case t.Quota == nil || t.Quota.Cmp(t.Minimum) < 0:
		return errp.Errorf("invalid quota")

	case t.Price == nil || t.Price.Sign() < 1:
		return errp.Errorf("invalid price")

	case t.Payee == ids.EmptyAddress:
		return errp.Errorf("invalid payee")
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
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
	return erring.ErrPrefix("ld.TxExchanger.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, t))
}

func (t *TxExchanger) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.TxExchanger.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(t))
}
