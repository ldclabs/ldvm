// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/database"
)

// Address ==========
type Address [20]byte

var AddressEmpty = Address{}

func AddressFrom(str string) (Address, error) {
	var id Address

	str = strings.TrimPrefix(str, "0x")
	if str == "" {
		return id, nil
	}

	b, err := hex.DecodeString(str)
	if err != nil {
		return id, errors.New("util.AddressFrom: " + err.Error())
	}
	if len(b) != 20 {
		return id, errors.New("util.AddressFrom: invalid bytes length, expected 20, got %d" +
			strconv.Itoa(len(b)))
	}

	copy(id[:], b)
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
	return CheckSumHex(id[:])
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
		return errors.New("util.Address.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	addr, err := AddressFrom(str)
	if err != nil {
		return errors.New("util.Address.UnmarshalText: " + err.Error())
	}

	*id = addr
	return nil
}

func (id Address) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *Address) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.Address.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("util.Address.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id Address) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.Address.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *Address) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.Address.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("util.Address.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id Address) ToStakeSymbol() (s StakeSymbol) {
	switch {
	case id == AddressEmpty:
		// Empty
	case StakeSymbol(id).Valid():
		copy(s[:], id[:])
	default:
		s[0] = '#'
		h := Sum256(id[:])
		copy(s[1:], []byte(base32.StdEncoding.EncodeToString(h[:])))
	}
	return
}

// DataID ==========
type DataID [32]byte

var DataIDEmpty = DataID{}

