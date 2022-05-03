// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

type TxCloseLending struct {
	*TxBase
}

func (tx *TxCloseLending) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required native LDC", util.EthID(tx.ld.Token))
	}

	if tx.ld.To != ids.ShortEmpty {
		return fmt.Errorf("TxCloseLending invalid to: %s", util.EthID(tx.ld.To).String())
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxCloseLending invalid amount, expected 0, got %v", tx.ld.Amount)
	}
	return nil
}

func (tx *TxCloseLending) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	return tx.from.CheckCloseLending()
}

func (tx *TxCloseLending) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.CloseLending(); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
