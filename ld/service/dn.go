// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

// Decentralized Name
type DN struct {
	name, ascii string
	isDomain    bool
}

func NewDN(name string) (*DN, error) {
	if !utf8.ValidString(name) || name == "" {
		return nil, fmt.Errorf("NewDN(%s): invalid utf8 name", strconv.Quote(name))
	}

	var err error
	dn := &DN{}

	if strings.IndexRune(name, '.') > -1 {
		dn.isDomain = true
		dn.ascii, err = idna.Registration.ToASCII(name)
		if err != nil {
			return nil, fmt.Errorf("NewDN(%s): ToASCII error, %v", strconv.Quote(name), err)
		}
		if dn.ascii[len(dn.ascii)-1] != '.' {
			return nil, fmt.Errorf("NewDN(%s): invalid domain name, no trailing dot", strconv.Quote(name))
		}
		dn.name, err = idna.Registration.ToUnicode(dn.ascii)
		if err != nil {
			return nil, fmt.Errorf("NewDN(%s): ToUnicode error, %v", strconv.Quote(name), err)
		}
		if name != dn.name && name != dn.ascii {
			return nil, fmt.Errorf("NewDN(%s): invalid domain name", strconv.Quote(name))
		}
		return dn, nil
	}

	ns := strings.Split(name, ":")
	as := make([]string, len(ns))
	for i, n := range ns {
		as[i], err = idna.Registration.ToASCII(n)
		if err != nil {
			return nil, fmt.Errorf("NewDN(%s): ToASCII error, %v", strconv.Quote(name), err)
		}
		ns[i], err = idna.Registration.ToUnicode(as[i])
		if err != nil {
			return nil, fmt.Errorf("NewDN(%s): ToUnicode error, %v", strconv.Quote(name), err)
		}
	}
	dn.name = strings.Join(ns, ":")
	dn.ascii = strings.Join(as, ":")
	if name != dn.name && name != dn.ascii {
		return nil, fmt.Errorf("NewDN(%s): invalid decentralized name", strconv.Quote(name))
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