func DataIDFrom(str string) (DataID, error) {
	var id DataID
	if str == "" {
		return id, nil
	}

	b, err := DecodeStringWithLen(str, 32)
	if err != nil {
		return id, errors.New("util.DataIDFrom: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

func (id DataID) Bytes() []byte {
	return id[:]
}

func (id DataID) Valid() bool {
	return true
}

func (id DataID) String() string {
	return EncodeToString(id[:])
}

func (id DataID) GoString() string {
	return id.String()
}

func (id DataID) VersionKey(version uint64) []byte {
	v := database.PackUInt64(version)
	key := make([]byte, 32+len(v))
	copy(key, id[:])
	copy(key[32:], v)
	return key
}

func (id DataID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *DataID) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("util.DataID.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	b, err := DecodeStringWithLen(str, 32)
	if err != nil {
		return errors.New("util.DataID.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id DataID) MarshalJSON() ([]byte, error) {
	return []byte(EncodeToQuoteString(id[:])), nil
}

func (id *DataID) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.DataID.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := DecodeQuoteStringWithLen(str, 32)
	if err != nil {
		return errors.New("util.DataID.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id DataID) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.DataID.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *DataID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.DataID.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 32)
	if err != nil {
		return errors.New("util.DataID.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

// Hash ==========
type Hash [32]byte

var HashEmpty = Hash{}

func HashFrom(str string) (Hash, error) {
	var id Hash
	if str == "" {
		return id, nil
	}

	b, err := DecodeStringWithLen(str, 32)
	if err != nil {
		return id, errors.New("util.HashFrom: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

func HashFromBytes(data []byte) (Hash, error) {
	var id Hash
	if bytesLen := len(data); bytesLen != 32 {
		return id, errors.New("util.HashFromBytes: invalid bytes length, expected 32, got " +
			strconv.Itoa(bytesLen))
	}

	copy(id[:], data)
	return id, nil
}

func (id Hash) Bytes() []byte {
	return id[:]
}

func (id Hash) Valid() bool {
	return true
}

func (id Hash) String() string {
	return EncodeToString(id[:])
}

func (id Hash) GoString() string {
	return id.String()
}

func (id Hash) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *Hash) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("util.Hash.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	b, err := DecodeStringWithLen(str, 32)
	if err != nil {
		return errors.New("util.Hash.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id Hash) MarshalJSON() ([]byte, error) {
	return []byte(EncodeToQuoteString(id[:])), nil
}

func (id *Hash) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.Hash.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := DecodeQuoteStringWithLen(str, 32)
	if err != nil {
		return errors.New("util.Hash.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id Hash) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.Hash.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *Hash) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.Hash.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 32)
	if err != nil {
		return errors.New("util.Hash.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

// ModelID ==========
type ModelID [20]byte

var ModelIDEmpty = ModelID{}

func ModelIDFrom(str string) (ModelID, error) {
	var id ModelID
	if str == "" {
		return id, nil
	}

	b, err := DecodeStringWithLen(str, 20)
	if err != nil {
		return id, errors.New("util.ModelIDFrom: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

func ModelIDFromHash(id Hash) ModelID {
	var modelID ModelID
	copy(modelID[:], id[:])
	return modelID
}

func (id ModelID) Bytes() []byte {
	return id[:]
}

func (id ModelID) Valid() bool {
	return true
}

func (id ModelID) String() string {
	return EncodeToString(id[:])
}

func (id ModelID) GoString() string {
	return id.String()
}

func (id ModelID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ModelID) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("util.ModelID.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	b, err := DecodeStringWithLen(str, 20)
	if err != nil {
		return errors.New("util.ModelID.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ModelID) MarshalJSON() ([]byte, error) {
	return []byte(EncodeToQuoteString(id[:])), nil
}

func (id *ModelID) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.ModelID.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := DecodeQuoteStringWithLen(str, 20)
	if err != nil {
		return errors.New("util.ModelID.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ModelID) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.ModelID.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *ModelID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.ModelID.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("util.ModelID.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

// TokenSymbol
type TokenSymbol [20]byte

var NativeToken = TokenSymbol{}

func TokenFrom(s string) (TokenSymbol, error) {
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

	return symbol, errors.New("util.TokenFrom: invalid token symbol %q" + strconv.Quote(s))
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

func (id TokenSymbol) AsKey() string {
	if id == NativeToken {
		return ""
	}
	if str := id.String(); str != "" {
		return str
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
		return errors.New("uitl.TokenSymbol.UnmarshalText: nil pointer")
	}

	sid, err := TokenFrom(string(b))
	if err != nil {
		return errors.New("uitl.TokenSymbol.UnmarshalText: " + err.Error())
	}

	*id = sid
	return nil
}

func (id TokenSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *TokenSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.TokenSymbol.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("util.TokenSymbol.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id TokenSymbol) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.TokenSymbol.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *TokenSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.TokenSymbol.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("util.TokenSymbol.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	if !id.Valid() {
		return errors.New("util.TokenSymbol.UnmarshalCBOR: invalid TokenSymbol " + id.GoString())
	}
	return nil
}

// StakeSymbol
type StakeSymbol [20]byte

var StakeEmpty = StakeSymbol{}

func StakeFrom(s string) (StakeSymbol, error) {
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

	return symbol, errors.New("util.StakeFrom: invalid stake symbol " + strconv.Quote(s))
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
		return errors.New("uitl.StakeSymbol.UnmarshalText: nil pointer")
	}

	sid, err := StakeFrom(string(b))
	if err != nil {
		return errors.New("uitl.StakeSymbol.UnmarshalText: " + err.Error())
	}

	*id = sid
	return nil
}

func (id StakeSymbol) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *StakeSymbol) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("util.StakeSymbol.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return errors.New("util.StakeSymbol.UnmarshalJSON: invalid input")
	}

	return id.UnmarshalText([]byte(str[1 : strLen-1]))
}

func (id StakeSymbol) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("util.StakeSymbol.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *StakeSymbol) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("util.StakeSymbol.UnmarshalCBOR: nil pointer")
	}

	b, err := UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("util.StakeSymbol.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	if *id != StakeEmpty && !id.Valid() {
		return errors.New("util.StakeSymbol.UnmarshalCBOR: invalid StakeSymbol " + id.GoString())
	}
	return nil
}
