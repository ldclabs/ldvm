// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
)

type TxCloseLending struct {
	TxBase
}

func (tx *TxCloseLending) SyntacticVerify() error {
	var err error
	errPrefix := "TxCloseLending.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)
	}
	return nil
}

func (tx *TxCloseLending) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxCloseLending.Verify failed: %v", err)
	}
	if err = tx.from.CheckCloseLending(); err != nil {
		return fmt.Errorf("TxCloseLending.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxCloseLending) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.CloseLending(); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
