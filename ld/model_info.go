// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"regexp"
	"unicode/utf8"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
)

var ModelNameReg = regexp.MustCompile(`^[A-Z][0-9A-Za-z]{1,127}$`)

type ModelInfo struct {
	// model name, should match ^[A-Z][0-9A-Za-z]{1,127}$
	Name string `cbor:"n" json:"name"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means any one using the model don't need to approve.
	// the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who owned this model, no more than 64
	// Creating data using this model requires keepers to sign.
	// no keepers or threshold is 0 means don't need sign.
	Keepers  signer.Keys `cbor:"kp" json:"keepers"`
	Approver signer.Key  `cbor:"ap" json:"approver,omitempty"`
	Schema   string      `cbor:"sc" json:"schema"`

	// external assignment fields
	ID    util.ModelID `cbor:"-" json:"id"`
	model *IPLDModel   `cbor:"-" json:"-"`
	raw   []byte       `cbor:"-" json:"-"`
}

func (t *ModelInfo) Model() *IPLDModel {
	return t.model
}

// SyntacticVerify verifies that a *ModelInfo is well-formed.
func (t *ModelInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("ld.ModelInfo.SyntacticVerify: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case !ModelNameReg.MatchString(t.Name):
		return errp.Errorf("invalid name")

	case len(t.Keepers) > MaxKeepers:
		return errp.Errorf("too many keepers")

	case int(t.Threshold) > len(t.Keepers):
		return errp.Errorf("invalid threshold")

	case len(t.Schema) < 10 || !utf8.ValidString(t.Schema):
		return errp.Errorf("invalid schema string")
	}

	if err = t.Keepers.Valid(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if t.Approver != nil {
		if err = t.Approver.Valid(); err != nil {
			return errp.Errorf("invalid approver, %v", err)
		}
	}

	if t.model, err = NewIPLDModel(t.Name, t.Schema); err != nil {
		return errp.ErrorIf(err)
	}
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *ModelInfo) Verify(digestHash []byte, sigs signer.Sigs) bool {
	if t.Threshold == 0 {
		return true
	}

	return t.Keepers.Verify(digestHash, sigs, t.Threshold)
}

func (t *ModelInfo) VerifyPlus(digestHash []byte, sigs signer.Sigs) bool {
	return t.Keepers.VerifyPlus(digestHash, sigs, t.Threshold)
}

func (t *ModelInfo) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *ModelInfo) Unmarshal(data []byte) error {
	return util.ErrPrefix("ld.ModelInfo.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *ModelInfo) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.ModelInfo.Marshal: ").
		ErrorMap(util.MarshalCBOR(t))
}
