// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/ava-labs/avalanchego/ids"
)

var modelNameReg = regexp.MustCompile(`^[A-Z][0-9A-Za-z]{1,63}$`)

type ModelMeta struct {
	Name      string        // model 中输出的类型名称
	Threshold uint8         // 使用 model 创建数据时要求的签名阈值. 必须小于等于 Keepers 数量，0 表示无需 keeper 签名，1 表示任何一个 keeper 签名即可
	Keepers   []ids.ShortID // 当 Keepers 不存在时，任何人都能使用该 model 创建数据，否则需要 keepers 签名确认，不能大于 16 个
	Data      []byte
	raw       []byte
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

// SyntacticVerify verifies that a *ModelMeta is well-formed.
func (m *ModelMeta) SyntacticVerify() error {
	if !modelNameReg.MatchString(m.Name) {
		return fmt.Errorf("invalid model Name")
	}
	if len(m.Keepers) > 16 {
		return fmt.Errorf("too many model Keepers")
	}
	if int(m.Threshold) > len(m.Keepers) {
		return fmt.Errorf("invalid model Threshold")
	}
	for _, id := range m.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid model Keeper")
		}
	}
	if _, err := m.Marshal(); err != nil {
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
