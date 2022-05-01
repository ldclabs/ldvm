// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
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
	signers []ids.ShortID
	status  choices.Status
	fee     *big.Int
	tip     *big.Int
	cost    *big.Int // fee + tip + amount
}

func (tx *TxBase) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}

	return tx.ld.MarshalJSON()
}

func (tx *TxBase) LD() *ld.Transaction {
	return tx.ld
}

func (tx *TxBase) ID() ids.ID {
	return tx.ld.ID()
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
	if tx.ld.Token != constants.LDCAccount && util.TokenSymbol(tx.ld.Token).String() == "" {
		return fmt.Errorf("invalid token %s", util.EthID(tx.ld.Token))
	}
	if tx.ld.From == ids.ShortEmpty {
		return fmt.Errorf("invalid from")
	}
	if tx.ld.From == tx.ld.To {
		return fmt.Errorf("invalid to")
	}

	var err error
	tx.signers, err = util.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxBase) Verify(blk *Block, bs BlockState) error {
	feeCfg := blk.FeeConfig()
	requireGas := tx.ld.RequireGas(feeCfg.ThresholdGas)
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
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return err
	}
	if tx.ld.To != ids.ShortEmpty {
		if tx.to, err = bs.LoadAccount(tx.ld.To); err != nil {
			return err
		}
	}

	if !tx.from.SatisfySigning(tx.signers) {
		return fmt.Errorf("sender account need more signers")
	}
	if tx.ld.Nonce != tx.from.Nonce() {
		return fmt.Errorf("sender account nonce not matching, expected %d, got %d",
			tx.from.Nonce(), tx.ld.Nonce)
	}

	ldcB := tx.from.BalanceOf(constants.LDCAccount)
	if ldcB.Cmp(tx.cost) < 0 {
		return fmt.Errorf("sender account insufficient balance for fee, expected %d, got %d",
			tx.cost, ldcB)
	}

	switch tx.from.Type() {
	case ld.TokenAccount:
		switch tx.ld.Type {
		case ld.TypeUpdateAccountKeepers, ld.TypeDestroyTokenAccount:
			// just go ahead
		case ld.TypeTransfer:
			if tx.ld.Token == constants.LDCAccount {
				total := new(big.Int).Add(tx.ld.Amount, tx.cost)
				total.Add(total, blk.FeeConfig().MinTokenPledge)
				if ldcB.Cmp(total) < 0 {
					return fmt.Errorf("sender account insufficient balance, expected %d, got %d",
						total, ldcB)
				}
			}
		default:
			return fmt.Errorf("invalid from account for %s", ld.TxTypeString(tx.ld.Type))
		}
	case ld.StakeAccount:
		switch tx.ld.Type {
		case ld.TypeUpdateAccountKeepers, ld.TypeTakeStake, ld.TypeWithdrawStake, ld.TypeResetStakeAccount:
			// just go ahead
		default:
			return fmt.Errorf("invalid from account for %s", ld.TxTypeString(tx.ld.Type))
		}
	}

	if tx.to != nil {
		switch tx.to.Type() {
		case ld.TokenAccount:
			switch tx.ld.Type {
			case ld.TypeEth, ld.TypeTransfer, ld.TypeExchange, ld.TypeCreateTokenAccount:
				// just go ahead
			default:
				return fmt.Errorf("invalid to account for %s", ld.TxTypeString(tx.ld.Type))
			}
		case ld.StakeAccount:
			switch tx.ld.Type {
			case ld.TypeCreateStakeAccount, ld.TypeTakeStake, ld.TypeWithdrawStake:
				// just go ahead
			default:
				return fmt.Errorf("invalid to account for %s", ld.TxTypeString(tx.ld.Type))
			}
		}
	}

	if tx.ld.Amount != nil {
		switch tx.ld.Token {
		case constants.LDCAccount:
			if total := new(big.Int).Add(tx.ld.Amount, tx.cost); ldcB.Cmp(total) < 0 {
				return fmt.Errorf("sender account insufficient balance, expected %d, got %d",
					total, ldcB)
			}
		default:
			tokenB := tx.from.BalanceOf(tx.ld.Token)
			if tokenB.Cmp(tx.ld.Amount) < 0 {
				return fmt.Errorf("sender account %s insufficient balance, expected %d, got %d",
					tx.ld.Token.String(), tx.ld.Amount, tokenB)
			}
		}
	}
	return nil
}

func (tx *TxBase) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.SubByNonce(constants.LDCAccount, tx.ld.Nonce, tx.cost); err != nil {
		return err
	}

	if tx.ld.Amount != nil && tx.ld.From != tx.ld.To {
		if err = tx.from.Sub(tx.ld.Token, tx.ld.Amount); err != nil {
			return err
		}
		if err = tx.to.Add(tx.ld.Token, tx.ld.Amount); err != nil {
			return err
		}
	}
	if err = tx.miner.Add(constants.LDCAccount, tx.tip); err != nil {
		return err
	}
	// burn fee
	if err = tx.ldc.Add(constants.LDCAccount, tx.fee); err != nil {
		return err
	}
	return nil
}

func (tx *TxBase) Event(ts int64) *Event {
	return nil
}
