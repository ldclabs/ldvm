// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferMultiple struct {
	TxBase
	input ld.SendOutputs
}

func (tx *TxTransferMultiple) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("transactions.TxTransferMultiple.MarshalJSON: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxTransferMultiple) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("transactions.TxTransferMultiple.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

// Apply skipping signature verification
func (tx *TxTransferMultiple) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("transactions.TxTransferMultiple.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	totalAmount := new(big.Int)
	recipients := make(map[util.Address]*Account, len(tx.input))
	for _, output := range tx.input {
		totalAmount = totalAmount.Add(totalAmount, output.Amount)
		if recipients[output.To], err = cs.LoadAccount(output.To); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if err = tx.TxBase.accept(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.from.Sub(tx.token, totalAmount); err != nil {
		return err
	}

	for _, output := range tx.input {
		if err = recipients[output.To].Add(tx.token, output.Amount); err != nil {
			return err
		}
	}

	return nil
}