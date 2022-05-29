// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
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
	status  choices.Status
}

func (tx *TxBase) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	return json.Marshal(tx.ld)
}

func (tx *TxBase) LD() *ld.Transaction {
	return tx.ld
}

func (tx *TxBase) ID() ids.ID {
	return tx.ld.ID
}

func (tx *TxBase) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxBase) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxBase) Status() string {
	return tx.status.String()
}

func (tx *TxBase) SetStatus(s choices.Status) {
	tx.status = s
}

func (tx *TxBase) SyntacticVerify() error {
	if tx == nil || tx.ld == nil {
		return fmt.Errorf("TxBase.SyntacticVerify failed: nil pointer")
	}
	if tx.ld.From == util.EthIDEmpty {
		return fmt.Errorf("TxBase.SyntacticVerify failed: invalid from")
	}
	if tx.ld.To != nil && tx.ld.From == *tx.ld.To {
		return fmt.Errorf("TxBase.SyntacticVerify failed: invalid to")
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxBase) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	feeCfg := bctx.FeeConfig()
	requireGas := tx.ld.RequiredGas(feeCfg.ThresholdGas)
	if price := bctx.GasPrice().Uint64(); tx.ld.GasFeeCap < price {
		return fmt.Errorf("TxBase.Verify failed: invalid gasFeeCap, expected >= %d, got %d",
			price, tx.ld.GasFeeCap)
	}
	if tx.ld.Gas != requireGas || tx.ld.Gas > feeCfg.MaxTxGas {
		return fmt.Errorf("TxBase.Verify failed: invalid gas, expected %d, got %d",
			requireGas, tx.ld.Gas)
	}

	tx.signers, err = tx.ld.Signers()
	if err != nil {
		return fmt.Errorf("TxBase.SyntacticVerify failed: %v", err)
	}
	if tx.ld.Token != nil {
		tx.token = *tx.ld.Token
	}
	tx.amount = new(big.Int)
	if tx.ld.Amount != nil {
		tx.amount.Set(tx.ld.Amount)
	}
	tx.tip = new(big.Int).Mul(tx.ld.GasUnits(), new(big.Int).SetUint64(tx.ld.GasTip))
	tx.fee = new(big.Int).Mul(tx.ld.GasUnits(), bctx.GasPrice())
	tx.cost = new(big.Int).Add(tx.tip, tx.fee)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return err
	}
	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return err
	}
	if err = tx.from.CheckAsFrom(tx.ld.Type); err != nil {
		return err
	}
	if tx.ld.To != nil {
		if tx.to, err = bs.LoadAccount(*tx.ld.To); err != nil {
			return err
		}
		if err = tx.to.CheckAsTo(tx.ld.Type); err != nil {
			return err
		}
	}
	if tx.ld.Nonce != tx.from.Nonce() {
		return fmt.Errorf("TxBase.Verify failed: invalid nonce for sender, expected %d, got %d",
			tx.from.Nonce(), tx.ld.Nonce)
	}
	if !tx.from.SatisfySigning(tx.signers) {
		return fmt.Errorf("TxBase.Verify failed: invalid signatures for sender")
	}
	if tx.ld.NeedApprove(tx.from.ld.Approver, tx.from.ld.ApproveList) &&
		!tx.signers.Has(*tx.from.ld.Approver) {
		return fmt.Errorf("TxBase.Verify failed: TxBase.Verify: invalid signature for approver")
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

func (tx *TxBase) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.SubByNonce(constants.NativeToken, tx.ld.Nonce, tx.cost); err != nil {
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
