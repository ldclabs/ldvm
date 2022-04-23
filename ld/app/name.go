// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Name struct {
	Name     string           `json:"name"` // case insensitive
	Linked   string           `json:"linked"`
	ExtraMID string           `json:"extraMID"`
	Extra    *ld.MapStringAny `json:"extra"`
}

type bindName struct {
	Entity *Name
	raw    []byte
}

func NameSchema() (string, []byte) {
	return bindNameLDBuilder.Name(), bindNameLDBuilder.Schema()
}

func NameFrom(data []byte) (*bindName, error) {
	n := NewName(nil)
	if err := n.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := n.SyntacticVerify(); err != nil {
		return nil, err
	}
	return n, nil
}

func NewName(n *Name) *bindName {
	b := new(bindName)
	b.Entity = n
	if n == nil {
		b.Entity = new(Name)
	}
	if b.Entity.Extra == nil {
		b.Entity.Extra = new(ld.MapStringAny)
	}
	return b
}

func (b *bindName) MarshalJSON() ([]byte, error) {
	if b == nil {
		return util.Null, nil
	}
	return json.Marshal(b.Entity)
}

// SyntacticVerify verifies that a *Name is well-formed.
func (b *bindName) SyntacticVerify() error {
	if b == nil || b.Entity == nil {
		return fmt.Errorf("invalid bindName")
	}
	if !util.ValidDomainName(b.Entity.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(b.Entity.Name))
	}
	if !util.ValidLink(b.Entity.Linked) {
		return fmt.Errorf("invalid linked string %s", strconv.Quote(b.Entity.Linked))
	}
	if !util.ValidMID(b.Entity.ExtraMID) {
		return fmt.Errorf("invalid model id %s", strconv.Quote(b.Entity.ExtraMID))
	}

	if _, err := b.Marshal(); err != nil {
		return fmt.Errorf("bindName marshal error: %v", err)
	}
	return nil
}

func (b *bindName) Equal(o *bindName) bool {
	if o == nil {
		return b == nil
	}
	if o.Entity == nil {
		return b.Entity == nil
	}
	if len(o.raw) > 0 && len(b.raw) > 0 {
		return bytes.Equal(o.raw, b.raw)
	}
	if o.Entity.Name != b.Entity.Name {
		return false
	}
	if o.Entity.Linked != b.Entity.Linked {
		return false
	}
	if o.Entity.ExtraMID != b.Entity.ExtraMID {
		return false
	}
	if !o.Entity.Extra.Equal(b.Entity.Extra) {
		return false
	}
	return true
}

func (a *bindName) Bytes() []byte {
	if len(a.raw) == 0 {
		if _, err := a.Marshal(); err != nil {
			panic(err)
		}
	}

	return a.raw
}

func (b *bindName) Unmarshal(data []byte) error {
	p, err := bindNameLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*Name); ok {
		b.Entity = v
		b.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *Name")
}

func (b *bindName) Marshal() ([]byte, error) {
	data, err := bindNameLDBuilder.Marshal(b.Entity)
	if err != nil {
		return nil, err
	}
	b.raw = data
	return data, nil
}

func (b *bindName) ToJSON() ([]byte, error) {
	return bindNameLDBuilder.ToJSON(b.Entity)
}

var bindNameLDBuilder *ld.LDBuilder

func init() {
	sch := `
	type ID20 bytes
	type NameApp struct {
		name     String
		linked   String
		extraMID String
		extra    {String:Any}
	}
`
	builder, err := ld.NewLDBuilder("NameApp", []byte(sch), (*Name)(nil))
	if err != nil {
		panic(err)
	}
	bindNameLDBuilder = builder
}
