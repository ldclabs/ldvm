// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ldclabs/ldvm/util"
	"golang.org/x/net/idna"
)

// Decentralized Name
type DN struct {
	name, ascii string
	isDomain    bool
}

func NewDN(name string) (*DN, error) {
	errp := util.ErrPrefix(fmt.Sprintf("NewDN(%q) error: ", name))
	if !utf8.ValidString(name) || name == "" {
		return nil, errp.Errorf("invalid utf8 name")
	}

	var err error
	dn := &DN{}

	if strings.ContainsRune(name, '.') {
		dn.isDomain = true
		dn.ascii, err = idna.Registration.ToASCII(name)
		if err != nil {
			return nil, errp.Errorf("ToASCII error, %v", err)
		}
		if dn.ascii[len(dn.ascii)-1] != '.' {
			return nil, errp.Errorf("invalid domain name, no trailing dot")
		}
		dn.name, err = idna.Registration.ToUnicode(dn.ascii)
		if err != nil {
			return nil, errp.Errorf("ToUnicode error, %v", err)
		}
		if name != dn.name && name != dn.ascii {
			return nil, errp.Errorf("invalid domain name")
		}
		return dn, nil
	}

	ns := strings.Split(name, ":")
	as := make([]string, len(ns))
	for i, n := range ns {
		as[i], err = idna.Registration.ToASCII(n)
		if err != nil {
			return nil, errp.Errorf("ToASCII error, %v", err)
		}
		ns[i], err = idna.Registration.ToUnicode(as[i])
		if err != nil {
			return nil, errp.Errorf("ToUnicode error, %v", err)
		}
	}
	dn.name = strings.Join(ns, ":")
	dn.ascii = strings.Join(as, ":")
	if name != dn.name && name != dn.ascii {
		return nil, errp.Errorf("invalid decentralized name")
	}
	return dn, nil
}

// String returns the Unicode form of the DN
func (d *DN) String() string {
	return d.name
}

// ASCII returns the ASCII form of the DN
func (d *DN) ASCII() string {
	return d.ascii
}

func (d *DN) IsDomain() bool {
	return d.isDomain
}

// func NameToASCII(name string) (string, error) {
// 	dn, err := NewDN(name)
// 	if err != nil {
// 		return "", err
// 	}
// 	if dn.String() != name {
// 		return "", fmt.Errorf("%q is not unicode form", name)
// 	}

// 	return dn.ASCII(), nil
// }
