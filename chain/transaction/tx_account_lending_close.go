// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"strconv"
)

type TxCloseLending struct {
	TxBase
}

func (tx *TxCloseLending) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.To != nil {
		return fmt.Errorf("TxCloseLending invalid to")
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxCloseLending invalid amount, expected 0, got %v", tx.ld.Amount)
	}
	return nil
}

func (tx *TxCloseLending) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	return tx.from.CheckCloseLending()
}

func (tx *TxCloseLending) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.CloseLending(); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
