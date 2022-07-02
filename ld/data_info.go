// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"

	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

type DataInfo struct {
	ModelID util.ModelID `cbor:"m" json:"mid"` // model id
	// data version，the initial value is 1, should increase 1 when updating,
	// 0 indicates that the data is invalid, for example, deleted or punished.
	Version uint64 `cbor:"v" json:"version"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can update the data.
	// the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who owned this data, no more than 1024
	Keepers     util.EthIDs  `cbor:"kp" json:"keepers"`
	Approver    *util.EthID  `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes      `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Data        util.RawData `cbor:"d" json:"data"`
	// full data signature signing by Data Keeper
	KSig util.Signature `cbor:"ks" json:"kSig"`
	// full data signature signing by ModelService Authority
	MSig *util.Signature `cbor:"ms,omitempty" json:"mSig,omitempty"`

	// external assignment fields
	ID  util.DataID `cbor:"-" json:"id"`
	raw []byte      `cbor:"-" json:"-"`
}

func (t *DataInfo) Clone() *DataInfo {
	x := new(DataInfo)
	*x = *t
	x.Keepers = make(util.EthIDs, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	if t.Approver != nil {
		id := *t.Approver
		x.Approver = &id
	}
	if t.ApproveList != nil {
		x.ApproveList = make(TxTypes, len(t.ApproveList))
		copy(x.ApproveList, t.ApproveList)
	}
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	if t.MSig != nil {
		mSig := *t.MSig
		x.MSig = &mSig
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *DataInfo is well-formed.
func (t *DataInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("DataInfo.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case len(t.Keepers) > MaxKeepers:
		return errp.Errorf("too many keepers")

	case int(t.Threshold) > len(t.Keepers):
		return errp.Errorf("invalid threshold")

	case t.Approver != nil && *t.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver")
	}

	if err = t.Keepers.CheckDuplicate(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if err = t.Keepers.CheckEmptyID(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
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

	if t.KSig != util.SignatureEmpty && len(t.Keepers) > 0 {
		if err = t.VerifySig(t.Keepers, t.KSig); err != nil {
			return errp.Errorf("invalid kSig, %v", err)
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *DataInfo) VerifySig(signers util.EthIDs, sig util.Signature) error {
	signer, err := util.DeriveSigner(t.Data, sig[:])
	switch {
	case err != nil:
		return err
	case !signers.Has(signer):
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func (t *DataInfo) MarkDeleted(data []byte) error {
	t.Version = 0
	t.KSig = util.SignatureEmpty
	t.MSig = nil
	if data != nil {
		t.Data = data
	}
	return t.SyntacticVerify()
}

type patcher interface {
	Apply(doc []byte) ([]byte, error)
}

// Patch applies a patch to the data, returns the patched data.
// It will not change the data.
func (t *DataInfo) Patch(operations []byte) ([]byte, error) {
	var err error
	var p patcher
	errp := util.ErrPrefix("DataInfo.Patch error: ")

	switch t.ModelID {
	case constants.RawModelID:
		return operations, nil

	case constants.CBORModelID:
		p, err = cborpatch.NewPatch(operations)
		if err != nil {
			return nil, errp.Errorf("invalid CBOR patch, %v", err)
		}

	case constants.JSONModelID:
		p, err = jsonpatch.NewPatch(operations)
		if err != nil {
			return nil, errp.Errorf("invalid JSON patch, %v", err)
		}

	default:
		return nil, errp.Errorf("unsupport mid %s", t.ModelID)
	}

	return errp.ErrorMap(p.Apply(t.Data))
}

func (t *DataInfo) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *DataInfo) Unmarshal(data []byte) error {
	return util.ErrPrefix("DataInfo.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *DataInfo) Marshal() ([]byte, error) {
	return util.ErrPrefix("DataInfo.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}
