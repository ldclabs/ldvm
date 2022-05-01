// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/util"
)

// TxTester TODO
type TxTester struct {
	Data string
	// external assignment
	raw []byte
}

type jsonTxTester struct {
	Data string `json:"data,omitempty"`
}

func (t *TxTester) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxTester{
		Data: t.Data,
	}
	return json.Marshal(v)
}

func (t *TxTester) Copy() *TxTester {
	x := new(TxTester)
	*x = *t
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxTester is well-formed.
func (t *TxTester) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxTester")
	}

	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxTester marshal error: %v", err)
	}
	return nil
}

func (t *TxTester) Equal(o *TxTester) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	return true
}

func (t *TxTester) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxTester) Unmarshal(data []byte) error {
	p, err := txTesterLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxTester); ok {
		t.Data = v.Data
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxTester")
}

func (t *TxTester) Marshal() ([]byte, error) {
	v := &bindTxTester{
		Data: t.Data,
	}

	data, err := txTesterLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindTxTester struct {
	Data string
}

var txTesterLDBuilder *LDBuilder

func init() {
	sch := `
	type TxTester struct {
		Data String  (rename "d")
	}
`

	builder, err := NewLDBuilder("TxTester", []byte(sch), (*bindTxTester)(nil))
	if err != nil {
		panic(err)
	}
	txTesterLDBuilder = builder
}
