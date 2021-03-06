// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxExchange struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxExchanger
	quantity  *big.Int
}

func (tx *TxExchange) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxExchange.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxExchange) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxExchange.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("invalid to")

	case tx.ld.Amount == nil:
		return errp.Errorf("invalid amount")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxExchanger{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	// quantity = amount * 1_000_000_000 / price
	tx.quantity = new(big.Int).SetUint64(constants.LDC)
	tx.quantity.Mul(tx.quantity, tx.ld.Amount)
	tx.quantity.Quo(tx.quantity, tx.input.Price)

	switch {
	case tx.quantity.Cmp(tx.input.Minimum) < 0:
		min := new(big.Int).Mul(tx.input.Minimum, tx.input.Price)
		min.Quo(min, new(big.Int).SetUint64(constants.LDC))
		return errp.Errorf("invalid amount, expected >=%v, got %v",
			min, tx.ld.Amount)

	case tx.quantity.Cmp(tx.input.Quota) > 0:
		max := new(big.Int).Mul(tx.input.Quota, tx.input.Price)
		max.Quo(max, new(big.Int).SetUint64(constants.LDC))
		return errp.Errorf("invalid amount, expected <=%v, got %v",
			max, tx.ld.Amount)

	case tx.input.Purchaser != nil && *tx.input.Purchaser != tx.ld.From:
		return errp.Errorf("invalid from, expected %s, got %s",
			*tx.input.Purchaser, tx.ld.From)

	case tx.input.Payee != *tx.ld.To:
		return errp.Errorf("invalid to, expected %s, got %s",
			tx.input.Payee, tx.ld.To)

	case tx.input.Receive != tx.token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Receive.GoString(), tx.token.GoString())

	case tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired")
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return errp.Errorf("invalid exSignatures, %v", err)
	}
	return nil
}

func (tx *TxExchange) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxExchange.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	// verify seller's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return errp.Errorf("invalid signatures for seller")
	}

	if err = tx.to.SubByNonceTable(
		tx.input.Sell, tx.input.Expire, tx.input.Nonce, tx.quantity); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.Add(tx.input.Sell, tx.quantity); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
