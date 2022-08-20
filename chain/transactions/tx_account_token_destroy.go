// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"github.com/ldclabs/ldvm/util"
)

type TxDestroyToken struct {
	TxBase
}

func (tx *TxDestroyToken) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxDestroyToken.SyntacticVerify error: ")

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

func (tx *TxDestroyToken) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxDestroyToken.Apply error: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return errp.Errorf("invalid signature for keepers")
	}

	if err = cs.LoadLedger(tx.from); err != nil {
		return errp.ErrorIf(err)
	}

	if err := tx.TxBase.accept(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	// DestroyToken after TxBase.accept
	tx.from.pledge.SetUint64(0)
	return errp.ErrorIf(tx.from.DestroyToken(tx.to))
}
