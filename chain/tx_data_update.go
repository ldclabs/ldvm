// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateData struct {
	TxBase
	exSigners util.EthIDs
	data      *ld.TxUpdater
	dm        *ld.DataMeta
	prevDM    *ld.DataMeta
	jsonPatch jsonpatch.Patch
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateData) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateData invalid")
	}

	if len(tx.ld.ExSignatures) > 0 {
		tx.exSigners, err = tx.ld.ExSigners()
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
	if tx.data.ID == nil ||
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

	tx.dm, err = bs.LoadData(*tx.data.ID)
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
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.signers.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateData.Verify failed: no approver signing")
	}

	tx.prevDM = tx.dm.Clone()
	switch tx.dm.ModelID {
	case constants.RawModelID:
		tx.dm.Version++
		tx.dm.Data = tx.data.Data
		return nil
	case constants.CBORModelID:
		// TODO cbor patch
		if err = cbor.Valid(tx.data.Data); err != nil {
			return fmt.Errorf("TxUpdateData invalid CBOR encoding data: %v", err)
		}
		tx.dm.Version++
		tx.dm.Data = tx.data.Data
		return nil
	case constants.JSONModelID:
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

	if tx.dm.Data, err = mm.Model().ApplyPatch(tx.dm.Data, tx.data.Data); err != nil {
		return fmt.Errorf("TxUpdateData apply patch error: %v", err)
	}
	tx.dm.Version++
	if blk.ctx.Chain().IsNameService(tx.dm.ModelID) {
		n1, err := service.GetName(tx.prevDM.Data)
		if err != nil {
			return err
		}
		n2, err := service.GetName(tx.dm.Data)
		if err != nil {
			return err
		}
		if n1 != n2 {
			return fmt.Errorf("TxUpdateData should not update name, expected %s, got %s",
				strconv.Quote(n1), strconv.Quote(n2))
		}
	}

	return nil
}

func (tx *TxUpdateData) Accept(blk *Block, bs BlockState) error {
	var err error

	if err = bs.SavePrevData(*tx.data.ID, tx.prevDM); err != nil {
		return err
	}
	if err = bs.SaveData(*tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}

func (tx *TxUpdateData) Event(ts int64) *Event {
	e := NewEvent(ids.ShortID(*tx.data.ID), SrcData, ActionUpdate)
	e.Time = ts
	return nil
}
