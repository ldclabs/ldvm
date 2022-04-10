// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"net/url"
	"unicode"

	"github.com/ldclabs/ldvm/ld"
)

func ValidDomainName(name string) bool {
	// [0-9A-Za-z-.]
	prePunct := true
	for _, r := range name {
		u := uint32(r)
		switch {
		case u == 45 || u == 46:
			if prePunct {
				return false
			}
			prePunct = true
			continue
		case (u >= 48 && u <= 57) || (u >= 65 && u <= 90) || (u >= 97 && u <= 122):
			prePunct = false
			continue
		default:
			return false
		}
	}

	return !prePunct && len(name) < 256
}

func ValidName(name string) bool {
	preSpace := true
	lastRune := ' '
	for _, r := range name {
		lastRune = r
		switch {
		case r == ' ':
			if preSpace {
				return false
			}
			preSpace = true
			continue
		case unicode.IsSpace(r):
			return false
		case unicode.IsPrint(r):
			preSpace = false
			continue
		default:
			return false
		}
	}

	return lastRune != ' ' && len(name) < 256
}

func ValidLink(link string) bool {
	if link == "" {
		return true
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.String() == link && len(link) < 512
}

func ValidMID(mid string) bool {
	if mid == "" {
		return true
	}
	_, err := ld.ModelIDFromString(mid)
	return err == nil
}
