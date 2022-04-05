// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
)

type DataMeta struct {
	ModelID   ids.ShortID   // model id
	Version   uint64        // data version，数据更新状态号，初始值为 0，每次更新 +1, -1 表示该数据作废，如删除、被屏蔽
	Threshold uint8         // 修改、删除数据时要求的签名阈值. 必须小于等于 Owners 数量，0 表示不可修改，1 表示任何一个 owner 签名即可修改。
	Owners    []ids.ShortID // 数据拥有者，为空时则数据没有拥有者，不能大于 16 个
	Data      []byte
	raw       []byte
}

func (d *DataMeta) Copy() *DataMeta {
	x := new(DataMeta)
	*x = *d
	x.Owners = make([]ids.ShortID, len(d.Owners))
	copy(x.Owners, d.Owners)
	x.Data = make([]byte, len(d.Data))
	copy(x.Data, d.Data)
	x.raw = make([]byte, len(d.raw))
	copy(x.raw, d.raw)
	return x
}

// SyntacticVerify verifies that a *DataMeta is well-formed.
func (d *DataMeta) SyntacticVerify() error {
	if d.ModelID == ids.ShortEmpty {
		return fmt.Errorf("invalid data ModelID")
	}
	if len(d.Owners) > 16 {
		return fmt.Errorf("too many data Owners")
	}
	if int(d.Threshold) > len(d.Owners) {
		return fmt.Errorf("invalid data Threshold")
	}
	for _, id := range d.Owners {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid data Owner")
		}
	}
	if _, err := d.Marshal(); err != nil {
		return fmt.Errorf("datameta marshal error: %v", err)
	}
	return nil
}

func (d *DataMeta) Equal(o *DataMeta) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(d.raw) > 0 {
		return bytes.Equal(o.raw, d.raw)
	}
	if o.ModelID != d.ModelID {
		return false
	}
	if o.Version != d.Version {
		return false
	}
	if o.Threshold != d.Threshold {
		return false
	}
	if len(o.Owners) != len(d.Owners) {
		return false
	}
	for i := range d.Owners {
		if o.Owners[i] != d.Owners[i] {
			return false
		}
	}
	return bytes.Equal(o.Data, d.Data)
}

func (d *DataMeta) Bytes() []byte {
	if len(d.raw) == 0 {
		if _, err := d.Marshal(); err != nil {
			panic(err)
		}
	}

	return d.raw
}

func (d *DataMeta) Unmarshal(data []byte) error {
	p, err := dataMetaLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindDataMeta); ok {
		d.Version = v.Version.Value()
		d.Threshold = v.Threshold.Value()
		d.Data = v.Data
		if d.ModelID, err = ToShortID(v.ModelID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if d.Owners, err = ToShortIDs(v.Owners); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		d.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindModelMeta")
}

func (d *DataMeta) Marshal() ([]byte, error) {
	v := &bindDataMeta{
		ModelID:   FromShortID(d.ModelID),
		Version:   FromUint64(d.Version),
		Threshold: FromUint8(d.Threshold),
		Owners:    FromShortIDs(d.Owners),
		Data:      d.Data,
	}
	data, err := dataMetaLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	d.raw = data
	return data, nil
}

type bindDataMeta struct {
	ModelID   []byte
	Version   Uint64
	Threshold Uint8
	Owners    [][]byte
	Data      []byte
}

var dataMetaLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type DataMeta struct {
		ModelID   ID20   (rename "m")
		Version   Uint64 (rename "v")
		Threshold Uint8  (rename "th")
		Owners    [ID20] (rename "os")
		Data      Bytes  (rename "d")
	}
`

	builder, err := NewLDBuilder("DataMeta", []byte(sch), (*bindDataMeta)(nil))
	if err != nil {
		panic(err)
	}
	dataMetaLDBuilder = builder
}
