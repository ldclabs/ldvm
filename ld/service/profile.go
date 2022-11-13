// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strings"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/validating"
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
	Type  ProfileType `cbor:"t" json:"type"`        // Thing, Person, Organization...
	Name  string      `cbor:"n" json:"name"`        // Thing property
	Desc  string      `cbor:"d" json:"description"` // Thing property
	Image string      `cbor:"i" json:"image"`       // Thing property, relay url
	URL   string      `cbor:"u" json:"url"`         // Thing property, relay url
	// follow other ProfileService data id
	Follows ids.IDList[ids.DataID] `cbor:"fs" json:"follows"`
	// optional, other ProfileService data id
	Members    ids.IDList[ids.DataID] `cbor:"ms,omitempty" json:"members,omitempty"`
	Extensions Extensions             `cbor:"es" json:"extensions"`

	// external assignment fields
	DataID ids.DataID `cbor:"-" json:"did"`
	raw    []byte     `cbor:"-" json:"-"`
}

func ProfileModel() (*ld.IPLDModel, error) {
	schema := `
	type ID20 bytes
	type ProfileService struct {
		type        Int             (rename "t")
		name        String          (rename "n")
		description String          (rename "d")
		image       String          (rename "i")
		url         String          (rename "u")
		follows     [ID20]          (rename "fs")
		members     optional [ID20] (rename "ms")
		extensions  [Any]           (rename "es")
	}
`
	return ld.NewIPLDModel("ProfileService", strings.TrimSpace(schema))
}

// SyntacticVerify verifies that a *Profile is well-formed.
func (p *Profile) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("service.Profile.SyntacticVerify: ")

	switch {
	case p == nil:
		return errp.Errorf("nil pointer")

	case !validating.ValidName(p.Name):
		return errp.Errorf("invalid name %q", p.Name)

	case !validating.ValidMessage(p.Desc):
		return errp.Errorf("invalid description %q", p.Desc)

	case !validating.ValidLink(p.Image):
		return errp.Errorf("invalid image %q", p.Image)

	case !validating.ValidLink(p.URL):
		return errp.Errorf("invalid url %q", p.URL)

	case p.Follows == nil:
		return errp.Errorf("nil follows")
	}

	if len(p.Follows) > 1024 {
		return errp.Errorf("too many follows, should not exceed 1024")
	}

	if err = p.Follows.Valid(); err != nil {
		return errp.Errorf("invalid follows, %v", err)
	}

	if len(p.Members) > 1024 {
		return errp.Errorf("too many follows, should not exceed 1024")
	}

	if err = p.Members.Valid(); err != nil {
		return errp.Errorf("invalid members, %v", err)
	}

	if err = p.Extensions.SyntacticVerify(); err != nil {
		return errp.Errorf("nil extensions")
	}

	if p.raw, err = p.Marshal(); err != nil {
		return errp.ErrorIf(err)
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
	return erring.ErrPrefix("service.Profile.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, p))
}

func (p *Profile) Marshal() ([]byte, error) {
	return erring.ErrPrefix("service.Profile.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(p))
}
