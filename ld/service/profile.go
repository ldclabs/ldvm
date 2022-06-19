// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type ProfileType uint16

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
	Type       ProfileType  `cbor:"t" json:"@type"`                        // Thing, Person, Organization...
	Name       string       `cbor:"n" json:"name"`                         // Thing property
	Image      string       `cbor:"i" json:"image"`                        // Thing property
	URL        string       `cbor:"u" json:"url"`                          // Thing property
	Follows    util.DataIDs `cbor:"fs" json:"follows"`                     // optional, other ProfileService data id
	Members    util.DataIDs `cbor:"ms,omitempty" json:"members,omitempty"` // optional, other ProfileService data id
	Extensions []*Extension `cbor:"ex" json:"extensions"`                  // optional, extra properties

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

type Extension struct {
	ModelID    util.ModelID           `cbor:"m" json:"mid"`   // model id
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
		follows    [ID20]          (rename "fs")
		members    optional [ID20] (rename "ms")
		extensions [Any]           (rename "ex")
	}
`
	return ld.NewIPLDModel("ProfileService", []byte(sch))
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (p *Profile) SyntacticVerify() error {
	var err error
	errPrefix := "Profile.SyntacticVerify failed: "
	switch {
	case p == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case !util.ValidName(p.Name):
		return fmt.Errorf("%s invalid name %s", errPrefix, strconv.Quote(p.Name))

	case !util.ValidLink(p.Image):
		return fmt.Errorf("%s invalid image %s", errPrefix, strconv.Quote(p.Image))

	case !util.ValidLink(p.URL):
		return fmt.Errorf("%s invalid url %s", errPrefix, strconv.Quote(p.URL))

	case p.Follows == nil:
		return fmt.Errorf("%s nil follows", errPrefix)
	}

	if err = p.Follows.CheckDuplicate(); err != nil {
		return fmt.Errorf("%s invalid follows, %v", errPrefix, err)
	}

	if err = p.Follows.CheckEmptyID(); err != nil {
		return fmt.Errorf("%s invalid follows, %v", errPrefix, err)
	}

	if err = p.Members.CheckDuplicate(); err != nil {
		return fmt.Errorf("%s invalid members, %v", errPrefix, err)
	}

	if err = p.Members.CheckEmptyID(); err != nil {
		return fmt.Errorf("%s invalid members, %v", errPrefix, err)
	}

	if p.Extensions == nil {
		return fmt.Errorf("%s nil extensions", errPrefix)
	}
	set := make(map[string]struct{}, len(p.Extensions))
	for i, ex := range p.Extensions {
		if ex == nil {
			return fmt.Errorf("%s nil extension at %d", errPrefix, i)
		}
		if !util.ValidName(ex.Title) {
			return fmt.Errorf("%s invalid extension title %s at %d",
				errPrefix, strconv.Quote(ex.Title), i)
		}
		id := string(ex.ModelID[:]) + ex.Title
		if _, ok := set[id]; ok {
			return fmt.Errorf("%s %s exists in extensions at %d",
				errPrefix, strconv.Quote(ex.Title), i)
		}
		set[id] = struct{}{}
	}

	if p.raw, err = p.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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
	return ld.UnmarshalCBOR(data, p)
}

func (p *Profile) Marshal() ([]byte, error) {
	return ld.MarshalCBOR(p)
}
