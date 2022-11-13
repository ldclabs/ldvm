// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import "github.com/ldclabs/ldvm/util/erring"

type TxRepay struct {
	TxBase
}

func (tx *TxRepay) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxRepay.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as lender")

	case tx.ld.Tx.Amount == nil || tx.ld.Tx.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected > 0, got %v", tx.ld.Tx.Amount)
	}
	return nil
}

func (tx *TxRepay) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxRepay.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	actual, err := tx.to.Repay(tx.token, tx.ld.Tx.From, tx.ld.Tx.Amount)
	if err != nil {
		return errp.ErrorIf(err)
	}
	tx.amount.Set(actual)
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
