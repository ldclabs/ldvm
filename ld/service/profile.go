// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

// https://schema.org/Thing
type Profile struct {
	Type    string           // Thing, Person, Organization...
	Name    string           // Thing property
	Image   string           // Thing property
	URL     string           // Thing property
	KYC     ids.ShortID      // optional, KYC (SomeKYCService) data id
	Follows []ids.ShortID    // optional, other ProfileService data id
	ExMID   ids.ShortID      // optional, extra model id
	Extra   *ld.MapStringAny // optional, extra properties

	// external assignment
	raw []byte
}

type jsonProfile struct {
	Type    string           `json:"@type"`
	Name    string           `json:"name"`
	Image   string           `json:"image"`
	URL     string           `json:"url"`
	KYC     string           `json:"kyc"`
	Follows []string         `json:"follows"`
	ExMID   string           `json:"exmID"`
	Extra   *ld.MapStringAny `json:"extra"`
}

func ProfileSchema() (string, []byte) {
	return profileLDBuilder.Name(), profileLDBuilder.Schema()
}

func (p *Profile) MarshalJSON() ([]byte, error) {
	if p == nil {
		return util.Null, nil
	}
	v := &jsonProfile{
		Type:    p.Type,
		Name:    p.Name,
		Image:   p.Image,
		URL:     p.URL,
		Follows: make([]string, len(p.Follows)),
		Extra:   p.Extra,
	}
	if p.KYC != ids.ShortEmpty {
		v.KYC = util.DataID(p.KYC).String()
	}
	if p.ExMID != ids.ShortEmpty {
		v.ExMID = util.DataID(p.ExMID).String()
	}
	for i, id := range p.Follows {
		v.Follows[i] = util.DataID(id).String()
	}
	return json.Marshal(v)
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (p *Profile) SyntacticVerify() error {
	if p == nil {
		return fmt.Errorf("invalid Profile")
	}
	if !util.ValidName(p.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(p.Name))
	}
	if !util.ValidLink(p.Image) {
		return fmt.Errorf("invalid image string %s", strconv.Quote(p.Image))
	}
	if !util.ValidLink(p.URL) {
		return fmt.Errorf("invalid url string %s", strconv.Quote(p.URL))
	}

	for _, id := range p.Follows {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid follow address")
		}
	}
	if _, err := p.Marshal(); err != nil {
		return fmt.Errorf("bindProfile marshal error: %v", err)
	}
	return nil
}

func (p *Profile) Equal(o *Profile) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(p.raw) > 0 {
		return bytes.Equal(o.raw, p.raw)
	}
	if o.Type != p.Type {
		return false
	}
	if o.Name != p.Name {
		return false
	}
	if o.Image != p.Image {
		return false
	}
	if o.URL != p.URL {
		return false
	}
	if o.KYC != p.KYC {
		return false
	}
	if o.ExMID != p.ExMID {
		return false
	}

	if len(o.Follows) != len(p.Follows) {
		return false
	}
	for i := range p.Follows {
		if o.Follows[i] != p.Follows[i] {
			return false
		}
	}

	if !o.Extra.Equal(p.Extra) {
		return false
	}
	return true
}

func (b *Profile) Bytes() []byte {
	if len(b.raw) == 0 {
		if _, err := b.Marshal(); err != nil {
			panic(err)
		}
	}

	return b.raw
}

func (p *Profile) Unmarshal(data []byte) error {
	pp, err := profileLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := pp.(*bindProfile); ok {
		p.Type = v.Type
		p.Name = v.Name
		p.Image = v.Image
		p.URL = v.Url
		p.Extra = v.Extra
		if p.Follows, err = ld.ToShortIDs(v.Follows); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if p.KYC, err = ld.PtrToShortID(v.Kyc); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if p.ExMID, err = ld.PtrToShortID(v.ExMID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		p.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *Profile")
}

func (p *Profile) Marshal() ([]byte, error) {
	v := &bindProfile{
		Type:    p.Type,
		Name:    p.Name,
		Image:   p.Image,
		Url:     p.URL,
		Follows: ld.FromShortIDs(p.Follows),
		Kyc:     ld.PtrFromShortID(p.KYC),
		ExMID:   ld.PtrFromShortID(p.ExMID),
		Extra:   p.Extra,
	}
	data, err := profileLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	p.raw = data
	return data, nil
}

type bindProfile struct {
	Type    string
	Name    string
	Image   string
	Url     string
	Follows [][]byte
	Kyc     *[]byte
	ExMID   *[]byte
	Extra   *ld.MapStringAny
}

var profileLDBuilder *ld.LDBuilder

func init() {
	sch := `
	type ID20 bytes
	type ProfileService struct {
		type    String        (rename "t")
		name    String        (rename "n")
		image   String        (rename "i")
		url     String        (rename "u")
		follows [ID20]        (rename "fs")
		kyc     nullable ID20 (rename "k")
		exMID   nullable ID20 (rename "eid")
		extra   nullable {String:Any} (rename "ex")
	}
`
	builder, err := ld.NewLDBuilder("ProfileService", []byte(sch), (*bindProfile)(nil))
	if err != nil {
		panic(err)
	}
	profileLDBuilder = builder
}
