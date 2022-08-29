// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"strings"
	"unicode/utf8"

	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Name struct {
	Name    string       `cbor:"n" json:"name"`                       // should be Unicode form
	Linked  *util.DataID `cbor:"l,omitempty" json:"linked,omitempty"` // optional, linked (ProfileService) data id
	Records []string     `cbor:"rs" json:"records"`                   // DNS resource records

	// external assignment fields
	DataID util.DataID `cbor:"-" json:"did"`
	raw    []byte      `cbor:"-" json:"-"`
	dn     *DN         `cbor:"-" json:"-"`
}

func NameModel() (*ld.IPLDModel, error) {
	schema := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  optional ID20 (rename "l")
		records [String]      (rename "rs")
	}
`
	return ld.NewIPLDModel("NameService", strings.TrimSpace(schema))
}

type lazyName struct {
	Name    string          `cbor:"n"`
	Linked  cbor.RawMessage `cbor:"l"`
	Records cbor.RawMessage `cbor:"rs"`
}

func GetName(data []byte) (string, error) {
	n := &lazyName{}
	if err := util.UnmarshalCBOR(data, n); err != nil {
		return "", util.ErrPrefix("GetName error: ").ErrorIf(err)
	}
	return n.Name, nil
}

// SyntacticVerify verifies that a *Name is well-formed.
func (n *Name) SyntacticVerify() error {
	errp := util.ErrPrefix("Name.SyntacticVerify error: ")
	if n == nil {
		return errp.Errorf("nil pointer")
	}

	dn, err := NewDN(n.Name)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if dn.String() != n.Name {
		return errp.Errorf("%q is not unicode form", n.Name)
	}

	if n.Records == nil {
		return errp.Errorf("nil records")
	}
	for _, s := range n.Records {
		if !utf8.ValidString(s) {
			return errp.Errorf("invalid utf8 record %q", s)
		}
	}
	if n.raw, err = n.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	n.dn = dn
	return nil
}

func (n *Name) ASCII() string {
	if n.dn == nil {
		dn, err := NewDN(n.Name)
		if err != nil {
			panic(err)
		}
		n.dn = dn
	}
	return n.dn.ASCII()
}

func (n *Name) Bytes() []byte {
	if len(n.raw) == 0 {
		n.raw = ld.MustMarshal(n)
	}
	return n.raw
}

func (n *Name) Unmarshal(data []byte) error {
	return util.ErrPrefix("Name.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, n))
}

func (n *Name) Marshal() ([]byte, error) {
	return util.ErrPrefix("Name.Marshal error: ").
		ErrorMap(util.MarshalCBOR(n))
}
