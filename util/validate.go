// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"net/url"
	"unicode"
	"unicode/utf8"
)

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

	return lastRune != ' ' && len(name) < 256 && utf8.ValidString(name)
}

func ValidLink(link string) bool {
	if link == "" {
		return true
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.String() == link && len(link) < 512 && utf8.ValidString(link)
}

func ValidMID(mid string) bool {
	if mid == "" {
		return true
	}
	_, err := ModelIDFromString(mid)
	return err == nil
}

func ValidMessage(msg string) bool {
	for i, r := range msg {
		if i == 0 && unicode.IsSpace(r) {
			return false
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}

	return len(msg) <= 1024 && utf8.ValidString(msg)
}

func ValidStakeAddress(id EthID) bool {
	return id[0] == '$'
}
