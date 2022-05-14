// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
)

type TxRepay struct {
	TxBase
}

func (tx *TxRepay) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.To == nil {
		return fmt.Errorf("TxRepay invalid to")
	}
	if tx.ld.Amount.Sign() == 0 {
		return fmt.Errorf("TxRepay invalid amount, got 0")
	}
	return nil
}

func (tx *TxRepay) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	_, err = tx.to.CheckRepay(tx.token, tx.ld.From, tx.ld.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (tx *TxRepay) Accept(blk *Block, bs BlockState) error {
	actual, err := tx.to.Repay(tx.token, tx.ld.From, tx.ld.Amount)
	if err != nil {
		return err
	}
	tx.ld.Amount.Set(actual)
	return tx.TxBase.Accept(blk, bs)
}