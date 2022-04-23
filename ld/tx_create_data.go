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
	"github.com/ldclabs/ldvm/util"
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

func (t *DataMeta) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonDataMeta{
		ID:        util.DataID(t.ID).String(),
		ModelID:   util.ModelID(t.ModelID).String(),
		Version:   t.Version,
		Threshold: t.Threshold,
		Data:      util.JSONMarshalData(t.Data),
		Keepers:   make([]string, len(t.Keepers)),
	}
	for i := range t.Keepers {
		v.Keepers[i] = util.EthID(t.Keepers[i]).String()
	}
	return json.Marshal(v)
}

func (t *DataMeta) Copy() *DataMeta {
	x := new(DataMeta)
	*x = *t
	x.Keepers = make([]ids.ShortID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = nil
	return x
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
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid keeper")
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("DataMeta marshal error: %v", err)
	}
	return nil
}

func (t *DataMeta) Validate(st schema.Type) error {
	// TODO: validate data with schema.Type
	return nil
}

func (t *DataMeta) Equal(o *DataMeta) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.ModelID != t.ModelID {
		return false
	}
	if o.Version != t.Version {
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

func (t *DataMeta) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *DataMeta) Unmarshal(data []byte) error {
	p, err := dataMetaLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindDataMeta); ok {
		t.Version = v.Version.Value()
		t.Threshold = v.Threshold.Value()
		t.Data = v.Data
		if t.ModelID, err = ToShortID(v.ModelID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.Keepers, err = ToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindDataMeta")
}

func (t *DataMeta) Marshal() ([]byte, error) {
	v := &bindDataMeta{
		ModelID:   FromShortID(t.ModelID),
		Version:   FromUint64(t.Version),
		Threshold: FromUint8(t.Threshold),
		Keepers:   FromShortIDs(t.Keepers),
		Data:      t.Data,
	}
	data, err := dataMetaLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
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
