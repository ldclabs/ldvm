// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateStakeAccount struct {
	TxBase
	data  *ld.TxAccounter
	stake *ld.StakeConfig
}

func (tx *TxCreateStakeAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}

	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxCreateStakeAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("TxCreateStakeAccount invalid token %s, required LDC", tx.ld.Token)
	}
	if !util.ValidStakeAddress(tx.ld.To) {
		return fmt.Errorf("TxCreateStakeAccount invalid stake address: %s", tx.ld.To)
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid")
	}

	tx.data = &ld.TxAccounter{}
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
	if tx.data.Amount.Sign() != 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, please take stake after created")
	}

	tx.stake = &ld.StakeConfig{}
	if err = tx.stake.Unmarshal(tx.data.Data); err != nil {
		return fmt.Errorf("TxCreateStakeAccount unmarshal data failed: %v", err)
	}
	if err = tx.stake.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateStakeAccount SyntacticVerify failed: %v", err)
	}
	if tx.stake.LockTime < tx.ld.Timestamp {
		return fmt.Errorf("TxCreateStakeAccount invalid lockTime")
	}
	return nil
}

func (tx *TxCreateStakeAccount) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	feeCfg := blk.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinStakePledge) < 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected >= %v, got %v",
			feeCfg.MinStakePledge, tx.ld.Amount)
	}
	return tx.to.CheckCreateStake(tx.ld.From, tx.ld.Amount, tx.data, tx.stake)
}

func (tx *TxCreateStakeAccount) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.to.CreateStake(tx.ld.From, tx.ld.Amount, tx.data, tx.stake); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
