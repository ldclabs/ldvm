// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"strings"
	"unicode/utf8"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

type Name struct {
	// name should be Unicode form
	Name string `cbor:"n" json:"name"`
	// optional, linked (ProfileService) data id
	Linked     *ids.DataID `cbor:"l,omitempty" json:"linked,omitempty"`
	Records    []string    `cbor:"rs" json:"records"` // DNS resource records
	Extensions Extensions  `cbor:"es" json:"extensions"`

	// external assignment fields
	DataID ids.DataID `cbor:"-" json:"did"`
	raw    []byte     `cbor:"-" json:"-"`
	dn     *DN        `cbor:"-" json:"-"`
}

func NameModel() (*ld.IPLDModel, error) {
	schema := `
	type ID20 bytes
	type NameService struct {
		name       String        (rename "n")
		linked     optional ID20 (rename "l")
		records    [String]      (rename "rs")
		extensions [Any]         (rename "es")
	}
`
	return ld.NewIPLDModel("NameService", strings.TrimSpace(schema))
}

type lazyName struct {
	Name string `cbor:"n"`
}

func GetName(data []byte) (string, error) {
	n := &lazyName{}
	if err := encoding.UnmarshalCBOR(data, n); err != nil {
		return "", erring.ErrPrefix("service.GetName: ").ErrorIf(err)
	}
	return n.Name, nil
}

// SyntacticVerify verifies that a *Name is well-formed.
func (n *Name) SyntacticVerify() error {
	errp := erring.ErrPrefix("service.Name.SyntacticVerify: ")
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

	if err = n.Extensions.SyntacticVerify(); err != nil {
		return errp.Errorf("nil extensions")
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
	return erring.ErrPrefix("service.Name.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, n))
}

func (n *Name) Marshal() ([]byte, error) {
	return erring.ErrPrefix("service.Name.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(n))
}
