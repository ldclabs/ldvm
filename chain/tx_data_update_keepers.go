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

type TxUpdateDataKeepers struct {
	TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
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
	if tx.data.ID == nil ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid txUpdater")
	}
	if len(tx.data.Keepers) == 0 && tx.data.Approver == nil {
		return fmt.Errorf("TxUpdateDataKeepers no keepers nor approver")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.data.ID)
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
	if tx.dm.Approver != nil && !tx.signers.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateDataKeepers no approver signing")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(blk *Block, bs BlockState) error {
	var err error

	tx.dm.Version++
	if tx.data.Approver != nil {
		if *tx.data.Approver == util.EthIDEmpty {
			tx.dm.Approver = nil
		} else {
			tx.dm.Approver = tx.data.Approver
		}
	}
	if len(tx.data.Keepers) > 0 {
		tx.dm.Threshold = tx.data.Threshold
		tx.dm.Keepers = tx.data.Keepers
	}
	if err = bs.SaveData(*tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
