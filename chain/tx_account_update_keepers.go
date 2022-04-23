// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxAccountUpdateKeepers struct {
	*TxBase
	data *ld.TxUpdater
}

func (tx *TxAccountUpdateKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxAccountUpdateKeepers unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

// VerifyGenesis skipping signature verification
func (tx *TxAccountUpdateKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxAccountUpdateKeepers invalid")
	}
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxAccountUpdateKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxAccountUpdateKeepers SyntacticVerify failed: %v", err)
	}
	if len(tx.data.Keepers) == 0 ||
		tx.data.Threshold == 0 {
		return fmt.Errorf("TxAccountUpdateKeepers invalid keepers")
	}
	return nil
}

func (tx *TxAccountUpdateKeepers) VerifyGenesis(blk *Block) error {
	var err error
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	bs := blk.State()
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxAccountUpdateKeepers) Accept(blk *Block) error {
	var err error
	if err = tx.from.UpdateKeepers(tx.data.Threshold, tx.data.Keepers); err != nil {
		return err
	}

	return tx.TxBase.Accept(blk)
}
