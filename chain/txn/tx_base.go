// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/chain/acct"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxBase struct {
	ld        *ld.Transaction
	ldc       *acct.Account // native token account
	miner     *acct.Account
	from      *acct.Account
	to        *acct.Account
	amount    *big.Int
	fee       *big.Int
	tip       *big.Int
	cost      *big.Int // fee + tip
	token     ids.TokenSymbol
	senderKey signer.Key
}

func (tx *TxBase) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	return erring.ErrPrefix("txn.TxBase.MarshalJSON: ").
		ErrorMap(json.Marshal(tx.ld))
}

func (tx *TxBase) LD() *ld.Transaction {
	return tx.ld
}

func (tx *TxBase) ID() ids.ID32 {
	return tx.ld.ID
}

func (tx *TxBase) Type() ld.TxType {
	return tx.ld.Tx.Type
}

func (tx *TxBase) Gas() uint64 {
	return tx.ld.Gas()
}

func (tx *TxBase) Bytes() []byte {
	return tx.ld.Tx.Bytes()
}

func (tx *TxBase) SyntacticVerify() error {
	errp := erring.ErrPrefix("txn.TxBase.SyntacticVerify: ")

	switch {
	case tx == nil || tx.ld == nil:
		return errp.Errorf("nil pointer")

	case tx.ld.Tx.From == ids.EmptyAddress:
		return errp.Errorf("invalid from")

	case tx.ld.Tx.To != nil && tx.ld.Tx.From == *tx.ld.Tx.To:
		return errp.Errorf("invalid to")

	case len(tx.ld.Signatures) == 0:
		return errp.Errorf("no signatures")
	}

	tx.senderKey = signer.Key(tx.ld.Tx.From.Bytes())
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

	if tx.ldc, err = cs.LoadAccount(ids.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = cs.LoadAccount(ctx.Builder()); err != nil {
		return err
	}
	if tx.from, err = cs.LoadAccount(tx.ld.Tx.From); err != nil {
		return err
	}

	if err = tx.from.LD().CheckAsFrom(tx.ld.Tx.Type); err != nil {
		return err
	}

	if tx.ld.Tx.To != nil {
		if tx.to, err = cs.LoadAccount(*tx.ld.Tx.To); err != nil {
			return err
		}
		if err = tx.to.LD().CheckAsTo(tx.ld.Tx.Type); err != nil {
			return err
		}
	}

	switch {
	case tx.ld.Tx.Nonce != tx.from.Nonce():
		return fmt.Errorf("invalid nonce for sender, expected %d, got %d",
			tx.from.Nonce(), tx.ld.Tx.Nonce)

	case tx.ld.Tx.Type != ld.TypeEth &&
		!tx.from.Verify(tx.ld.TxHash(), tx.ld.Signatures, tx.senderKey):
		return fmt.Errorf("invalid signatures for sender")

	case !tx.ld.IsApproved(tx.from.LD().Approver, tx.from.LD().ApproveList, false):
		return fmt.Errorf("invalid signature for approver")
	}

	switch tx.token {
	case ids.NativeToken:
		if err = tx.from.CheckBalance(ids.NativeToken,
			new(big.Int).Add(tx.amount, tx.cost), tx.amount.Sign() > 0); err != nil {
			return err
		}

	default:
		if err = tx.from.CheckBalance(ids.NativeToken, tx.cost, false); err != nil {
			return err
		}
		if err = tx.from.CheckBalance(tx.token, tx.amount, false); err != nil {
			return err
		}
	}
	return nil
}

func (tx *TxBase) accept(ctx ChainContext, cs ChainState) error {
	var err error
	if err = tx.from.SubGasByNonce(ids.NativeToken, tx.ld.Tx.Nonce, tx.cost); err != nil {
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
	if err = tx.miner.Add(ids.NativeToken, tx.tip); err != nil {
		return err
	}
	// burning fee
	if err = tx.ldc.Add(ids.NativeToken, tx.fee); err != nil {
		return err
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxBase) Apply(ctx ChainContext, cs ChainState) error {
	errp := erring.ErrPrefix("txn.TxBase.Apply: ")
	if err := tx.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.accept(ctx, cs))
}
