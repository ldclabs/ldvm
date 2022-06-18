// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"

	"github.com/ldclabs/ldvm/util"
)

type DataInfo struct {
	ModelID util.ModelID `cbor:"mid" json:"mid"` // model id
	// data version，the initial value is 1, should increase 1 when updating,
	// 0 indicates that the data is invalid, for example, deleted or punished.
	Version uint64 `cbor:"v" json:"version"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can update the data.
	// the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who owned this data, no more than 1024
	Keepers     util.EthIDs `cbor:"kp" json:"keepers"`
	Approver    *util.EthID `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes     `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Data        RawData     `cbor:"d" json:"data"`
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
	errPrefix := "DataInfo.SyntacticVerify failed:"

	switch {
	case t == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case len(t.Keepers) > MaxKeepers:
		return fmt.Errorf("%s too many keepers", errPrefix)

	case int(t.Threshold) > len(t.Keepers):
		return fmt.Errorf("%s invalid threshold", errPrefix)

	case t.Approver != nil && *t.Approver == util.EthIDEmpty:
		return fmt.Errorf("%s invalid approver", errPrefix)
	}

	if err = t.Keepers.CheckDuplicate(); err != nil {
		return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
	}

	if err = t.Keepers.CheckEmptyID(); err != nil {
		return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return fmt.Errorf("%s invalid approveList, %v", errPrefix, err)
		}

		for _, ty := range t.ApproveList {
			if !DataTxTypes.Has(ty) {
				return fmt.Errorf("%s invalid TxType %s in approveList", errPrefix, ty)
			}
		}
	}

	if t.KSig != util.SignatureEmpty && len(t.Keepers) > 0 {
		kSigner, err := util.DeriveSigner(t.Data, t.KSig[:])
		if err != nil {
			return fmt.Errorf("%s %v", errPrefix, err)
		}
		if !t.Keepers.Has(kSigner) {
			return fmt.Errorf("%s invalid kSig", errPrefix)
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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

func (t *DataInfo) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *DataInfo) Unmarshal(data []byte) error {
	return UnmarshalCBOR(data, t)
}

func (t *DataInfo) Marshal() ([]byte, error) {
	return MarshalCBOR(t)
}
