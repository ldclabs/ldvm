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

	// external assignment
	DisplayName string `cbor:"-" json:"display"` // Unicode form
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
		return fmt.Errorf("invalid Name")
	}
	dn, err := idna.Registration.ToASCII(n.Name)
	if err != nil {
		return fmt.Errorf("invalid name %s, error: %v",
			strconv.Quote(n.Name), err)
	}
	if dn != n.Name {
		return fmt.Errorf("invalid name %s, should be ASCII form (IDNA2008)",
			strconv.Quote(n.Name))
	}
	name, err := idna.Registration.ToUnicode(n.Name)
	if err != nil {
		return fmt.Errorf("invalid name %s, error: %v", strconv.Quote(n.Name), err)
	}
	n.DisplayName = name
	for _, s := range n.Records {
		if !utf8.ValidString(s) {
			return fmt.Errorf("invalid utf8 record %s", strconv.Quote(s))
		}
	}
	if _, err := n.Marshal(); err != nil {
		return fmt.Errorf("Name marshal error: %v", err)
	}
	return nil
}

func (n *Name) Unmarshal(data []byte) error {
	return ld.DecMode.Unmarshal(data, n)
}

func (n *Name) Marshal() ([]byte, error) {
	data, err := ld.EncMode.Marshal(n)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var NameModel *ld.IPLDModel

func init() {
	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
		records [String]      (rename "rs")
	}
`
	var err error
	NameModel, err = ld.NewIPLDModel("NameService", []byte(sch))
	if err != nil {
		panic(err)
	}
}
