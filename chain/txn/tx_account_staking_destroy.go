// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxDestroyStake struct {
	TxBase
}

func (tx *TxDestroyStake) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxDestroyStake.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as pledge recipient")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")
	}

	if stake := ids.StakeSymbol(tx.ld.Tx.From); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}
	return nil
}

func (tx *TxDestroyStake) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxDestroyStake.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.VerifyPlus(tx.ld.TxHash(), tx.ld.Signatures, nil) {
		return errp.Errorf("invalid signatures for stake keepers")
	}

	if err = cs.LoadLedger(tx.from); err != nil {
		return errp.ErrorIf(err)
	}

	if err := tx.TxBase.accept(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	// do it after TxBase.Accept
	tx.from.ResetPledge()
	return errp.ErrorIf(tx.from.DestroyStake(tx.to))
}
