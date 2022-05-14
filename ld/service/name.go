// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"golang.org/x/net/idna"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Name struct {
	Name    string       `cbor:"n" json:"name"`                       // should be ASCII form (IDNA2008)
	Linked  *util.DataID `cbor:"l,omitempty" json:"linked,omitempty"` // optional, linked (ProfileService) data id
	Records []string     `cbor:"rs" json:"records"`                   // DNS resource records

	// external assignment fields
	DisplayName string `cbor:"-" json:"displayName"` // Unicode form
	raw         []byte `cbor:"-" json:"-"`
}

func NameModel() (*ld.IPLDModel, error) {
	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
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
	if n == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil pointer")
	}
	dn, err := idna.Registration.ToASCII(n.Name)
	if err != nil {
		return fmt.Errorf("Name.SyntacticVerify failed: converts %s error: %v",
			strconv.Quote(n.Name), err)
	}
	if dn != n.Name {
		return fmt.Errorf("Name.SyntacticVerify failed: %s is not ASCII form (IDNA2008)",
			strconv.Quote(n.Name))
	}
	name, err := idna.Registration.ToUnicode(n.Name)
	if err != nil {
		return fmt.Errorf("Name.SyntacticVerify failed: converts %s error: %v", strconv.Quote(n.Name), err)
	}
	n.DisplayName = name
	if n.Records == nil {
		return fmt.Errorf("Name.SyntacticVerify failed: nil records")
	}
	for _, s := range n.Records {
		if !utf8.ValidString(s) {
			return fmt.Errorf("Name.SyntacticVerify failed: invalid utf8 record %s", strconv.Quote(s))
		}
	}
	if n.raw, err = n.Marshal(); err != nil {
		return fmt.Errorf("Name.SyntacticVerify marshal error: %v", err)
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
