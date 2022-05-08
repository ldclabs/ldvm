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
	// The minimum value is 0, means no one can change the data.
	// the maximum value is len(keepers)
	Threshold uint8 `cbor:"th" json:"threshold"`
	// keepers who owned this data, no more than 255
	Keepers  []util.EthID    `cbor:"kp" json:"keepers"`
	Approver *util.EthID     `cbor:"ap" json:"approver,omitempty"`
	KSig     util.Signature  `cbor:"ks" json:"kSig"`                     // full data signature signing by Data Keeper
	MSig     *util.Signature `cbor:"ms,omitempty" json:"mSig,omitempty"` // full data signature signing by Service Authority
	Data     RawData         `cbor:"d" json:"data"`

	// external assignment
	ID  util.DataID `cbor:"-" json:"id"`
	raw []byte      `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *DataMeta is well-formed.
func (t *DataMeta) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid DataMeta")
	}

	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("invalid keeper")
		}
	}
	if t.Approver != nil && *t.Approver == util.EthIDEmpty {
		return fmt.Errorf("invalid approver")
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("DataMeta marshal error: %v", err)
	}
	return nil
}

func (t *DataMeta) Copy() *DataMeta {
	x := new(DataMeta)
	*x = *t
	x.Keepers = make([]util.EthID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	if t.MSig != nil {
		mSig := util.Signature{}
		copy(mSig[:], (*t.MSig)[:])
		x.MSig = &mSig
	}
	x.raw = nil
	return x
}

func (t *DataMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		MustMarshal(t)
	}
	return t.raw
}

func (t *DataMeta) Unmarshal(data []byte) error {
	t.raw = data
	return DecMode.Unmarshal(data, t)
}

func (t *DataMeta) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}
