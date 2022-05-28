// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

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
	ld         *ld.Transaction
	genesisAcc *Account // native token account
	miner      *Account
	from       *Account
	to         *Account
	token      util.TokenSymbol
	signers    util.EthIDs
	status     choices.Status
	fee        *big.Int
	tip        *big.Int
	cost       *big.Int // fee + tip + amount
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
		return fmt.Errorf("tx is nil")
	}
	if tx.ld.Token != nil {
		if tx.token = *tx.ld.Token; !tx.token.Valid() {
			return fmt.Errorf("invalid token %s", tx.token.GoString())
		}
	}
	if tx.ld.From == util.EthIDEmpty {
		return fmt.Errorf("invalid from")
	}
	if tx.ld.To != nil && tx.ld.From == *tx.ld.To {
		return fmt.Errorf("invalid to")
	}

	var err error
	tx.signers, err = tx.ld.Signers()
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxBase) Verify(blk *Block, bs BlockState) error {
	feeCfg := blk.FeeConfig()
	requireGas := tx.ld.RequiredGas(feeCfg.ThresholdGas)
	if price := blk.GasPrice().Uint64(); tx.ld.GasFeeCap < price {
		return fmt.Errorf("tx gasFeeCap not matching, require %d", price)
	}
	if tx.ld.Gas != requireGas || tx.ld.Gas > feeCfg.MaxTxGas {
		return fmt.Errorf("tx gas not matching, require %d", requireGas)
	}

	tx.tip = new(big.Int).Mul(tx.ld.GasUnits(), new(big.Int).SetUint64(tx.ld.GasTip))
	tx.fee = new(big.Int).Mul(tx.ld.GasUnits(), blk.GasPrice())
	tx.cost = new(big.Int).Add(tx.tip, tx.fee)

	var err error
	if tx.genesisAcc, err = bs.LoadAccount(constants.GenesisAccount); err != nil {
		return err
	}
	if tx.miner, err = bs.LoadMiner(blk.ld.Miner); err != nil {
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

	if !tx.from.SatisfySigning(tx.signers) {
		return fmt.Errorf("sender account need more signers")
	}
	if tx.ld.NeedApprove(tx.from.ld.Approver, tx.from.ld.ApproveList) && !tx.signers.Has(*tx.from.ld.Approver) {
		return fmt.Errorf("TxBase.Verify: no approver signing")
	}
	if tx.ld.Nonce != tx.from.Nonce() {
		return fmt.Errorf("sender account nonce not matching, expected %d, got %d",
			tx.from.Nonce(), tx.ld.Nonce)
	}

	switch tx.token {
	case constants.NativeToken:
		if err = tx.from.CheckBalance(constants.NativeToken, new(big.Int).Add(tx.ld.Amount, tx.cost)); err != nil {
			return err
		}
	default:
		if err = tx.from.CheckBalance(constants.NativeToken, tx.cost); err != nil {
			return fmt.Errorf("check fee failed: %v", err)
		}
		if err = tx.from.CheckBalance(tx.token, tx.ld.Amount); err != nil {
			return err
		}
	}
	return nil
}

func (tx *TxBase) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.SubByNonce(constants.NativeToken, tx.ld.Nonce, tx.cost); err != nil {
		return err
	}

	if tx.ld.Amount.Sign() > 0 {
		if err = tx.from.Sub(tx.token, tx.ld.Amount); err != nil {
			return err
		}
		if err = tx.to.Add(tx.token, tx.ld.Amount); err != nil {
			return err
		}
	}
	if err = tx.miner.Add(constants.NativeToken, tx.tip); err != nil {
		return err
	}
	// revoke fee to genesis account
	if err = tx.genesisAcc.Add(constants.NativeToken, tx.fee); err != nil {
		return err
	}
	return nil
}

func (tx *TxBase) Event(ts int64) *Event {
	return nil
}
