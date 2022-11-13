// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"errors"
	"strconv"

	"github.com/ldclabs/ldvm/util/encoding"
)

// StakeSymbol
type StakeSymbol ID20

var EmptyStake = StakeSymbol{}

func StakeFromStr(s string) (StakeSymbol, error) {
	var symbol StakeSymbol
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

	return symbol, errors.New("ids.StakeFromStr: invalid stake symbol " + strconv.Quote(s))
}

func (id StakeSymbol) Bytes() []byte {
	return id[:]
}

func (id StakeSymbol) Ptr() *StakeSymbol {
	return &id
}

func (id StakeSymbol) String() string {
	start := -1
	for i, r := range id[:] {
		switch {
		case r == 0:
			if start >= 0 || i == 18 {
				return ""
			}
		case r == '#':
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

func (id StakeSymbol) Valid() bool {
	return id.String() != ""
}

func (id StakeSymbol) GoString() string {
	if str := id.String(); str != "" {
		return str
	}
	return Address(id).String()
}

func (id StakeSymbol) Address() Address {
	return Address(id)
}

func (id StakeSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *StakeSymbol) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.StakeSymbol.UnmarshalText: nil pointer")
	}

	sid, err := StakeFromStr(string(b))
	if err != nil {
		return errors.New("ids.StakeSymbol.UnmarshalText: " + err.Error())
	}

	*id = sid
	return nil
}

func (id StakeSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *StakeSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.StakeSymbol.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("ids.StakeSymbol.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id StakeSymbol) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.StakeSymbol.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *StakeSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.StakeSymbol.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("ids.StakeSymbol.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	if *id != EmptyStake && !id.Valid() {
		return errors.New("ids.StakeSymbol.UnmarshalCBOR: invalid StakeSymbol " + id.GoString())
	}
	return nil
}
