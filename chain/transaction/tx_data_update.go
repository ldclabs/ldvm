// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	input     *ld.TxUpdater
	dm        *ld.DataMeta
	prevDM    *ld.DataMeta
	jsonPatch jsonpatch.Patch
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateData.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
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

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid to, should be nil")
	case tx.ld.Token != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid token, should be nil")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf(
			"TxUpdateData.SyntacticVerify failed: invalid data id")
	case tx.input.Version == 0:
		return fmt.Errorf(
			"TxUpdateData.SyntacticVerify failed: invalid data version")
	case tx.input.Threshold != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid threshold, should be nil")
	case tx.input.Keepers != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid keepers, should be nil")
	case tx.input.Approver != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid approver, should be nil")
	case tx.input.ApproveList != nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid approveList, should be nil")
	case len(tx.input.Data) == 0:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid data")
	case tx.input.KSig == nil:
		return fmt.Errorf("TxUpdateData.SyntacticVerify failed: nil kSig")
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid to, should be nil")
		case tx.ld.Amount != nil:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid amount, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid to, expected %s, got %s",
				tx.input.To, tx.ld.To)
		case tx.input.Expire < tx.ld.Timestamp:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: data expired")
		case tx.input.MSig == nil:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: nil mSig")
		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: nil amount")
		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Amount)
		}

		tx.exSigners, err = tx.ld.ExSigners()
		if err != nil {
			return fmt.Errorf("TxUpdateData.SyntacticVerify failed: invalid exSignatures: %v", err)
		}
	}
	return nil
}

func (tx *TxUpdateData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateData load data failed: %v", err)
	}
	if tx.dm.Version != tx.input.Version {
		return fmt.Errorf("TxUpdateData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.input.Version)
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
		tx.dm.Data = tx.input.Data
		return nil
	case constants.CBORModelID:
		// TODO cbor patch
		if err = cbor.Valid(tx.input.Data); err != nil {
			return fmt.Errorf("TxUpdateData invalid CBOR encoding data: %v", err)
		}
		tx.dm.Version++
		tx.dm.Data = tx.input.Data
		return nil
	case constants.JSONModelID:
		tx.jsonPatch, err = jsonpatch.DecodePatch(tx.input.Data)
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

	if tx.dm.Data, err = mm.Model().ApplyPatch(tx.dm.Data, tx.input.Data); err != nil {
		return fmt.Errorf("TxUpdateData apply patch error: %v", err)
	}
	tx.dm.Version++
	if bctx.Chain().IsNameService(tx.dm.ModelID) {
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

func (tx *TxUpdateData) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if err = bs.SavePrevData(*tx.input.ID, tx.prevDM); err != nil {
		return err
	}
	if err = bs.SaveData(*tx.input.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
