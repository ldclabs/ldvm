// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"github.com/ldclabs/ldvm/util"
)

type TxRepay struct {
	TxBase
}

func (tx *TxRepay) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxRepay.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as lender")

	case tx.ld.Amount == nil || tx.ld.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected > 0, got %v", tx.ld.Amount)
	}
	return nil
}

func (tx *TxRepay) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxRepay.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.CheckRepay(tx.token, tx.ld.From, tx.ld.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxRepay) Accept(bctx BlockContext, bs BlockState) error {
	errp := util.ErrPrefix("TxRepay.Accept error: ")

	actual, err := tx.to.Repay(tx.token, tx.ld.From, tx.ld.Amount)
	if err != nil {
		return errp.ErrorIf(err)
	}
	tx.amount.Set(actual)
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
