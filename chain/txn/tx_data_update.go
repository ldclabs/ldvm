// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxUpdateData struct {
	TxBase
	input  *ld.TxUpdater
	di     *ld.DataInfo
	prevDI *ld.DataInfo
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxUpdateData.MarshalJSON: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxUpdateData) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateData.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == ids.EmptyDataID:
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
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.Tx.To != nil:
			return errp.Errorf("invalid to, should be nil")

		case tx.ld.Tx.Amount != nil:
			return errp.Errorf("invalid amount, should be nil")

		case tx.ld.ExSignatures != nil:
			return errp.Errorf("invalid exSignatures, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.Tx.To == nil || *tx.input.To != *tx.ld.Tx.To:
			return errp.Errorf("invalid to, expected %s, got %s",
				tx.input.To, tx.ld.Tx.To)

		case tx.input.Amount == nil || tx.ld.Tx.Amount == nil:
			return errp.Errorf("nil amount")

		case tx.input.Amount.Cmp(tx.ld.Tx.Amount) != 0:
			return errp.Errorf("invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Tx.Amount)

		case tx.input.Expire < tx.ld.Timestamp:
			return errp.Errorf("data expired")

		case len(tx.ld.ExSignatures) == 0:
			return errp.Errorf("no exSignatures")
		}
	}
	return nil
}

func (tx *TxUpdateData) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateData.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.di, err = cs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case tx.di.Version != tx.input.Version:
		return errp.Errorf("invalid version, expected %d, got %d",
			tx.di.Version, tx.input.Version)

	case tx.di.SigClaims != nil && tx.input.SigClaims == nil:
		return errp.Errorf("invalid sigClaims, should not be nil")

	case !tx.di.Verify(tx.ld.TxHash(), tx.ld.Signatures):
		return errp.Errorf("invalid signatures for data keepers")

	case !tx.ld.IsApproved(tx.di.Approver, tx.di.ApproveList, false):
		return errp.Errorf("invalid signature for data approver")
	}

	tx.prevDI = tx.di.Clone()
	switch tx.di.ModelID {
	case ld.RawModelID, ld.CBORModelID, ld.JSONModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		if tx.di.Payload, err = tx.di.Patch(tx.input.Data); err != nil {
			return errp.ErrorIf(err)
		}

	default:
		mi, err := cs.LoadModel(tx.di.ModelID)
		if err != nil {
			return errp.Errorf("load model error, %v", err)
		}

		if tx.di.Payload, err = mi.Model().ApplyPatch(tx.di.Payload, tx.input.Data); err != nil {
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
			if !mi.Verify(tx.ld.ExHash(), tx.ld.ExSignatures) {
				return errp.Errorf("invalid exSignature for model keepers")
			}
		}
	}

	tx.di.Version++
	if tx.input.SigClaims != nil {
		tx.di.SigClaims = tx.input.SigClaims
		tx.di.Sig = tx.input.Sig
	}

	if err = tx.di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.di.ValidSigClaims(); err != nil {
		return errp.ErrorIf(err)
	}

	if ctx.ChainConfig().IsNameService(tx.di.ModelID) {
		var n1, n2 string
		if n1, err = service.GetName(tx.prevDI.Payload); err != nil {
			return errp.Errorf("invalid NameService data, %v", err)
		}
		if n2, err = service.GetName(tx.di.Payload); err != nil {
			return errp.Errorf("invalid NameService data, %v", err)
		}
		if n1 != n2 {
			return errp.Errorf("can't update name, expected %q, got %q", n1, n2)
		}
	}

	if err = cs.SavePrevData(tx.prevDI); err != nil {
		return errp.ErrorIf(err)
	}
	if err = cs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
