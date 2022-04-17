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
type Profile struct {
	Name     string           `json:"name"`
	Image    string           `json:"image"`
	Url      string           `json:"url"`
	Kyc      string           `json:"kyc"`
	Follows  []string         `json:"follows"`
	Addrs    *MapStringString `json:"addrs"`
	ExtraMID string           `json:"extraMID"`
	Extra    *MapStringAny    `json:"extra"`
}

func ProfileSchema() (string, []byte) {
	return bindProfileLDBuilder.Name(), bindProfileLDBuilder.Schema()
}

// entity
type bindProfile struct {
	Entity *Profile
	raw    []byte
}

func NewProfile(profile *Profile) *bindProfile {
	b := new(bindProfile)
	b.Entity = profile
	if profile == nil {
		b.Entity = new(Profile)
	}
	if b.Entity.Addrs == nil {
		b.Entity.Addrs = new(MapStringString)
	}
	if b.Entity.Extra == nil {
		b.Entity.Extra = new(MapStringAny)
	}
	return b
}

func (b *bindProfile) MarshalJSON() ([]byte, error) {
	if b == nil {
		return ld.Null, nil
	}
	return json.Marshal(b.Entity)
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (b *bindProfile) SyntacticVerify() error {
	if b == nil || b.Entity == nil {
		return fmt.Errorf("invalid bindProfile")
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
		return fmt.Errorf("bindProfile marshal error: %v", err)
	}
	return nil
}

func (b *bindProfile) Equal(o *bindProfile) bool {
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

func (b *bindProfile) Bytes() []byte {
	if len(b.raw) == 0 {
		if _, err := b.Marshal(); err != nil {
			panic(err)
		}
	}

	return b.raw
}

func (b *bindProfile) Unmarshal(data []byte) error {
	p, err := bindProfileLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*Profile); ok {
		b.Entity = v
		b.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *Profile")
}

func (b *bindProfile) Marshal() ([]byte, error) {
	data, err := bindProfileLDBuilder.Marshal(b.Entity)
	if err != nil {
		return nil, err
	}
	b.raw = data
	return data, nil
}

func (b *bindProfile) ToJSON() ([]byte, error) {
	return bindProfileLDBuilder.ToJSON(b.Entity)
}

var bindProfileLDBuilder *ld.LDBuilder

func init() {
	sch := `
	type ProfileApp struct {
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
	builder, err := ld.NewLDBuilder("ProfileApp", []byte(sch), (*Profile)(nil))
	if err != nil {
		panic(err)
	}
	bindProfileLDBuilder = builder
}
