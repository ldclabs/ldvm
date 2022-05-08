// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

// https://schema.org/Thing
type Profile struct {
	Type    string                 `cbor:"t" json:"@type"`                       // Thing, Person, Organization...
	Name    string                 `cbor:"n" json:"name"`                        // Thing property
	Image   string                 `cbor:"i" json:"image"`                       // Thing property
	URL     string                 `cbor:"u" json:"url"`                         // Thing property
	Follows []util.DataID          `cbor:"fs" json:"follows"`                    // optional, other ProfileService data id
	KYC     util.DataID            `cbor:"k,omitempty" json:"kyc,omitempty"`     // optional, KYC (SomeKYCService) data id
	ExMID   util.ModelID           `cbor:"eid,omitempty" json:"exMid,omitempty"` // optional, extra model id
	Extra   map[string]interface{} `cbor:"ex" json:"extra"`                      // optional, extra properties
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
		if id == util.DataIDEmpty {
			return fmt.Errorf("invalid follow address")
		}
	}
	if _, err := p.Marshal(); err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}
	return nil
}

func (p *Profile) Unmarshal(data []byte) error {
	return ld.DecMode.Unmarshal(data, p)
}

func (p *Profile) Marshal() ([]byte, error) {
	data, err := ld.EncMode.Marshal(p)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var ProfileModel *ld.IPLDModel

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
		extra   {String:Any}  (rename "ex")
	}
`
	var err error
	ProfileModel, err = ld.NewIPLDModel("ProfileService", []byte(sch))
	if err != nil {
		panic(err)
	}
}
