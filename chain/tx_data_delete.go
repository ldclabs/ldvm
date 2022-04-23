// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxDeleteData struct {
	*TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxDeleteData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxDeleteData unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxDeleteData) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxDeleteData invalid")
	}
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxDeleteData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxDeleteData SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxDeleteData invalid TxUpdater")
	}
	return nil
}

func (tx *TxDeleteData) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	bs := blk.State()
	tx.dm, err = bs.LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxDeleteData load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxDeleteData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("TxDeleteData need more signatures")
	}
	return nil
}

func (tx *TxDeleteData) Accept(blk *Block) error {
	var err error

	tx.dm.Version = 0 // mark dropped
	if err = blk.State().SaveData(tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}

func (tx *TxDeleteData) Event(ts int64) *Event {
	e := NewEvent(tx.data.ID, SrcData, ActionDelete)
	e.Time = ts
	return e
}