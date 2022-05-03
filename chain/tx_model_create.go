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

type TxCreateModel struct {
	*TxBase
	mm *ld.ModelMeta
}

func (tx *TxCreateModel) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.mm == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := tx.mm.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxCreateModel) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateModel invalid")
	}
	tx.mm = &ld.ModelMeta{}
	if err = tx.mm.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.mm.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateModel) VerifyGenesis(blk *Block, bs BlockState) error {
	var err error
	tx.mm = &ld.ModelMeta{}
	if err = tx.mm.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.mm.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateModel) Accept(blk *Block, bs BlockState) error {
	var err error

	if err = bs.SaveModel(tx.ld.ShortID(), tx.mm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}

func (tx *TxCreateModel) Event(ts int64) *Event {
	e := NewEvent(tx.ld.ShortID(), SrcModel, ActionAdd)
	e.Time = ts
	return e
}
