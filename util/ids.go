// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/base32"
	"encoding/hex"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
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

func (id EthID) AsKey() string {
	return id.String() // TODO: string(id[:])
}

func (id EthID) MarshalText() ([]byte, error) {
	return []byte(common.Address(id).Hex()), nil
}

func (id *EthID) UnmarshalText(b []byte) error {
	errp := ErrPrefix("EthID.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
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
	return errp.ErrorIf(err)
}

func (id EthID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + common.Address(id).Hex() + "\""), nil
}

func (id *EthID) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("EthID.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %q", str)
	}

	str = str[1:lastIndex]
	return errp.ErrorIf(id.UnmarshalText([]byte(str)))
}

func (id EthID) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("EthID.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *EthID) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("EthID.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := UnmarshalCBOR(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return errp.Errorf("invalid bytes length, expected 20, got %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

func (id EthID) ToStakeSymbol() (s StakeSymbol) {
	switch {
	case id == EthIDEmpty:
		// Empty
	case StakeSymbol(id).Valid():
		copy(s[:], id[:])
	default:
		s[0] = '#'
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
		return ModelIDEmpty, ErrPrefix("ModelIDFromString error: ").ErrorIf(err)
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
	errp := ErrPrefix("ModelID.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	sid, err := ModelIDFromString(str)
	if err == nil {
		*id = sid
	}
	return errp.ErrorIf(err)
}

func (id ModelID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *ModelID) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("ModelID.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %q", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id ModelID) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("ModelID.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *ModelID) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("ModelID.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := UnmarshalCBOR(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return errp.Errorf("invalid bytes length, expected 20, got %d", len(b))
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
		return DataIDEmpty, ErrPrefix("DataIDFromString error: ").ErrorIf(err)
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
	errp := ErrPrefix("DataID.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	sid, err := DataIDFromString(str)
	if err == nil {
		*id = sid
	}
	return errp.ErrorIf(err)
}

func (id DataID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *DataID) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("DataID.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %q", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id DataID) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("DataID.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *DataID) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("DataID.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := UnmarshalCBOR(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return errp.Errorf("invalid bytes length, expected 20, got %d", len(b))
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
	case l > 1 && l <= 20:
		copy(symbol[20-l:], []byte(s))
		if symbol.String() == s {
			return symbol, nil
		}
	}

	return symbol, ErrPrefix("NewToken error: ").Errorf("invalid token symbol %q", s)
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
	return EthID(id).GoString()
}

func (id TokenSymbol) AsKey() string {
	if id == NativeToken {
		return ""
	}
	if str := id.String(); str != "" {
		return str
	}
	return EthID(id).AsKey()
}

func (id TokenSymbol) EthID() EthID {
	return EthID(id)
}

func (id TokenSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *TokenSymbol) UnmarshalText(b []byte) error {
	errp := ErrPrefix("TokenSymbol.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	sid, err := NewToken(str)
	if err == nil {
		*id = sid
	}
	return errp.ErrorIf(err)
}

func (id TokenSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *TokenSymbol) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("TokenSymbol.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %q", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id TokenSymbol) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("TokenSymbol.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *TokenSymbol) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("TokenSymbol.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := UnmarshalCBOR(data, &b); err != nil {
		return err
	}
	if len(b) != 20 {
		return errp.Errorf("invalid bytes length, expected 20, got %d", len(b))
	}
	copy((*id)[:], b)
	if !id.Valid() {
		return errp.Errorf("invalid TokenSymbol: %s", id.GoString())
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

	return symbol, ErrPrefix("NewStake error: ").Errorf("invalid stake symbol")
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
	return EthID(id).String()
}

func (id StakeSymbol) EthID() EthID {
	return EthID(id)
}

func (id StakeSymbol) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *StakeSymbol) UnmarshalText(b []byte) error {
	errp := ErrPrefix("StakeSymbol.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	sid, err := NewStake(str)
	if err == nil {
		*id = sid
	}
	return errp.ErrorIf(err)
}

func (id StakeSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *StakeSymbol) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("StakeSymbol.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %q", str)
	}

	str = str[1:lastIndex]
	return id.UnmarshalText([]byte(str))
}

func (id StakeSymbol) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("StakeSymbol.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *StakeSymbol) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("StakeSymbol.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := UnmarshalCBOR(data, &b); err != nil {
		return errp.ErrorIf(err)
	}
	if len(b) != 20 {
		return errp.Errorf("invalid bytes length, expected 20, got %d", len(b))
	}
	copy((*id)[:], b)
	if *id != StakeEmpty && !id.Valid() {
		return errp.Errorf("invalid StakeSymbol: %s", id.GoString())
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
