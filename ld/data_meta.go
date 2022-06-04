// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"

	"github.com/ldclabs/ldvm/util"
)

type DataMeta struct {
	ModelID util.ModelID `cbor:"mid" json:"mid"` // model id
	// data version，the initial value is 1, should increase 1 when updating,
	// 0 indicates that the data is invalid, for example, deleted or punished.
	Version uint64 `cbor:"v" json:"version"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can update the data.
	// the maximum value is len(keepers)
	Threshold uint8 `cbor:"th" json:"threshold"`
	// keepers who owned this data, no more than 255
	Keepers     util.EthIDs `cbor:"kp" json:"keepers"`
	Approver    *util.EthID `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList []TxType    `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Data        RawData     `cbor:"d" json:"data"`
	// full data signature signing by Data Keeper
	KSig util.Signature `cbor:"ks" json:"kSig"`
	// full data signature signing by ModelService Authority
	MSig *util.Signature `cbor:"ms,omitempty" json:"mSig,omitempty"`

	// external assignment fields
	ID  util.DataID `cbor:"-" json:"id"`
	raw []byte      `cbor:"-" json:"-"`
}

func (t *DataMeta) Clone() *DataMeta {
	x := new(DataMeta)
	*x = *t
	x.Keepers = make([]util.EthID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	if t.Approver != nil {
		id := *t.Approver
		x.Approver = &id
	}
	if t.ApproveList != nil {
		x.ApproveList = make([]TxType, len(t.ApproveList))
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

// SyntacticVerify verifies that a *DataMeta is well-formed.
func (t *DataMeta) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("DataMeta.SyntacticVerify failed: nil pointer")
	}

	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("DataMeta.SyntacticVerify failed: too many keepers")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("DataMeta.SyntacticVerify failed: invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("DataMeta.SyntacticVerify failed: invalid keeper")
		}
	}
	if t.Approver != nil && *t.Approver == util.EthIDEmpty {
		return fmt.Errorf("DataMeta.SyntacticVerify failed: invalid approver")
	}
	if t.ApproveList != nil {
		for _, ty := range t.ApproveList {
			if ty > TypeDeleteData || ty < TypeCreateData {
				return fmt.Errorf("DataMeta.SyntacticVerify failed: invalid TxType %d in approveList", ty)
			}
		}
	}

	if t.KSig != util.SignatureEmpty {
		kSigner, err := util.DeriveSigner(t.Data, t.KSig[:])
		if err != nil {
			return fmt.Errorf("DataMeta.SyntacticVerify failed: %v", err)
		}
		if !t.Keepers.Has(kSigner) {
			return fmt.Errorf("DataMeta.SyntacticVerify failed: invalid kSig")
		}
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("DataMeta.SyntacticVerify marshal error: %v", err)
	}
	return nil
}

func (t *DataMeta) MarkDeleted(data []byte) error {
	t.Version = 0
	t.KSig = util.SignatureEmpty
	t.MSig = nil
	if data != nil {
		t.Data = data
	}
	return t.SyntacticVerify()
}

func (t *DataMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *DataMeta) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *DataMeta) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
