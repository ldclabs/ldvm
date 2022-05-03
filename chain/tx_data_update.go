// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	jsonpatch "github.com/evanphx/json-patch/v5"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateData struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxUpdater
	dm        *ld.DataMeta
	prevDM    *ld.DataMeta
	jsonPatch jsonpatch.Patch
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateData) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateData invalid")
	}

	if len(tx.ld.ExSignatures) > 0 {
		tx.exSigners, err = util.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.ExSignatures)
		if err != nil {
			return fmt.Errorf("TxUpdateData invalid exSignatures")
		}
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateData SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxUpdateData invalid TxUpdater for TxUpdateData")
	}
	return nil
}

func (tx *TxUpdateData) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateData load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("TxUpdateData need more signatures")
	}

	tx.prevDM = tx.dm.Copy()
	switch util.ModelID(tx.dm.ModelID) {
	case constants.RawModelID:
		tx.dm.Version++
		tx.dm.Data = tx.data.Data
		return nil
	case constants.JsonModelID:
		tx.jsonPatch, err = jsonpatch.DecodePatch(tx.data.Data)
		if err != nil {
			return fmt.Errorf("TxUpdateData invalid JSON patch: %v", err)
		}

		tx.dm.Data, err = tx.jsonPatch.Apply(tx.dm.Data)
		if err != nil {
			return fmt.Errorf("TxUpdateData apply patch failed: %v", err)
		}
		tx.dm.Version++
		return tx.dm.SyntacticVerify()
	}

	mm, err := bs.LoadModel(tx.dm.ModelID)
	if err != nil {
		return fmt.Errorf("TxUpdateData load data model failed: %v", err)
	}

	if !util.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("TxUpdateData need more exSignatures")
	}
	// TODO: apply patch operations
	// tx.data.Validate(mm.SchemaType)
	if blk.ctx.Chain().IsNameService(tx.dm.ModelID) {
		// TODO: should not update name
	}
	return nil
}

func (tx *TxUpdateData) Accept(blk *Block, bs BlockState) error {
	var err error

	if err = bs.SavePrevData(tx.data.ID, tx.prevDM); err != nil {
		return err
	}
	if err = bs.SaveData(tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}

func (tx *TxUpdateData) Event(ts int64) *Event {
	e := NewEvent(tx.data.ID, SrcData, ActionUpdate)
	e.Time = ts
	return nil
}
