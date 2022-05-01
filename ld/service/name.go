// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"golang.org/x/net/idna"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Name struct {
	Name    string      // should be ASCII form (IDNA2008)
	Linked  ids.ShortID // optional, linked (ProfileService) data id
	Records []string    // DNS resource records

	// external assignment
	raw []byte
}

type jsonName struct {
	Name    string   `json:"name"`    // Unicode form
	Domain  string   `json:"domain"`  // ASCII form
	Linked  string   `json:"linked"`  // optional, linked (ProfileService) data id
	Records []string `json:"records"` // DNS resource records
}

func NameSchema() (string, []byte) {
	return nameLDBuilder.Name(), nameLDBuilder.Schema()
}

func (n *Name) MarshalJSON() ([]byte, error) {
	if n == nil {
		return util.Null, nil
	}
	name, err := idna.Registration.ToUnicode(n.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s, error: %v",
			strconv.Quote(n.Name), err)
	}
	v := &jsonName{
		Name:    name,
		Domain:  n.Name,
		Linked:  util.DataID(n.Linked).String(),
		Records: n.Records,
	}
	return json.Marshal(v)
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

func (n *Name) Equal(o *Name) bool {
	if o == nil {
		return false
	}

	if len(o.raw) > 0 && len(n.raw) > 0 {
		return bytes.Equal(o.raw, n.raw)
	}
	if o.Name != n.Name {
		return false
	}
	if o.Linked != n.Linked {
		return false
	}
	if len(o.Records) != len(n.Records) {
		return false
	}
	for i, v := range o.Records {
		if n.Records[i] != v {
			return false
		}
	}
	return true
}

func (n *Name) Bytes() []byte {
	if len(n.raw) == 0 {
		if _, err := n.Marshal(); err != nil {
			panic(err)
		}
	}

	return n.raw
}

func (n *Name) Unmarshal(data []byte) error {
	p, err := nameLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindName); ok {
		n.Name = v.Name
		n.Records = v.Records
		if n.Linked, err = ld.PtrToShortID(v.Linked); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		n.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *Name")
}

func (n *Name) Marshal() ([]byte, error) {
	v := &bindName{
		Name:    n.Name,
		Linked:  ld.PtrFromShortID(ids.ShortID(n.Linked)),
		Records: n.Records,
	}
	data, err := nameLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	n.raw = data
	return data, nil
}

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
