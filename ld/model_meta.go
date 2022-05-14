// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"regexp"

	"github.com/ldclabs/ldvm/util"
)

var ModelNameReg = regexp.MustCompile(`^[A-Z][0-9A-Za-z]{1,127}$`)

type ModelMeta struct {
	// model name, should match ^[A-Z][0-9A-Za-z]{1,127}$
	Name string `cbor:"n" json:"name"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can change the data.
	// the maximum value is len(keepers)
	Threshold uint8 `cbor:"th" json:"threshold"`
	// keepers who owned this model, no more than 255
	// Creating data using this model requires keepers to sign.
	// no keepers or threshold is 0 means don't need sign.
	Keepers  util.EthIDs `cbor:"kp" json:"keepers"`
	Approver *util.EthID `cbor:"ap,omitempty" json:"approver,omitempty"`
	Data     RawData     `cbor:"d" json:"data"`

	// external assignment fields
	ID    util.ModelID `cbor:"-" json:"id"`
	model *IPLDModel   `cbor:"-" json:"-"`
	raw   []byte       `cbor:"-" json:"-"`
}

func (t *ModelMeta) Model() *IPLDModel {
	return t.model
}

// SyntacticVerify verifies that a *ModelMeta is well-formed.
func (t *ModelMeta) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: nil pointer")
	}

	if !ModelNameReg.MatchString(t.Name) {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: invalid name")
	}
	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: too many keepers")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("ModelMeta.SyntacticVerify failed: invalid keeper")
		}
	}
	if t.Approver != nil && *t.Approver == util.EthIDEmpty {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: invalid approver")
	}
	if len(t.Data) < 10 {
		return fmt.Errorf("ModelMeta.SyntacticVerify failed: invalid data bytes")
	}

	var err error
	if t.model, err = NewIPLDModel(t.Name, t.Data); err != nil {
		return fmt.Errorf("ModelMeta.SyntacticVerify error: %v", err)
	}
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("ModelMeta.SyntacticVerify marshal error: %v", err)
	}
	return nil
}

func (t *ModelMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *ModelMeta) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *ModelMeta) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
