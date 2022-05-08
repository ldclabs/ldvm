// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fxamacker/cbor/v2"
)

// EthID ==========
type EthID [20]byte

var EthIDEmpty = EthID{}

func EthIDFromString(str string) (EthID, error) {
	id := new(EthID)
	err := id.UnmarshalText([]byte(str))
	return *id, err
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
	if id == nil {
		return fmt.Errorf("EthID: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" {
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
	if id == nil {
		return fmt.Errorf("EthID: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("EthID: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id EthID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *EthID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("EthID: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("EthID: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

func (id EthID) ToStakeSymbol() (s StakeSymbol) {
	switch {
	case StakeSymbol(id).Valid():
		copy(s[:], id[:])
	default:
		s[0] = '@'
		h := sha3.Sum256(id[:])
		copy(s[1:], []byte(base32.StdEncoding.EncodeToString(h[:])))
	}
	return
}

// ModelID ==========
type ModelID [20]byte

var ModelIDEmpty = ModelID{}

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
	if id == nil {
		return fmt.Errorf("ModelID: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" {
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
	if id == nil {
		return fmt.Errorf("ModelID: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("ModelID: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id ModelID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *ModelID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("ModelID: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("ModelID: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

// DataID ==========
type DataID [20]byte

var DataIDEmpty = DataID{}

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
	if id == nil {
		return fmt.Errorf("DataID: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
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
	if id == nil {
		return fmt.Errorf("DataID: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("DataID: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id DataID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *DataID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("DataID: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("DataID: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

// TokenSymbol
type TokenSymbol [20]byte

var NativeToken = TokenSymbol{}

func NewToken(s string) (TokenSymbol, error) {
	symbol := TokenSymbol{}
	l := len(s)
	switch {
	case l == 0:
		return NativeToken, nil
	case l > 1 && l <= 10:
		copy(symbol[20-l:], []byte(s))
		if symbol.String() == s {
			return symbol, nil
		}
	}

	return symbol, fmt.Errorf("invalid token symbol")
}

func (id TokenSymbol) String() string {
	start := 0
	for i, r := range id[:] {
		switch {
		case r == 0:
			if start > 0 || i == 18 {
				return ""
			}
		case r >= 48 && r <= 57:
			if start == 0 {
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
	return string(id[start:])
}

func (id TokenSymbol) GoString() string {
	if id == NativeToken {
		return "NativeLDC"
	}
	if str := id.String(); str != "" {
		return str
	}
	return EthID(id).String()
}

func (id TokenSymbol) Valid() bool {
	return id == NativeToken || id.String() != ""
}

func (id TokenSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *TokenSymbol) UnmarshalText(b []byte) error {
	if id == nil {
		return fmt.Errorf("TokenSymbol: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	sid, err := NewToken(str)
	if err == nil {
		*id = sid
	}
	return err
}

func (id TokenSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *TokenSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return fmt.Errorf("TokenSymbol: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("TokenSymbol: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id TokenSymbol) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *TokenSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("TokenSymbol: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("TokenSymbol: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	if (*id) != NativeToken && id.String() == "" {
		return fmt.Errorf("TokenSymbol: UnmarshalCBOR on invalid TokenSymbol bytes: %s", EthID(*id))
	}
	return nil
}

// StakeSymbol
type StakeSymbol [20]byte

var StakeEmpty = StakeSymbol{}

func NewStake(s string) (StakeSymbol, error) {
	symbol := StakeSymbol{}
	l := len(s)
	switch {
	case l == 0:
		return StakeEmpty, nil
	case l > 1 && l <= 20:
		copy(symbol[20-l:], []byte(s))
		if symbol.String() == s {
			return symbol, nil
		}
	}

	return symbol, fmt.Errorf("invalid stake symbol")
}

func (id StakeSymbol) String() string {
	start := -1
	for i, r := range id[:] {
		switch {
		case r == 0:
			if start >= 0 || i == 18 {
				return ""
			}
		case r == '@':
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
	return EthID(id).String()
}

func (id StakeSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *StakeSymbol) UnmarshalText(b []byte) error {
	if id == nil {
		return fmt.Errorf("StakeSymbol: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	sid, err := NewStake(str)
	if err == nil {
		*id = sid
	}
	return err
}

func (id StakeSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *StakeSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return fmt.Errorf("StakeSymbol: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("StakeSymbol: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id StakeSymbol) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *StakeSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("StakeSymbol: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return fmt.Errorf("StakeSymbol: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	if (*id) != StakeEmpty && id.String() == "" {
		return fmt.Errorf("StakeSymbol: UnmarshalCBOR on invalid StakeSymbol bytes: %s", EthID(*id))
	}
	return nil
}

func EthIDToStakeSymbol(ids ...EthID) []StakeSymbol {
	rt := make([]StakeSymbol, 0, len(ids))
	for _, id := range ids {
		rt = append(rt, id.ToStakeSymbol())
	}
	return rt
}
