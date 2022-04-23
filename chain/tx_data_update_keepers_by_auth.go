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

type TxUpdateDataKeepersByAuth struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateDataKeepersByAuth) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateDataKeepersByAuth) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateModelKeepers invalid")
	}

	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("invalid exSignatures: %v", err)
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth syntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 ||
		tx.data.Amount == nil ||
		tx.data.Amount.Cmp(tx.ld.Amount) != 0 ||
		tx.data.To != tx.ld.To {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	bs := blk.State()
	tx.dm, err = bs.LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepersByAuth version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	// verify seller's signatures
	if !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners, false) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth need more exSignatures")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Accept(blk *Block) error {
	var err error

	bs := blk.State()
	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if len(tx.dm.Keepers) == 0 {
		tx.dm.Threshold = tx.from.Threshold()
		tx.dm.Keepers = tx.from.Keepers()
	}
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}
	if err = bs.SaveData(tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
