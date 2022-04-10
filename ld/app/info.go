// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
)

// https://schema.org/Person
type Info struct {
	Name     string           `json:"name"`
	Image    string           `json:"image"`
	Url      string           `json:"url"`
	Kyc      string           `json:"kyc"`
	Follows  []string         `json:"follows"`
	Addrs    *MapStringString `json:"addrs"`
	ExtraMID string           `json:"extraMID"`
	Extra    *MapStringAny    `json:"extra"`
}

func InfoSchema() (string, []byte) {
	return bindInfoLDBuilder.Name(), bindInfoLDBuilder.Schema()
}

// entity
type bindInfo struct {
	Entity *Info
	raw    []byte
}

func NewInfo(info *Info) *bindInfo {
	b := new(bindInfo)
	b.Entity = info
	if info == nil {
		b.Entity = new(Info)
	}
	if b.Entity.Addrs == nil {
		b.Entity.Addrs = new(MapStringString)
	}
	if b.Entity.Extra == nil {
		b.Entity.Extra = new(MapStringAny)
	}
	return b
}

func (b *bindInfo) MarshalJSON() ([]byte, error) {
	if b == nil {
		return ld.Null, nil
	}
	return json.Marshal(b.Entity)
}

// SyntacticVerify verifies that a *Info is well-formed.
func (b *bindInfo) SyntacticVerify() error {
	if b == nil || b.Entity == nil {
		return fmt.Errorf("invalid bindInfo")
	}
	if !ValidName(b.Entity.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(b.Entity.Name))
	}
	if !ValidLink(b.Entity.Image) {
		return fmt.Errorf("invalid image string %s", strconv.Quote(b.Entity.Image))
	}
	if !ValidLink(b.Entity.Url) {
		return fmt.Errorf("invalid url string %s", strconv.Quote(b.Entity.Url))
	}
	if !ValidLink(b.Entity.Kyc) {
		return fmt.Errorf("invalid KYC string %s", strconv.Quote(b.Entity.Kyc))
	}
	if !ValidMID(b.Entity.ExtraMID) {
		return fmt.Errorf("invalid model id %s", strconv.Quote(b.Entity.ExtraMID))
	}
	for _, id := range b.Entity.Follows {
		if id == ld.EthIDEmpty.String() {
			return fmt.Errorf("invalid follow address")
		}
	}
	if _, err := b.Marshal(); err != nil {
		return fmt.Errorf("bindInfo marshal error: %v", err)
	}
	return nil
}

func (b *bindInfo) Equal(o *bindInfo) bool {
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
	if o.Entity.Image != b.Entity.Image {
		return false
	}
	if o.Entity.Url != b.Entity.Url {
		return false
	}
	if o.Entity.Kyc != b.Entity.Kyc {
		return false
	}
	if o.Entity.ExtraMID != b.Entity.ExtraMID {
		return false
	}

	if len(o.Entity.Follows) != len(b.Entity.Follows) {
		return false
	}
	for i := range b.Entity.Follows {
		if o.Entity.Follows[i] != b.Entity.Follows[i] {
			return false
		}
	}
	if !o.Entity.Addrs.Equal(b.Entity.Addrs) {
		return false
	}
	if !o.Entity.Extra.Equal(b.Entity.Extra) {
		return false
	}
	return true
}

func (b *bindInfo) Bytes() []byte {
	if len(b.raw) == 0 {
		if _, err := b.Marshal(); err != nil {
			panic(err)
		}
	}

	return b.raw
}

func (b *bindInfo) Unmarshal(data []byte) error {
	p, err := bindInfoLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*Info); ok {
		b.Entity = v // TODO: nil point
		b.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *Info")
}

func (b *bindInfo) Marshal() ([]byte, error) {
	data, err := bindInfoLDBuilder.Marshal(b.Entity)
	if err != nil {
		return nil, err
	}
	b.raw = data
	return data, nil
}

func (b *bindInfo) ToJSON() ([]byte, error) {
	return bindInfoLDBuilder.ToJSON(b.Entity)
}

var bindInfoLDBuilder *ld.LDBuilder

func init() {
	sch := `
	type Info struct {
		name     String
		image    String
		url      String
		kyc      String
		follows  [String]
		addrs    {String:String}
		extraMID String
		extra    {String:Any}
	}
`
	builder, err := ld.NewLDBuilder("Info", []byte(sch), (*Info)(nil))
	if err != nil {
		panic(err)
	}
	bindInfoLDBuilder = builder
}
