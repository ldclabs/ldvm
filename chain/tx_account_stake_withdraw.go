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
	data *ld.TxTransfer
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
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

	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	if !util.ValidStakeAddress(tx.ld.To) {
		return fmt.Errorf("TxWithdrawStake invalid stake address: %s", util.EthID(tx.ld.To))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxWithdrawStake invalid")
	}
	tx.data = &ld.TxTransfer{}
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

func (tx *TxWithdrawStake) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	_, err = tx.to.CheckWithdrawStake(tx.ld.Token, tx.ld.From, tx.data.Amount)
	return err
}

func (tx *TxWithdrawStake) Accept(blk *Block, bs BlockState) error {
	withdraw, err := tx.to.WithdrawStake(tx.ld.Token, tx.ld.From, tx.data.Amount)
	if err != nil {
		return err
	}
	if err = tx.to.Sub(tx.ld.Token, withdraw); err != nil {
		return err
	}
	if err = tx.from.Add(tx.ld.Token, withdraw); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
