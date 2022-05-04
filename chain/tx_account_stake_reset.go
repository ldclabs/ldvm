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

type TxResetStakeAccount struct {
	TxBase
	data *ld.StakeConfig
}

func (tx *TxResetStakeAccount) MarshalJSON() ([]byte, error) {
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

func (tx *TxResetStakeAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("invalid token %s, required LDC", tx.ld.Token)
	}
	if !util.ValidStakeAddress(tx.ld.From) {
		return fmt.Errorf("TxResetStakeAccount invalid stake address: %s", tx.ld.From)
	}
	if tx.ld.Amount == nil || tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxResetStakeAccount invalid amount")
	}

	// reset stake account
	if tx.ld.To == util.EthIDEmpty {
		if len(tx.ld.Data) == 0 {
			return fmt.Errorf("TxResetStakeAccount invalid data")
		}
		tx.data = &ld.StakeConfig{}
		if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
			return fmt.Errorf("TxResetStakeAccount unmarshal data failed: %v", err)
		}
		if err = tx.data.SyntacticVerify(); err != nil {
			return fmt.Errorf("TxResetStakeAccount SyntacticVerify failed: %v", err)
		}
		if tx.data.LockTime < tx.ld.Timestamp {
			return fmt.Errorf("TxResetStakeAccount invalid lockTime")
		}
	}
	return nil
}

func (tx *TxResetStakeAccount) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("sender account need more signers")
	}
	switch tx.ld.To {
	case util.EthIDEmpty:
		return tx.from.CheckResetStake(tx.data)
	default:
		return tx.from.CheckDestroyStake(tx.to)
	}
}

func (tx *TxResetStakeAccount) Accept(blk *Block, bs BlockState) error {
	if err := tx.TxBase.Accept(blk, bs); err != nil {
		return err
	}
	// do it after TxBase.Accept
	switch tx.ld.To {
	case util.EthIDEmpty:
		return tx.from.ResetStake(tx.data)
	default:
		return tx.from.DestroyStake(tx.to)
	}
}
