// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"time"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateStakeAccount struct {
	*TxBase
	data *ld.TxMinter
}

func (tx *TxCreateStakeAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxMinter{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxCreateStakeAccount unmarshal failed: %v", err)
		}
	}

	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxCreateStakeAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if !util.ValidStakeAddress(tx.ld.To) {
		return fmt.Errorf("TxCreateStakeAccount invalid stake address: %s", util.EthID(tx.ld.To))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid")
	}
	tx.data = &ld.TxMinter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateStakeAccount unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateStakeAccount SyntacticVerify failed: %v", err)
	}

	if tx.data.Threshold == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid threshold")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid keepers")
	}
	if tx.data.LockTime < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxCreateStakeAccount invalid lockTime")
	}
	if tx.data.DelegationFee < 1 || tx.data.DelegationFee > 500 {
		return fmt.Errorf("TxCreateStakeAccount invalid delegationFee")
	}
	return nil
}

func (tx *TxCreateStakeAccount) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}
	if !tx.to.IsEmpty() {
		return fmt.Errorf("TxCreateStakeAccount invalid address, stake account %s exists", util.EthID(tx.ld.To))
	}
	feeCfg := blk.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinValidatorStake) < 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected >= %v, got %v",
			feeCfg.MinValidatorStake, tx.ld.Amount)
	}
	if tx.ld.Amount.Cmp(feeCfg.MaxValidatorStake) > 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected <= %v, got %v",
			feeCfg.MaxValidatorStake, tx.ld.Amount)
	}
	if tx.data.DelegationFee < feeCfg.MinDelegationFee {
		return fmt.Errorf("TxCreateStakeAccount invalid delegationFee")
	}
	return nil
}

func (tx *TxCreateStakeAccount) Accept(blk *Block) error {
	var err error
	if err = tx.to.CreateStake(tx.ld.From, tx.ld.Amount, tx.data); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
