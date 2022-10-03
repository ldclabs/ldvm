// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxBase struct {
	ld      *ld.Transaction
	ldc     *Account // native token account
	miner   *Account
	from    *Account
	to      *Account
	amount  *big.Int
	fee     *big.Int
	tip     *big.Int
	cost    *big.Int // fee + tip
	token   util.TokenSymbol
	signers util.EthIDs
}

func (tx *TxBase) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	return util.ErrPrefix("TxBase.MarshalJSON error: ").
		ErrorMap(json.Marshal(tx.ld))
}

func (tx *TxBase) LD() *ld.Transaction {
	return tx.ld
}

func (tx *TxBase) ID() ids.ID {
	return tx.ld.ID
}

func (tx *TxBase) Type() ld.TxType {
	return tx.ld.Tx.Type
}

func (tx *TxBase) Bytes() []byte {
	return tx.ld.Tx.Bytes()
}

func (tx *TxBase) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxBase.SyntacticVerify error: ")

	switch {
	case tx == nil || tx.ld == nil:
		return errp.Errorf("nil pointer")

	case tx.ld.Tx.From == util.EthIDEmpty:
		return errp.Errorf("invalid from")

	case tx.ld.Tx.To != nil && tx.ld.Tx.From == *tx.ld.Tx.To:
		return errp.Errorf("invalid to")
	}

	tx.signers, err = tx.ld.Signers()
	if err != nil {
		return errp.ErrorIf(err)
	}
	if tx.ld.Tx.Token != nil {
		tx.token = *tx.ld.Tx.Token
	}
	tx.amount = new(big.Int)
	if tx.ld.Tx.Amount != nil {
		tx.amount.Set(tx.ld.Tx.Amount)
	}
	return nil
}

func (tx *TxBase) verify(ctx ChainContext, cs ChainState) error {
	var err error
	feeCfg := ctx.FeeConfig()

	if price := ctx.GasPrice().Uint64(); tx.ld.Tx.GasFeeCap < price {
		return fmt.Errorf("invalid gasFeeCap, expected >= %d, got %d",
			price, tx.ld.Tx.GasFeeCap)
	}

	gas := tx.ld.Gas()
	if gas > feeCfg.MaxTxGas {
		return fmt.Errorf("gas too large, expected <= %d, got %d",
			feeCfg.MaxTxGas, gas)
	}
	gb := new(big.Int).SetUint64(gas)
	tx.tip = new(big.Int).Mul(gb, new(big.Int).SetUint64(tx.ld.Tx.GasTip))
	tx.fee = new(big.Int).Mul(gb, ctx.GasPrice())
	tx.cost = new(big.Int).Add(tx.tip, tx.fee)

	if tx.ldc, err = cs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = cs.LoadMiner(ctx.Miner()); err != nil {
		return err
	}
	if tx.from, err = cs.LoadAccount(tx.ld.Tx.From); err != nil {
		return err
	}
	if err = tx.from.CheckAsFrom(tx.ld.Tx.Type); err != nil {
		return err
	}
	if tx.ld.Tx.To != nil {
		if tx.to, err = cs.LoadAccount(*tx.ld.Tx.To); err != nil {
			return err
		}
		if err = tx.to.CheckAsTo(tx.ld.Tx.Type); err != nil {
			return err
		}
	}

	switch {
	case tx.ld.Tx.Nonce != tx.from.Nonce():
		return fmt.Errorf("invalid nonce for sender, expected %d, got %d",
			tx.from.Nonce(), tx.ld.Tx.Nonce)

	case !tx.from.SatisfySigning(tx.signers):
		return fmt.Errorf("invalid signatures for sender")

	case tx.ld.NeedApprove(tx.from.ld.Approver, tx.from.ld.ApproveList) &&
		!tx.signers.Has(*tx.from.ld.Approver):
		return fmt.Errorf("invalid signature for approver")
	}

	switch tx.token {
	case constants.NativeToken:
		if err = tx.from.CheckBalance(constants.NativeToken,
			new(big.Int).Add(tx.amount, tx.cost)); err != nil {
			return err
		}

	default:
		if err = tx.from.CheckBalance(constants.NativeToken, tx.cost); err != nil {
			return err
		}
		if err = tx.from.CheckBalance(tx.token, tx.amount); err != nil {
			return err
		}
	}
	return nil
}

func (tx *TxBase) accept(ctx ChainContext, cs ChainState) error {
	var err error
	if err = tx.from.SubByNonce(constants.NativeToken, tx.ld.Tx.Nonce, tx.cost); err != nil {
		return err
	}

	if tx.amount.Sign() > 0 {
		if err = tx.from.Sub(tx.token, tx.amount); err != nil {
			return err
		}
		if err = tx.to.Add(tx.token, tx.amount); err != nil {
			return err
		}
	}
	if err = tx.miner.Add(constants.NativeToken, tx.tip); err != nil {
		return err
	}
	// burning fee
	if err = tx.ldc.Add(constants.NativeToken, tx.fee); err != nil {
		return err
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxBase) Apply(ctx ChainContext, cs ChainState) error {
	errp := util.ErrPrefix("TxBase.Apply error: ")
	if err := tx.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.accept(ctx, cs))
}
