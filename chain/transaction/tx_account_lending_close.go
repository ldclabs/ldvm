// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"github.com/ldclabs/ldvm/util"
)

type TxCloseLending struct {
	TxBase
}

func (tx *TxCloseLending) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxCloseLending.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")
	}
	return nil
}

func (tx *TxCloseLending) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCloseLending.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.from.CloseLending(); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
