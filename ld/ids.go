// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ethereum/go-ethereum/common"
)

// EthID ==========
type EthID ids.ShortID

var EthIDEmpty = EthID(ids.ShortEmpty)

func EthIDFromString(str string) (EthID, error) {
	id := new(EthID)
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

func (id EthID) String() string {
	return common.Address(id).Hex()
}

func (id EthID) GoString() string {
	return id.GoString()
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

func ModelIDFromData(inputs ...[]byte) ModelID {
	s := 0
	for _, d := range inputs {
		s += len(d)
	}
	data := make([]byte, s)
	s = 0
	for _, d := range inputs {
		n := copy(data[s:], d)
		s += n
	}
	return ModelID(hashing.ComputeHash160Array(data))
}

func ModelIDFromString(str string) (ModelID, error) {
	if str == "" {
		return ModelIDEmpty, nil
	}
	id, err := ids.ShortFromPrefixedString(str, "M")
	if err != nil {
		return ModelIDEmpty, err
	}
	return ModelID(id), nil
}

func (id ModelID) String() string {
	return ids.ShortID(id).PrefixedString("M")
}

func (id ModelID) GoString() string {
	return id.GoString()
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

func DataIDFromData(inputs ...[]byte) DataID {
	s := 0
	for _, d := range inputs {
		s += len(d)
	}
	data := make([]byte, s)
	s = 0
	for _, d := range inputs {
		n := copy(data[s:], d)
		s += n
	}
	return DataID(hashing.ComputeHash160Array(data))
}

func DataIDFromString(str string) (DataID, error) {
	if str == "" {
		return DataIDEmpty, nil
	}
	id, err := ids.ShortFromPrefixedString(str, "D")
	if err != nil {
		return DataIDEmpty, err
	}
	return DataID(id), nil
}

func (id DataID) String() string {
	return ids.ShortID(id).PrefixedString("D")
}

func (id DataID) GoString() string {
	return id.GoString()
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
