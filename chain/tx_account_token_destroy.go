// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

type TxDestroyTokenAccount struct {
	*TxBase
}

func (tx *TxDestroyTokenAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if util.TokenSymbol(tx.ld.From).String() == "" {
		return fmt.Errorf("TxDestroyTokenAccount invalid token: %s", util.EthID(tx.ld.From))
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}
	// destroy
	ldcB := tx.from.BalanceOf(constants.LDCAccount)
	tx.cost = new(big.Int).Add(tx.tip, tx.fee)
	tx.ld.Amount = new(big.Int).Sub(ldcB, tx.cost)
	if tx.ld.Amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.TxDestroyTokenAccount %s insufficient balance to destroy, expected %v, got %v",
			util.EthID(tx.ld.From), tx.cost, ldcB)
	}
	tx.cost = tx.cost.Set(ldcB)
	return nil
}

func (tx *TxDestroyTokenAccount) Accept(blk *Block) error {
	if err := tx.from.DestroyToken(tx.ld.From, tx.ld.To); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
