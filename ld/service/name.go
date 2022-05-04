// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"golang.org/x/net/idna"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Name struct {
	Name    string      `cbor:"n" json:"name"`     // should be ASCII form (IDNA2008)
	Linked  util.DataID `cbor:"l" json:"linked"`   // optional, linked (ProfileService) data id
	Records []string    `cbor:"rs" json:"records"` // DNS resource records

	// external assignment
	DisplayName string `cbor:"-" json:"display"` // Unicode form
}

func NameSchema() (string, []byte) {
	return nameLDBuilder.Name(), nameLDBuilder.Schema()
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

// func (n *Name) Unmarshal(data []byte) error {
// 	p, err := nameLDBuilder.Unmarshal(data)
// 	if err != nil {
// 		return err
// 	}
// 	if v, ok := p.(*bindName); ok {
// 		n.Name = v.Name
// 		n.Records = v.Records
// 		linked, err := ld.PtrToShortID(v.Linked)
// 		if err != nil {
// 			return fmt.Errorf("unmarshal error: %v", err)
// 		}
// 		n.Linked = util.DataID(linked)
// 		n.raw = data
// 		return nil
// 	}
// 	return fmt.Errorf("unmarshal error: expected *Name")
// }

// func (n *Name) Marshal() ([]byte, error) {
// 	v := &bindName{
// 		Name:    n.Name,
// 		Linked:  ld.PtrFromShortID(ids.ShortID(n.Linked)),
// 		Records: n.Records,
// 	}
// 	data, err := nameLDBuilder.Marshal(v)
// 	if err != nil {
// 		return nil, err
// 	}
// 	n.raw = data
// 	return data, nil
// }

type bindName struct {
	Name    string
	Linked  *[]byte
	Records []string
}

var nameLDBuilder *ld.LDBuilder

func init() {
	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
		records [String]      (rename "rs")
	}
`
	builder, err := ld.NewLDBuilder("NameService", []byte(sch), (*bindName)(nil))
	if err != nil {
		panic(err)
	}
	nameLDBuilder = builder
}
