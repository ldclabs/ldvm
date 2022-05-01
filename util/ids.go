// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

var NativeToken = TokenSymbol(ids.ShortEmpty)

type TokenSymbol ids.ShortID

func NewSymbol(s string) (TokenSymbol, error) {
	if l := len(s); l > 10 || l < 2 {
		return TokenSymbol{}, fmt.Errorf("invalid token symbol")
	}
	symbol := TokenSymbol{}
	copy(symbol[20-len(s):], []byte(s))
	if symbol.String() == "" {
		return TokenSymbol{}, fmt.Errorf("invalid token symbol")
	}
	return symbol, nil
}

func (s TokenSymbol) String() string {
	start := 0
	for i, r := range s[:] {
		switch {
		case r == 0:
			if start > 0 || i == 18 {
				return ""
			}
		case r >= 65 && r <= 90:
			if i < 10 {
				return ""
			}
			if start == 0 {
				start = i
			}
		default:
			return ""
		}
	}
	return string(s[start:])
}

// EthID ==========
type EthID ids.ShortID

var EthIDEmpty = EthID(ids.ShortEmpty)

func EthIDFromString(str string) (EthID, error) {
	id := new(EthID)
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

func EthIDsToShort(eids ...EthID) []ids.ShortID {
	rt := make([]ids.ShortID, len(eids))
	for i, id := range eids {
		rt[i] = id.ShortID()
	}
	return rt
}

func (id EthID) ShortID() ids.ShortID {
	return ids.ShortID(id)
}

func (id EthID) String() string {
	return common.Address(id).Hex()
}

func (id EthID) GoString() string {
	return id.String()
}

func (id EthID) MarshalText() ([]byte, error) {
	return []byte(common.Address(id).Hex()), nil
}

func (id *EthID) UnmarshalText(b []byte) error {
	str := string(b)
	if str == "" { // If "null", do nothing
		*id = EthIDEmpty
		return nil
	}
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}

	var err error
	var sid ids.ShortID
	switch {
	case len(str) == 40:
		if b, err = hex.DecodeString(str); err == nil {
			sid, err = ids.ToShortID(b)
		}
	default:
		sid, err = ids.ShortFromString(str)
	}

	if err == nil {
		*id = EthID(sid)
	}
	return err
}

func (id EthID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + common.Address(id).Hex() + "\""), nil
}

func (id *EthID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("invalid ID string: %s", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

// ModelID ==========
type ModelID ids.ShortID

var ModelIDEmpty = ModelID(ids.ShortEmpty)

func ModelIDFromString(str string) (ModelID, error) {
	if str == "" {
		return ModelIDEmpty, nil
	}
	id, err := ids.ShortFromPrefixedString(str, "LM")
	if err != nil {
		return ModelIDEmpty, err
	}
	return ModelID(id), nil
}

func (id ModelID) ShortID() ids.ShortID {
	return ids.ShortID(id)
}

func (id ModelID) String() string {
	return ids.ShortID(id).PrefixedString("LM")
}

func (id ModelID) GoString() string {
	return id.String()
}

func (id ModelID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ModelID) UnmarshalText(b []byte) error {
	str := string(b)
	if str == "" { // If "null", do nothing
		*id = ModelIDEmpty
		return nil
	}

	sid, err := ModelIDFromString(str)
	if err == nil {
		*id = sid
	}
	return err
}

func (id ModelID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *ModelID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("invalid ID string: %s", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

// DataID ==========
type DataID ids.ShortID

var DataIDEmpty = DataID(ids.ShortEmpty)

func DataIDFromString(str string) (DataID, error) {
	if str == "" {
		return DataIDEmpty, nil
	}
	id, err := ids.ShortFromPrefixedString(str, "LD")
	if err != nil {
		return DataIDEmpty, err
	}
	return DataID(id), nil
}

func (id DataID) ShortID() ids.ShortID {
	return ids.ShortID(id)
}

func (id DataID) String() string {
	return ids.ShortID(id).PrefixedString("LD")
}

func (id DataID) GoString() string {
	return id.String()
}

func (id DataID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *DataID) UnmarshalText(b []byte) error {
	str := string(b)
	if str == "" { // If "null", do nothing
		*id = DataIDEmpty
		return nil
	}

	sid, err := DataIDFromString(str)
	if err == nil {
		*id = sid
	}
	return err
}

func (id DataID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *DataID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("invalid ID string: %s", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}
