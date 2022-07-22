// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"net/url"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

func ValidName(name string) bool {
	if len(name) >= 256 || !utf8.ValidString(name) {
		return false
	}

	r := ' '
	preSpace := true
	gs := uniseg.NewGraphemes(name)
	for gs.Next() {
		rs := gs.Runes()
		r = rs[0]
		switch {
		case r == 0x200d || rs[len(rs)-1] == 0x200d: // Zero Width Joiner
			return false

		case len(rs) > 1:
			preSpace = false
			continue

		case r == ' ':
			if preSpace {
				return false
			}
			preSpace = true
			continue

		case unicode.IsPrint(r):
			preSpace = false
			continue

		default:
			return false
		}
	}

	return r != ' '
}

func ValidLink(link string) bool {
	if len(link) > 512 || !utf8.ValidString(link) {
		return false
	}

	if link == "" {
		return true
	}
	if !ValidMessage(link) {
		return false
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.String() == link
}

func ValidMessage(msg string) bool {
	if msg == "" {
		return true
	}

	if len(msg) > 1024 || !utf8.ValidString(msg) {
		return false
	}

	r := ' '
	i := -1
	gs := uniseg.NewGraphemes(msg)
	for gs.Next() {
		rs := gs.Runes()
		r = rs[0]
		i++

		switch {
		case r == 0x200d || rs[len(rs)-1] == 0x200d: // Zero Width Joiner
			return false

		case len(rs) > 1:
			continue

		case unicode.IsSpace(r):
			if i == 0 {
				return false
			}
			continue

		case unicode.IsPrint(r):
			continue

		default:
			return false
		}
	}

	return !unicode.IsSpace(r)
}
