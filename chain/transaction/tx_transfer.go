// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

type TxTransfer struct {
	TxBase
}

func (tx *TxTransfer) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxTransfer.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("invalid to")

	case tx.ld.Amount == nil:
		return errp.Errorf("invalid amount")
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxTransfer) ApplyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxTransfer.ApplyGenesis error: ")

	tx.amount = new(big.Int).Set(tx.ld.Amount)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.to, err = bs.LoadAccount(*tx.ld.To); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
