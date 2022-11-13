// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/ldclabs/ldvm/util/encoding"
)

var (
	// 0x0000000000000000000000000000000000000000
	LDCAccount = Address{}
	// 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff
	GenesisAccount = Address{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
)

type Address ID20

var EmptyAddress = Address{}

func AddressFromStr(str string) (Address, error) {
	var id Address
	if err := addressFromStr(str, &id); err != nil {
		return id, errors.New("ids.AddressFromStr: " + err.Error())
	}

	return id, nil
}

func (id Address) Bytes() []byte {
	return id[:]
}

func (id Address) Ptr() *Address {
	return &id
}

func (id Address) Valid() bool {
	return true
}

func (id Address) String() string {
	return encoding.CheckSumHex(id[:])
}

func (id Address) GoString() string {
	return id.String()
}

func (id Address) AsKey() string {
	return string(id[:])
}

func (id Address) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *Address) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.Address.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	if err := addressFromStr(str, id); err != nil {
		return errors.New("ids.Address.UnmarshalText: " + err.Error())
	}

	return nil
}

func (id Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.String() + `"`), nil
}

func (id *Address) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.Address.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("ids.Address.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id Address) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.Address.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *Address) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.Address.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("ids.Address.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id Address) ToStakeSymbol() (s StakeSymbol) {
	switch {
	case id == EmptyAddress:
		// Empty
	case StakeSymbol(id).Valid():
		copy(s[:], id[:])
	default:
		s[0] = '#'
		h := encoding.Sum256(id[:])
		copy(s[1:], []byte(base32.StdEncoding.EncodeToString(h[:])))
	}
	return
}

func addressFromStr(str string, id *Address) error {
	str = strings.TrimPrefix(str, "0x")
	if str == "" {
		return nil
	}

	if l := hex.DecodedLen(len(str)); l != 20 {
		return errors.New("invalid bytes length, expected 20, got %d" +
			strconv.Itoa(l))
	}

	_, err := hex.Decode((*id)[:], []byte(str))
	return err
}
