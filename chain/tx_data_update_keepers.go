// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepers struct {
	TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateDataKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("invalid token %s, required LDC", tx.ld.Token)
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

func (tx *TxUpdateDataKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(util.DataID(tx.data.ID))
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepers load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepers version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateDataKeepers need more signatures")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(blk *Block, bs BlockState) error {
	var err error

	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}
	if err = bs.SaveData(util.DataID(tx.data.ID), tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
