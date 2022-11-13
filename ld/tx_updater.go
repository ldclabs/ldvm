// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

// TxUpdater is a hybrid data model for:
//
// TxCreateData{ModelID, Version, Threshold, Keepers, Data[, Approver, ApproveList]} no model keepers
// TxCreateData{ModelID, Version, To, Amount, Threshold, Keepers, Data, Expire[, Approver, ApproveList]} with model keepers
//
// TxUpdateData{ID, Version, Data} no model keepers
// TxUpdateData{ID, Version, SigClaims, Sig, Data} no model keepers
// TxUpdateData{ID, Version, To, Amount, Data, Expire} with model keepers
// TxUpdateData{ID, Version, SigClaims, Sig, To, Amount, Data, Expire} with model keepers
// TxUpgradeData{ID, Version, To, Amount, Data, Expire} with model keepers
// TxUpgradeData{ID, Version, SigClaims, Sig, To, Amount, Data, Expire} with model keepers
//
// TxDeleteData{ID, Version[, Data]}
//
// TxUpdateDataInfo{ID, Version, Threshold, Keepers[, SigClaims, Sig, Approver, ApproveList]}
// TxUpdateDataInfoByAuth{ID, Version, To, Amount, Threshold, Keepers, Expire[, Approver, ApproveList, Token]}
//
// TxUpdateModelInfo{ModelID, Threshold, Keepers[, Approver]}
type TxUpdater struct {
	ID          *ids.DataID      `cbor:"id,omitempty" json:"id,omitempty"`     // data id
	ModelID     *ids.ModelID     `cbor:"mid,omitempty" json:"mid,omitempty"`   // model id
	Version     uint64           `cbor:"v,omitempty" json:"version,omitempty"` // data version
	Threshold   *uint16          `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers     *signer.Keys     `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Approver    *signer.Key      `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList *TxTypes         `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Token       *ids.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"` // token symbol, default is NativeToken
	To          *ids.Address     `cbor:"to,omitempty" json:"to,omitempty"`    // optional recipient
	Amount      *big.Int         `cbor:"a,omitempty" json:"amount,omitempty"` // transfer amount
	SigClaims   *SigClaims       `cbor:"sc,omitempty" json:"sigClaims,omitempty"`
	Sig         *signer.Sig      `cbor:"s,omitempty" json:"sig,omitempty"`
	Expire      uint64           `cbor:"e,omitempty" json:"expire,omitempty"`
	Data        encoding.RawData `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxUpdater is well-formed.
func (t *TxUpdater) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("ld.TxUpdater.SyntacticVerify: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.Token != nil && !t.Token.Valid():
		return errp.Errorf("invalid token symbol %q", t.Token.GoString())

	case t.Amount != nil && t.Amount.Sign() < 0:
		return errp.Errorf("invalid amount")

	case t.Keepers == nil && t.Threshold != nil:
		return errp.Errorf("no keepers, threshold should be nil")

	case t.Keepers != nil && t.Threshold == nil:
		return errp.Errorf("invalid threshold")

	case t.SigClaims == nil && t.Sig != nil:
		return errp.Errorf("no sigClaims, typed signature should be nil")

	case t.SigClaims != nil && t.Sig == nil:
		return errp.Errorf("invalid typed signature")
	}

	if t.Keepers != nil {
		switch {
		case int(*t.Threshold) > len(*t.Keepers):
			return errp.Errorf("invalid threshold, expected <= %d, got %d",
				len(*t.Keepers), *t.Threshold)

		case len(*t.Keepers) > MaxKeepers:
			return errp.Errorf("invalid keepers, expected <= %d, got %d",
				MaxKeepers, len(*t.Keepers))
		}

		if err = t.Keepers.Valid(); err != nil {
			return errp.Errorf("invalid keepers, %v", err)
		}
	}

	if t.Approver != nil {
		if err = t.Approver.ValidOrEmpty(); err != nil {
			return errp.Errorf("invalid approver, %v", err)
		}
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid approveList, %v", err)
		}

		for _, ty := range *t.ApproveList {
			if !DataTxTypes.Has(ty) {
				return errp.Errorf("invalid TxType %s in approveList", ty)
			}
		}
	}

	if t.SigClaims != nil {
		if err = t.SigClaims.SyntacticVerify(); err != nil {
			return errp.Errorf("invalid sigClaims, %v", err)
		}
		if err = t.Sig.Valid(); err != nil {
			return errp.Errorf("invalid sig, %v", err)
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxUpdater) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxUpdater) Unmarshal(data []byte) error {
	return erring.ErrPrefix("ld.TxUpdater.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, t))
}

func (t *TxUpdater) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.TxUpdater.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(t))
}
