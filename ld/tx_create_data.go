// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ipld/go-ipld-prime/schema"
)

type DataMeta struct {
	ModelID ids.ShortID // model id
	// data version，the initial value is 1, should increase 1 when updating,
	// 0 indicates that the data is invalid, for example, deleted or punished.
	Version uint64
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can change the data.
	// the maximum value is len(keepers)
	Threshold uint8
	// keepers who owned this data, no more than 255
	Keepers []ids.ShortID
	Data    []byte

	// external assignment
	raw []byte
	ID  ids.ShortID
}

type jsonDataMeta struct {
	ID        string          `json:"id"`
	ModelID   string          `json:"modelID"`
	Version   uint64          `json:"version"`
	Threshold uint8           `json:"threshold"`
	Keepers   []string        `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (d *DataMeta) MarshalJSON() ([]byte, error) {
	if d == nil {
		return Null, nil
	}
	v := &jsonDataMeta{
		ID:        DataID(d.ID).String(),
		ModelID:   ModelID(d.ModelID).String(),
		Version:   d.Version,
		Threshold: d.Threshold,
		Data:      JsonMarshalData(d.Data),
		Keepers:   make([]string, len(d.Keepers)),
	}
	for i := range d.Keepers {
		v.Keepers[i] = EthID(d.Keepers[i]).String()
	}
	return json.Marshal(v)
}

func (d *DataMeta) Copy() *DataMeta {
	x := new(DataMeta)
	*x = *d
	x.Keepers = make([]ids.ShortID, len(d.Keepers))
	copy(x.Keepers, d.Keepers)
	x.Data = make([]byte, len(d.Data))
	copy(x.Data, d.Data)
	x.raw = make([]byte, len(d.raw))
	copy(x.raw, d.raw)
	return x
}

// SyntacticVerify verifies that a *DataMeta is well-formed.
func (d *DataMeta) SyntacticVerify() error {
	if d == nil {
		return fmt.Errorf("invalid DataMeta")
	}

	if len(d.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(d.Threshold) > len(d.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	for _, id := range d.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid keeper")
		}
	}
	if _, err := d.Marshal(); err != nil {
		return fmt.Errorf("DataMeta marshal error: %v", err)
	}
	return nil
}

func (d *DataMeta) Validate(st schema.Type) error {
	// TODO: validate data with schema.Type
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
	if len(o.Keepers) != len(d.Keepers) {
		return false
	}
	for i := range d.Keepers {
		if o.Keepers[i] != d.Keepers[i] {
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
		if d.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		d.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindDataMeta")
}

func (d *DataMeta) Marshal() ([]byte, error) {
	v := &bindDataMeta{
		ModelID:   FromShortID(d.ModelID),
		Version:   FromUint64(d.Version),
		Threshold: FromUint8(d.Threshold),
		Keepers:   FromShortIDs(d.Keepers),
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
	Keepers   [][]byte
	Data      []byte
}

var dataMetaLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type DataMeta struct {
		ModelID   ID20   (rename "mid")
		Version   Uint64 (rename "v")
		Threshold Uint8  (rename "th")
		Keepers   [ID20] (rename "ks")
		Data      Bytes  (rename "d")
	}
`

	builder, err := NewLDBuilder("DataMeta", []byte(sch), (*bindDataMeta)(nil))
	if err != nil {
		panic(err)
	}
	dataMetaLDBuilder = builder
}
