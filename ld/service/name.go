// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"
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
	raw []byte `cbor:"-" json:"-"`
}

func NameModel() (*ld.IPLDModel, error) {
	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  optional ID20 (rename "l")
		records [String]      (rename "rs")
	}
`
	return ld.NewIPLDModel("NameService", []byte(sch))
}

type lazyName struct {
	Name    string          `cbor:"n"`
	Linked  cbor.RawMessage `cbor:"l,omitempty"`
	Records cbor.RawMessage `cbor:"rs"`
}

func GetName(data []byte) (string, error) {
	n := &lazyName{}
	if err := ld.DecMode.Unmarshal(data, n); err != nil {
		return "", err
	}
	return n.Name, nil
}

// SyntacticVerify verifies that a *Name is well-formed.
func (n *Name) SyntacticVerify() error {
	errPrefix := "Name.SyntacticVerify failed:"
	if n == nil {
		return fmt.Errorf("%s nil pointer", errPrefix)
	}

	dn, err := NewDN(n.Name)
	if err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if dn.String() != n.Name {
		return fmt.Errorf("%s %s is not unicode form", errPrefix, strconv.Quote(n.Name))
	}

	if n.Records == nil {
		return fmt.Errorf("%s nil records", errPrefix)
	}
	for _, s := range n.Records {
		if !utf8.ValidString(s) {
			return fmt.Errorf("%s invalid utf8 record %s", errPrefix, strconv.Quote(s))
		}
	}
	if n.raw, err = n.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	return nil
}

func (n *Name) Bytes() []byte {
	if len(n.raw) == 0 {
		n.raw = ld.MustMarshal(n)
	}
	return n.raw
}

func (n *Name) Unmarshal(data []byte) error {
	return ld.DecMode.Unmarshal(data, n)
}

func (n *Name) Marshal() ([]byte, error) {
	return ld.EncMode.Marshal(n)
}
