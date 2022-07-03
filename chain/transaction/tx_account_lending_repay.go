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

func (tx *TxRepay) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxRepay.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = bs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	actual, err := tx.to.Repay(tx.token, tx.ld.From, tx.ld.Amount)
	if err != nil {
		return errp.ErrorIf(err)
	}
	tx.amount.Set(actual)
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
