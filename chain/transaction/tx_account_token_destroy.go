// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"github.com/ldclabs/ldvm/util"
)

type TxDestroyTokenAccount struct {
	TxBase
}

func (tx *TxDestroyTokenAccount) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxDestroyTokenAccount.SyntacticVerify error: ")

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

	if token := util.TokenSymbol(tx.ld.From); !token.Valid() {
		return errp.Errorf("invalid token %s", token.GoString())
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxDestroyTokenAccount.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return errp.Errorf("invalid signature for keepers")
	}
	if err = tx.from.CheckDestroyToken(tx.to); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Accept(bctx BlockContext, bs BlockState) error {
	errp := util.ErrPrefix("TxDestroyTokenAccount.Accept error: ")

	if err := tx.TxBase.Accept(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	// DestroyToken after TxBase.Accept
	tx.from.pledge.SetUint64(0)
	return errp.ErrorIf(tx.from.DestroyToken(tx.to))
}
