// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxWithdrawStake struct {
	*TxBase
	data *ld.TxMinter
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
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

func (tx *TxWithdrawStake) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}

	if tx.ld.Amount != nil {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	if !util.ValidStakeAddress(tx.ld.To) {
		return fmt.Errorf("TxWithdrawStake invalid stake address: %s", util.EthID(tx.ld.To))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxWithdrawStake invalid")
	}
	tx.data = &ld.TxMinter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxWithdrawStake unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxWithdrawStake SyntacticVerify failed: %v", err)
	}

	if tx.data.Amount == nil || tx.data.Amount.Sign() <= 0 {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	return nil
}

func (tx *TxWithdrawStake) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}
	if tx.to.IsEmpty() {
		return fmt.Errorf("TxTakeStake invalid address, stake account %s not exists", util.EthID(tx.ld.To))
	}
	return nil
}

func (tx *TxWithdrawStake) Accept(blk *Block) error {
	withdraw, err := tx.to.WithdrawStake(tx.ld.From, tx.data.Amount)
	if err != nil {
		return err
	}
	if err = tx.to.Sub(constants.LDCAccount, withdraw); err != nil {
		return err
	}
	if err = tx.from.Add(constants.LDCAccount, withdraw); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
