// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
)

type TxTransfer struct {
	*TxBase
}

func (tx *TxTransfer) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Amount == nil {
		return fmt.Errorf("TxTransfer invalid amount")
	}

	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxTransfer) VerifyGenesis(blk *Block) error {
	var err error
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)
	tx.cost = tx.cost.Add(tx.cost, tx.ld.Amount)

	bs := blk.State()
	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(tx.ld.To)
	return err
}
