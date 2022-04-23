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

type TxUpdateDataKeepers struct {
	*TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateDataKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid")
	}
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid txUpdater")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxUpdateDataKeepers no keepers")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	bs := blk.State()
	tx.dm, err = bs.LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepers load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepers version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("TxUpdateDataKeepers need more signatures")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(blk *Block) error {
	var err error

	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}
	if err = blk.State().SaveData(tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
