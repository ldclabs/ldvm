// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxTransfer struct {
	TxBase
}

func (tx *TxTransfer) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxTransfer.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("invalid to")

	case tx.ld.Tx.Amount == nil:
		return errp.Errorf("invalid amount")
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxTransfer) ApplyGenesis(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxTransfer.ApplyGenesis: ")

	tx.amount = new(big.Int).Set(tx.ld.Tx.Amount)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)
	if tx.ldc, err = cs.LoadAccount(ids.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = cs.LoadAccount(ctx.Builder()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.from, err = cs.LoadAccount(tx.ld.Tx.From); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.to, err = cs.LoadAccount(*tx.ld.Tx.To); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
