// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

var ProfileType = map[string]struct{}{
	"Thing":        {},
	"Person":       {},
	"Organization": {},
}

// https://schema.org/Thing
type Profile struct {
	Type    string                 `cbor:"t" json:"@type"`                       // Thing, Person, Organization...
	Name    string                 `cbor:"n" json:"name"`                        // Thing property
	Image   string                 `cbor:"i" json:"image"`                       // Thing property
	URL     string                 `cbor:"u" json:"url"`                         // Thing property
	Follows []util.DataID          `cbor:"fs" json:"follows"`                    // optional, other ProfileService data id
	KYC     *util.DataID           `cbor:"k,omitempty" json:"kyc,omitempty"`     // optional, KYC (SomeKYCService) data id
	ExMID   *util.ModelID          `cbor:"eid,omitempty" json:"exMid,omitempty"` // optional, extra model id
	Extra   map[string]interface{} `cbor:"ex" json:"extra"`                      // optional, extra properties

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

func ProfileModel() (*ld.IPLDModel, error) {
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
		extra   {String:Any}  (rename "ex")
	}
`
	return ld.NewIPLDModel("ProfileService", []byte(sch))
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (p *Profile) SyntacticVerify() error {
	if p == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil pointer")
	}
	if _, ok := ProfileType[p.Type]; !ok {
		return fmt.Errorf("Name.SyntacticVerify failed: invalid type %s", strconv.Quote(p.Type))
	}
	if !util.ValidName(p.Name) {
		return fmt.Errorf("Name.SyntacticVerify failed: invalid name %s", strconv.Quote(p.Name))
	}
	if !util.ValidLink(p.Image) {
		return fmt.Errorf("Name.SyntacticVerify failed: invalid image %s", strconv.Quote(p.Image))
	}
	if !util.ValidLink(p.URL) {
		return fmt.Errorf("Name.SyntacticVerify failed: invalid url %s", strconv.Quote(p.URL))
	}
	if p.Follows == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil follows")
	}
	for _, id := range p.Follows {
		if id == util.DataIDEmpty {
			return fmt.Errorf("Name.SyntacticVerify failed: invalid follow address")
		}
	}
	if p.Extra == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil extra")
	}
	if _, err := p.Marshal(); err != nil {
		return fmt.Errorf("Name.SyntacticVerify marshal error: %v", err)
	}
	return nil
}

func (p *Profile) Bytes() []byte {
	if len(p.raw) == 0 {
		p.raw = ld.MustMarshal(p)
	}
	return p.raw
}

func (p *Profile) Unmarshal(data []byte) error {
	return ld.DecMode.Unmarshal(data, p)
}

func (p *Profile) Marshal() ([]byte, error) {
	return ld.EncMode.Marshal(p)
}
