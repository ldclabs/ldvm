// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

// TxUpdater is a hybrid data model for:
//
// TxCreateData{ModelID, Version, Threshold, Keepers, Data, KSig[, Approver, ApproveList]} no model keepers
// TxCreateData{ModelID, Version, To, Amount, Threshold, Keepers, Data, KSig, MSig, Expire[, Approver, ApproveList]} with model keepers
// TxUpdateData{ID, Version, Data, KSig} no model keepers
// TxUpdateData{ID, Version, To, Amount, Data, KSig, MSig, Expire} with model keepers
// TxDeleteData{ID, Version[, Data]}
// TxUpdateDataKeepers{ID, Version, Threshold, Keepers, KSig[, Approver, ApproveList, Data]}
// TxUpdateDataKeepersByAuth{ID, Version, To, Amount, Threshold, Keepers, KSig, Expire[, Approver, ApproveList, Token, Data]}
// TxUpdateModelKeepers{ModelID, Threshold, Keepers[, Approver, Data]}
type TxUpdater struct {
	ID          *util.DataID      `cbor:"id,omitempty" json:"id,omitempty"`     // data id
	ModelID     *util.ModelID     `cbor:"mid,omitempty" json:"mid,omitempty"`   // model id
	Version     uint64            `cbor:"v,omitempty" json:"version,omitempty"` // data version
	Threshold   *uint16           `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers     *util.EthIDs      `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Approver    *util.EthID       `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes           `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Token       *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"` // token symbol, default is NativeToken
	To          *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`    // optional recipient
	Amount      *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"` // transfer amount
	KSig        *util.Signature   `cbor:"ks,omitempty" json:"kSig,omitempty"`  // full data signature signing by Data Keeper
	MSig        *util.Signature   `cbor:"ms,omitempty" json:"mSig,omitempty"`  // full data signature signing by Model Service Authority
	Expire      uint64            `cbor:"e,omitempty" json:"expire,omitempty"`
	Data        util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxUpdater is well-formed.
func (t *TxUpdater) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdater.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.Token != nil && !t.Token.Valid():
		return errp.Errorf("invalid token symbol %s", strconv.Quote(t.Token.GoString()))

	case t.Amount != nil && t.Amount.Sign() < 0:
		return errp.Errorf("invalid amount")
	}

	if t.Keepers != nil || t.Threshold != nil {
		switch {
		case t.Threshold == nil:
			return errp.Errorf("nil threshold together with keepers")

		case t.Keepers == nil:
			return errp.Errorf("nil keepers together with threshold")

		case int(*t.Threshold) > len(*t.Keepers):
			return errp.Errorf("invalid threshold, expected <= %d, got %d",
				len(*t.Keepers), *t.Threshold)

		case len(*t.Keepers) > MaxKeepers:
			return errp.Errorf("invalid keepers, expected <= %d, got %d",
				MaxKeepers, len(*t.Keepers))
		}

		if err = t.Keepers.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid keepers, %v", err)
		}

		if err = t.Keepers.CheckEmptyID(); err != nil {
			return errp.Errorf("invalid keepers, %v", err)
		}
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid approveList, %v", err)
		}

		for _, ty := range t.ApproveList {
			if !DataTxTypes.Has(ty) {
				return errp.Errorf("invalid TxType %s in approveList", ty)
			}
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
	return util.ErrPrefix("TxUpdater.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxUpdater) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxUpdater.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}
