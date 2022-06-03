// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type ProfileType uint8

func (pt ProfileType) String() string {
	switch pt {
	case 0:
		return "Thing"
	case 1:
		return "Person"
	case 2:
		return "Organization"
	default:
		return fmt.Sprintf("UnknownType(%d)", pt)
	}
}

func (pt ProfileType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + pt.String() + "\""), nil
}

// https://schema.org/Thing
type Profile struct {
	Type       ProfileType   `cbor:"t" json:"@type"`                        // Thing, Person, Organization...
	Name       string        `cbor:"n" json:"name"`                         // Thing property
	Image      string        `cbor:"i" json:"image"`                        // Thing property
	URL        string        `cbor:"u" json:"url"`                          // Thing property
	KYC        *util.DataID  `cbor:"k,omitempty" json:"kyc,omitempty"`      // optional, KYC (SomeKYCService) data id
	Follows    []util.DataID `cbor:"fs" json:"follows"`                     // optional, other ProfileService data id
	Members    []util.DataID `cbor:"ms,omitempty" json:"members,omitempty"` // optional, other ProfileService data id
	Extensions []*Extension  `cbor:"ex" json:"extensions"`                  // optional, extra properties

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

type Extension struct {
	ModelID    util.ModelID           `cbor:"mid" json:"mid"` // model id
	Title      string                 `cbor:"t" json:"title"` // extension title
	Properties map[string]interface{} `cbor:"ps" json:"properties"`
}

func ProfileModel() (*ld.IPLDModel, error) {
	sch := `
	type ID20 bytes
	type ProfileService struct {
		type       Int             (rename "t")
		name       String          (rename "n")
		image      String          (rename "i")
		url        String          (rename "u")
		kyc        optional ID20   (rename "k")
		follows    [ID20]          (rename "fs")
		members    optional [ID20] (rename "ms")
		extensions [Any]           (rename "ex")
	}
`
	return ld.NewIPLDModel("ProfileService", []byte(sch))
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (p *Profile) SyntacticVerify() error {
	if p == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil pointer")
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
	for _, id := range p.Members {
		if id == util.DataIDEmpty {
			return fmt.Errorf("Name.SyntacticVerify failed: invalid member address")
		}
	}
	if p.Extensions == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil extensions")
	}
	set := make(map[string]struct{}, len(p.Extensions))
	for _, ex := range p.Extensions {
		if !util.ValidName(ex.Title) {
			return fmt.Errorf("Name.SyntacticVerify failed: invalid extension title %s", strconv.Quote(ex.Title))
		}
		if _, ok := set[ex.Title]; ok {
			return fmt.Errorf("Name.SyntacticVerify failed: %s exists in extensions", strconv.Quote(ex.Title))
		}
		set[ex.Title] = struct{}{}
	}
	var err error
	if p.raw, err = p.Marshal(); err != nil {
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
