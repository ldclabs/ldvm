// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
)

type TxRepay struct {
	TxBase
}

func (tx *TxRepay) SyntacticVerify() error {
	var err error
	errPrefix := "TxRepay.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as lender", errPrefix)

	case tx.ld.Amount == nil || tx.ld.Amount.Sign() <= 0:
		return fmt.Errorf("%s invalid amount, expected > 0, got %v", errPrefix, tx.ld.Amount)
	}
	return nil
}

func (tx *TxRepay) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxRepay.Verify failed: %v", err)
	}
	if err = tx.to.CheckRepay(tx.token, tx.ld.From, tx.ld.Amount); err != nil {
		return fmt.Errorf("TxRepay.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxRepay) Accept(bctx BlockContext, bs BlockState) error {
	actual, err := tx.to.Repay(tx.token, tx.ld.From, tx.ld.Amount)
	if err != nil {
		return err
	}
	tx.amount.Set(actual)
	return tx.TxBase.Accept(bctx, bs)
}
