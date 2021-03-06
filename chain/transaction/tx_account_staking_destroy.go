// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"github.com/ldclabs/ldvm/util"
)

type TxDestroyStake struct {
	TxBase
}

func (tx *TxDestroyStake) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxDestroyStake.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as pledge recipient")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")
	}

	if stake := util.StakeSymbol(tx.ld.From); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}
	return nil
}

func (tx *TxDestroyStake) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxDestroyStake.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return errp.Errorf("invalid signatures for stake keepers")
	}

	if err = bs.LoadLedger(tx.from); err != nil {
		return errp.ErrorIf(err)
	}

	if err := tx.TxBase.accept(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	// do it after TxBase.Accept
	tx.from.pledge.SetUint64(0)
	return errp.ErrorIf(tx.from.DestroyStake(tx.to))
}
