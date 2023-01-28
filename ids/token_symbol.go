// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"errors"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/util/encoding"
)

// TokenSymbol
type TokenSymbol ID20

var NativeToken = TokenSymbol{}

func TokenFromStr(s string) (TokenSymbol, error) {
	var symbol TokenSymbol
	l := len(s)
	switch {
	case l == 0:
		return symbol, nil
	case l > 1 && l <= 20:
		copy(symbol[20-l:], []byte(s))
		if symbol.String() == s {
			return symbol, nil
		}
	}

	return symbol, errors.New("ids.TokenFromStr: invalid token symbol %q" + strconv.Quote(s))
}

func (id TokenSymbol) Bytes() []byte {
	return id[:]
}

func (id TokenSymbol) Ptr() *TokenSymbol {
	return &id
}

func (id TokenSymbol) String() string {
	start := -1
	for i, r := range id[:] {
		switch {
		case r == 0:
			if start >= 0 || i == 18 {
				return ""
			}
		case r == '$':
			if start == -1 {
				start = i
			} else {
				return ""
			}
		case (r >= 48 && r <= 57) || (r >= 65 && r <= 90):
			if start == -1 {
				return ""
			}
		default:
			return ""
		}
	}
	return string(id[start:])
}

func (id TokenSymbol) Valid() bool {
	return id == NativeToken || id.String() != ""
}

func (id TokenSymbol) GoString() string {
	if id == NativeToken {
		return "NativeLDC"
	}
	if str := id.String(); str != "" {
		return str
	}
	return Address(id).GoString()
}

func (id TokenSymbol) AsKey() cbor.ByteString {
	if id == NativeToken {
		return ""
	}
	if str := id.String(); str != "" {
		return cbor.ByteString(str)
	}
	return Address(id).AsKey()
}

func (id TokenSymbol) Address() Address {
	return Address(id)
}

func (id TokenSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *TokenSymbol) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.TokenSymbol.UnmarshalText: nil pointer")
	}

	sid, err := TokenFromStr(string(b))
	if err != nil {
		return errors.New("ids.TokenSymbol.UnmarshalText: " + err.Error())
	}

	*id = sid
	return nil
}

func (id TokenSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *TokenSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.TokenSymbol.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("ids.TokenSymbol.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id TokenSymbol) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.TokenSymbol.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *TokenSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.TokenSymbol.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("ids.TokenSymbol.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	if !id.Valid() {
		return errors.New("ids.TokenSymbol.UnmarshalCBOR: invalid TokenSymbol " + id.GoString())
	}
	return nil
}
