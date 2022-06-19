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
	errPrefix := "TxUpdateData.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf("%s invalid data id", errPrefix)

	case tx.input.Version == 0:
		return fmt.Errorf("%s invalid data version", errPrefix)

	case tx.input.Threshold != nil:
		return fmt.Errorf("%s invalid threshold, should be nil", errPrefix)

	case tx.input.Keepers != nil:
		return fmt.Errorf("%s invalid keepers, should be nil", errPrefix)

	case tx.input.Approver != nil:
		return fmt.Errorf("%s invalid approver, should be nil", errPrefix)

	case tx.input.ApproveList != nil:
		return fmt.Errorf("%s invalid approveList, should be nil", errPrefix)

	case len(tx.input.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)

	case tx.input.KSig == nil:
		return fmt.Errorf("%s nil kSig", errPrefix)
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)

		case tx.ld.Amount != nil:
			return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

		case tx.ld.ExSignatures != nil:
			return fmt.Errorf("%s invalid exSignatures, should be nil", errPrefix)

		case tx.input.MSig != nil:
			return fmt.Errorf("%s invalid mSig, should be nil", errPrefix)
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return fmt.Errorf("%s invalid to, expected %s, got %s",
				errPrefix, tx.input.To, tx.ld.To)

		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return fmt.Errorf("%s nil amount", errPrefix)

		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return fmt.Errorf("%s invalid amount, expected %s, got %s",
				errPrefix, tx.input.Amount, tx.ld.Amount)

		case tx.input.MSig == nil:
			return fmt.Errorf("%s nil mSig", errPrefix)

		case tx.input.Expire < tx.ld.Timestamp:
			return fmt.Errorf("%s data expired", errPrefix)
		}

		if tx.exSigners, err = tx.ld.ExSigners(); err != nil {
			return fmt.Errorf("%s invalid exSignatures: %v", errPrefix, err)
		}
	}
	return nil
}

func (tx *TxUpdateData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxUpdateData.Verify failed:"
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	tx.di, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return fmt.Errorf("%s %v", errPrefix, err)

	case tx.di.Version != tx.input.Version:
		return fmt.Errorf("%s invalid version, expected %d, got %d",
			errPrefix, tx.di.Version, tx.input.Version)

	case !util.SatisfySigning(tx.di.Threshold, tx.di.Keepers, tx.signers, false):
		return fmt.Errorf("%s invalid signatures for data keepers", errPrefix)

	case tx.ld.NeedApprove(tx.di.Approver, tx.di.ApproveList) && !tx.signers.Has(*tx.di.Approver):
		return fmt.Errorf("%s invalid signature for data approver", errPrefix)
	}

	tx.prevDI = tx.di.Clone()
	switch tx.di.ModelID {
	case constants.RawModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		tx.di.Data = tx.input.Data

	case constants.CBORModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		var patch cborpatch.Patch
		if patch, err = cborpatch.NewPatch(tx.input.Data); err != nil {
			return fmt.Errorf("%s invalid CBOR patch, %v", errPrefix, err)
		}

		if tx.di.Data, err = patch.Apply(tx.di.Data); err != nil {
			return fmt.Errorf("%s apply patch failed, %v", errPrefix, err)
		}

	case constants.JSONModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		var patch jsonpatch.Patch
		if patch, err = jsonpatch.NewPatch(tx.input.Data); err != nil {
			return fmt.Errorf("%s invalid JSON patch, %v", errPrefix, err)
		}
		if tx.di.Data, err = patch.Apply(tx.di.Data); err != nil {
			return fmt.Errorf("%s apply patch failed, %v", errPrefix, err)
		}

	default:
		mi, err := bs.LoadModel(tx.di.ModelID)
		if err != nil {
			return fmt.Errorf("%s load model error, %v", errPrefix, err)
		}

		if tx.di.Data, err = mi.Model().ApplyPatch(tx.di.Data, tx.input.Data); err != nil {
			return fmt.Errorf("%s apply patch error, %v", errPrefix, err)
		}

		switch {
		case mi.Threshold == 0:
			if tx.input.To != nil {
				return fmt.Errorf("%s invalid to, should be nil", errPrefix)
			}

		case mi.Threshold > 0:
			if tx.input.To == nil {
				return fmt.Errorf("%s nil to", errPrefix)
			}
			if err = tx.di.VerifySig(mi.Keepers, *tx.input.MSig); err != nil {
				return fmt.Errorf("%s invalid mSig for model keepers, %v", errPrefix, err)
			}
			if !util.SatisfySigning(mi.Threshold, mi.Keepers, tx.exSigners, true) {
				return fmt.Errorf("%s invalid exSignature for model keepers", errPrefix)
			}
			tx.di.MSig = tx.input.MSig
		}
	}

	if err = tx.di.VerifySig(tx.di.Keepers, *tx.input.KSig); err != nil {
		return fmt.Errorf("%s invalid data signature for data keepers, %v", errPrefix, err)
	}
	tx.di.KSig = *tx.input.KSig
	tx.di.Version++
	if err = tx.di.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	if bctx.Chain().IsNameService(tx.di.ModelID) {
		var n1, n2 string
		if n1, err = service.GetName(tx.prevDI.Data); err != nil {
			return fmt.Errorf("%s invalid NameService data, %v", errPrefix, err)
		}
		if n2, err = service.GetName(tx.di.Data); err != nil {
			return fmt.Errorf("%s invalid NameService data, %v", errPrefix, err)
		}
		if n1 != n2 {
			return fmt.Errorf("%s can't update name, expected %s, got %s",
				errPrefix, strconv.Quote(n1), strconv.Quote(n2))
		}
	}
	return nil
}

func (tx *TxUpdateData) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if err = bs.SavePrevData(*tx.input.ID, tx.prevDI); err != nil {
		return err
	}
	if err = bs.SaveData(*tx.input.ID, tx.di); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
