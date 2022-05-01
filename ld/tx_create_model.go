// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/ldclabs/ldvm/util"
)

var modelNameReg = regexp.MustCompile(`^[A-Z][0-9A-Za-z]{1,127}$`)

type ModelMeta struct {
	// model name, should match ^[A-Z][0-9A-Za-z]{1,127}$
	Name string
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can change the data.
	// the maximum value is len(keepers)
	Threshold uint8
	// keepers who owned this model, no more than 255
	// Creating data using this model requires keepers to sign.
	// no keepers or threshold is 0 means don't need sign.
	Keepers []ids.ShortID
	Data    []byte

	// external assignment
	st  schema.Type
	raw []byte
	ID  ids.ShortID
}

type jsonModelMeta struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Threshold uint8           `json:"threshold"`
	Keepers   []string        `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (t *ModelMeta) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonModelMeta{
		ID:        util.ModelID(t.ID).String(),
		Name:      t.Name,
		Threshold: t.Threshold,
		Data:      util.JSONMarshalData(t.Data),
		Keepers:   make([]string, len(t.Keepers)),
	}
	for i := range t.Keepers {
		v.Keepers[i] = util.EthID(t.Keepers[i]).String()
	}
	return json.Marshal(v)
}

func (t *ModelMeta) Copy() *ModelMeta {
	x := new(ModelMeta)
	*x = *t
	x.Keepers = make([]ids.ShortID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = nil
	return x
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
		if id == ids.ShortEmpty {
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

func (t *ModelMeta) Equal(o *ModelMeta) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Name != t.Name {
		return false
	}
	if o.Threshold != t.Threshold {
		return false
	}
	if len(o.Keepers) != len(t.Keepers) {
		return false
	}
	for i := range t.Keepers {
		if o.Keepers[i] != t.Keepers[i] {
			return false
		}
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *ModelMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *ModelMeta) Unmarshal(data []byte) error {
	p, err := modelMetaLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindModelMeta); ok {
		if !v.Threshold.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}
		t.Name = v.Name
		t.Threshold = v.Threshold.Value()
		t.Data = v.Data
		if t.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindModelMeta")
}

func (t *ModelMeta) Marshal() ([]byte, error) {
	v := &bindModelMeta{
		Name:      t.Name,
		Threshold: FromUint8(t.Threshold),
		Keepers:   FromShortIDs(t.Keepers),
		Data:      t.Data,
	}
	data, err := modelMetaLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindModelMeta struct {
	Name      string
	Threshold Uint8
	Keepers   [][]byte
	Data      []byte
}

var modelMetaLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type ID20 bytes
	type ModelMeta struct {
		Name      String (rename "n")
		Threshold Uint8  (rename "th")
		Keepers   [ID20] (rename "kp")
		Data      Bytes  (rename "d")
	}
`
	builder, err := NewLDBuilder("ModelMeta", []byte(sch), (*bindModelMeta)(nil))
	if err != nil {
		panic(err)
	}
	modelMetaLDBuilder = builder
}
