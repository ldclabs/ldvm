// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
)

type TxTransfer struct {
	TxBase
}

func (tx *TxTransfer) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransfer.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("TxTransfer.SyntacticVerify failed: invalid to")

	case tx.ld.Amount == nil:
		return fmt.Errorf("TxTransfer.SyntacticVerify failed: invalid amount")
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxTransfer) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.amount = new(big.Int).Set(tx.ld.Amount)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return err
	}
	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return err
	}

	if tx.to, err = bs.LoadAccount(*tx.ld.To); err != nil {
		return err
	}
	return nil
}
