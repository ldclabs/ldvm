// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"regexp"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/ldclabs/ldvm/util"
)

var modelNameReg = regexp.MustCompile(`^[A-Z][0-9A-Za-z]{1,127}$`)

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
	Keepers []util.EthID `cbor:"kp" json:"keepers"`
	Data    RawData      `cbor:"d" json:"data"`

	// external assignment
	ID  util.ModelID `cbor:"-" json:"id"`
	st  schema.Type  `cbor:"-" json:"-"`
	raw []byte       `cbor:"-" json:"-"`
}

func (t *ModelMeta) SchemaType() schema.Type {
	return t.st
}

// SyntacticVerify verifies that a *ModelMeta is well-formed.
func (t *ModelMeta) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid ModelMeta")
	}

	if !modelNameReg.MatchString(t.Name) {
		return fmt.Errorf("invalid name")
	}
	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("invalid model keeper")
		}
	}
	if len(t.Data) < 10 {
		return fmt.Errorf("invalid data, bytes should >= %d", 10)
	}

	var err error
	if t.st, err = NewSchemaType(t.Name, t.Data); err != nil {
		return fmt.Errorf("parse ipld schema error: %v", err)
	}
	if _, err = t.Marshal(); err != nil {
		return fmt.Errorf("ModelMeta marshal error: %v", err)
	}
	return nil
}
func (t *ModelMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		MustMarshal(t)
	}
	return t.raw
}

func (t *ModelMeta) Unmarshal(data []byte) error {
	t.raw = data
	return DecMode.Unmarshal(data, t)
}

func (t *ModelMeta) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}
