// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxResetStakeAccount struct {
	*TxBase
	data *ld.TxMinter
}

func (tx *TxResetStakeAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil && len(tx.ld.Data) > 0 {
		tx.data = &ld.TxMinter{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxResetStakeAccount unmarshal failed: %v", err)
		}
	}

	if tx.data != nil {
		d, err := tx.data.MarshalJSON()
		if err != nil {
			return nil, err
		}
		v.Data = d
	}
	return v.MarshalJSON()
}

func (tx *TxResetStakeAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if !util.ValidStakeAddress(tx.ld.From) {
		return fmt.Errorf("TxCreateStakeAccount invalid stake address: %s", util.EthID(tx.ld.From))
	}

	if len(tx.ld.Data) > 0 {
		tx.data = &ld.TxMinter{}
		if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
			return fmt.Errorf("TxCreateStakeAccount unmarshal data failed: %v", err)
		}
		if err = tx.data.SyntacticVerify(); err != nil {
			return fmt.Errorf("TxCreateStakeAccount SyntacticVerify failed: %v", err)
		}
		if tx.data.LockTime < uint64(time.Now().Unix()) {
			return fmt.Errorf("TxCreateStakeAccount invalid lockTime")
		}
		if tx.data.DelegationFee < 1 || tx.data.DelegationFee > 500 {
			return fmt.Errorf("TxCreateStakeAccount invalid delegationFee")
		}
	}
	return nil
}

func (tx *TxResetStakeAccount) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	switch {
	case tx.data == nil:
		// destroy
		ldcB := tx.from.BalanceOf(constants.LDCAccount)
		tx.cost = new(big.Int).Add(tx.tip, tx.fee)
		tx.ld.Amount = new(big.Int).Sub(ldcB, tx.cost)
		if tx.ld.Amount.Sign() < 0 {
			return fmt.Errorf(
				"Account.TxResetStakeAccount %s insufficient balance to destroy, expected %v, got %v",
				util.EthID(tx.ld.From), tx.cost, ldcB)
		}
		tx.cost = tx.cost.Set(ldcB)
	default:
		// reset
		feeCfg := blk.FeeConfig()
		if tx.data.DelegationFee < feeCfg.MinDelegationFee {
			return fmt.Errorf("TxResetStakeAccount invalid delegationFee")
		}
	}
	return nil
}

func (tx *TxResetStakeAccount) Accept(blk *Block, bs BlockState) error {
	var err error
	switch {
	case tx.data == nil:
		// destroy
		if err = tx.from.DestroyStake(tx.ld.To); err != nil {
			return err
		}
	default:
		// reset
		if err = tx.from.ResetStake(tx.ld.To, tx.data); err != nil {
			return err
		}
	}

	return tx.TxBase.Accept(blk, bs)
}
