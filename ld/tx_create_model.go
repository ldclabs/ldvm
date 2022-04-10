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

func (m *ModelMeta) MarshalJSON() ([]byte, error) {
	if m == nil {
		return Null, nil
	}
	v := &jsonModelMeta{
		ID:        ModelID(m.ID).String(),
		Name:      m.Name,
		Threshold: m.Threshold,
		Data:      JsonMarshalData(m.Data),
		Keepers:   make([]string, len(m.Keepers)),
	}
	for i := range m.Keepers {
		v.Keepers[i] = EthID(m.Keepers[i]).String()
	}
	return json.Marshal(v)
}

func (m *ModelMeta) Copy() *ModelMeta {
	x := new(ModelMeta)
	*x = *m
	x.Keepers = make([]ids.ShortID, len(m.Keepers))
	copy(x.Keepers, m.Keepers)
	x.Data = make([]byte, len(m.Data))
	copy(x.Data, m.Data)
	x.raw = make([]byte, len(m.raw))
	copy(x.raw, m.raw)
	return x
}

func (m *ModelMeta) SchemaType() schema.Type {
	return m.st
}

// SyntacticVerify verifies that a *ModelMeta is well-formed.
func (m *ModelMeta) SyntacticVerify() error {
	if !modelNameReg.MatchString(m.Name) {
		return fmt.Errorf("invalid model name")
	}
	if len(m.Keepers) > math.MaxUint8 {
		return fmt.Errorf("too many model keepers")
	}
	if int(m.Threshold) > len(m.Keepers) {
		return fmt.Errorf("invalid model threshold")
	}
	for _, id := range m.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid model keeper")
		}
	}
	if len(m.Data) < 10 {
		return fmt.Errorf("model schema bytes should >= %d", 10)
	}

	var err error
	if m.st, err = NewSchemaType(m.Name, m.Data); err != nil {
		return fmt.Errorf("parse ipld model schema error: %v", err)
	}
	if _, err = m.Marshal(); err != nil {
		return fmt.Errorf("modelmeta marshal error: %v", err)
	}
	return nil
}

func (m *ModelMeta) Equal(o *ModelMeta) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(m.raw) > 0 {
		return bytes.Equal(o.raw, m.raw)
	}
	if o.Name != m.Name {
		return false
	}
	if o.Threshold != m.Threshold {
		return false
	}
	if len(o.Keepers) != len(m.Keepers) {
		return false
	}
	for i := range m.Keepers {
		if o.Keepers[i] != m.Keepers[i] {
			return false
		}
	}
	return bytes.Equal(o.Data, m.Data)
}

func (m *ModelMeta) Bytes() []byte {
	if len(m.raw) == 0 {
		if _, err := m.Marshal(); err != nil {
			panic(err)
		}
	}

	return m.raw
}

func (m *ModelMeta) Unmarshal(data []byte) error {
	p, err := modelMetaLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindModelMeta); ok {
		m.Name = v.Name
		m.Threshold = v.Threshold.Value()
		m.Data = v.Data
		if m.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		m.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindModelMeta")
}

func (m *ModelMeta) Marshal() ([]byte, error) {
	v := &bindModelMeta{
		Name:      m.Name,
		Threshold: FromUint8(m.Threshold),
		Keepers:   FromShortIDs(m.Keepers),
		Data:      m.Data,
	}
	data, err := modelMetaLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	m.raw = data
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
		Keepers   [ID20] (rename "ks")
		Data      Bytes  (rename "d")
	}
`
	builder, err := NewLDBuilder("ModelMeta", []byte(sch), (*bindModelMeta)(nil))
	if err != nil {
		panic(err)
	}
	modelMetaLDBuilder = builder
}
