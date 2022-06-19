// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateData struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxUpdater
	di        *ld.DataInfo
	prevDI    *ld.DataInfo
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateData.MarshalJSON error: invalid tx.input")
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
	errp := util.ErrPrefix("TxUpdateData.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return errp.Errorf("invalid data id")

	case tx.input.Version == 0:
		return errp.Errorf("invalid data version")

	case tx.input.Threshold != nil:
		return errp.Errorf("invalid threshold, should be nil")

	case tx.input.Keepers != nil:
		return errp.Errorf("invalid keepers, should be nil")

	case tx.input.Approver != nil:
		return errp.Errorf("invalid approver, should be nil")

	case tx.input.ApproveList != nil:
		return errp.Errorf("invalid approveList, should be nil")

	case len(tx.input.Data) == 0:
		return errp.Errorf("invalid data")

	case tx.input.KSig == nil:
		return errp.Errorf("nil kSig")
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return errp.Errorf("invalid to, should be nil")

		case tx.ld.Amount != nil:
			return errp.Errorf("invalid amount, should be nil")

		case tx.ld.ExSignatures != nil:
			return errp.Errorf("invalid exSignatures, should be nil")

		case tx.input.MSig != nil:
			return errp.Errorf("invalid mSig, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return errp.Errorf("invalid to, expected %s, got %s",
				tx.input.To, tx.ld.To)

		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return errp.Errorf("nil amount")

		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return errp.Errorf("invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Amount)

		case tx.input.MSig == nil:
			return errp.Errorf("nil mSig")

		case tx.input.Expire < tx.ld.Timestamp:
			return errp.Errorf("data expired")
		}

		if tx.exSigners, err = tx.ld.ExSigners(); err != nil {
			return errp.Errorf("invalid exSignatures, %v", err)
		}
	}
	return nil
}

func (tx *TxUpdateData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateData.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.di, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case tx.di.Version != tx.input.Version:
		return errp.Errorf("invalid version, expected %d, got %d",
			tx.di.Version, tx.input.Version)

	case !util.SatisfySigning(tx.di.Threshold, tx.di.Keepers, tx.signers, false):
		return errp.Errorf("invalid signatures for data keepers")

	case tx.ld.NeedApprove(tx.di.Approver, tx.di.ApproveList) && !tx.signers.Has(*tx.di.Approver):
		return errp.Errorf("invalid signature for data approver")
	}

	tx.prevDI = tx.di.Clone()
	switch tx.di.ModelID {
	case constants.RawModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		tx.di.Data = tx.input.Data

	case constants.CBORModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		var patch cborpatch.Patch
		if patch, err = cborpatch.NewPatch(tx.input.Data); err != nil {
			return errp.Errorf("invalid CBOR patch, %v", err)
		}

		if tx.di.Data, err = patch.Apply(tx.di.Data); err != nil {
			return errp.Errorf("apply patch failed, %v", err)
		}

	case constants.JSONModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		var patch jsonpatch.Patch
		if patch, err = jsonpatch.NewPatch(tx.input.Data); err != nil {
			return errp.Errorf("invalid JSON patch, %v", err)
		}
		if tx.di.Data, err = patch.Apply(tx.di.Data); err != nil {
			return errp.Errorf("apply patch failed, %v", err)
		}

	default:
		mi, err := bs.LoadModel(tx.di.ModelID)
		if err != nil {
			return errp.Errorf("load model error, %v", err)
		}

		if tx.di.Data, err = mi.Model().ApplyPatch(tx.di.Data, tx.input.Data); err != nil {
			return errp.Errorf("apply patch error, %v", err)
		}

		switch {
		case mi.Threshold == 0:
			if tx.input.To != nil {
				return errp.Errorf("invalid to, should be nil")
			}

		case mi.Threshold > 0:
			if tx.input.To == nil {
				return errp.Errorf("nil to")
			}
			if err = tx.di.VerifySig(mi.Keepers, *tx.input.MSig); err != nil {
				return errp.Errorf("invalid mSig for model keepers, %v", err)
			}
			if !util.SatisfySigning(mi.Threshold, mi.Keepers, tx.exSigners, true) {
				return errp.Errorf("invalid exSignature for model keepers")
			}
			tx.di.MSig = tx.input.MSig
		}
	}

	if err = tx.di.VerifySig(tx.di.Keepers, *tx.input.KSig); err != nil {
		return errp.Errorf("invalid data signature for data keepers, %v", err)
	}
	tx.di.KSig = *tx.input.KSig
	tx.di.Version++
	if err = tx.di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if bctx.Chain().IsNameService(tx.di.ModelID) {
		var n1, n2 string
		if n1, err = service.GetName(tx.prevDI.Data); err != nil {
			return errp.Errorf("invalid NameService data, %v", err)
		}
		if n2, err = service.GetName(tx.di.Data); err != nil {
			return errp.Errorf("invalid NameService data, %v", err)
		}
		if n1 != n2 {
			return errp.Errorf("can't update name, expected %s, got %s",
				strconv.Quote(n1), strconv.Quote(n2))
		}
	}
	return nil
}

func (tx *TxUpdateData) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateData.Accept error: ")

	if err = bs.SavePrevData(*tx.input.ID, tx.prevDI); err != nil {
		return errp.ErrorIf(err)
	}
	if err = bs.SaveData(*tx.input.ID, tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
