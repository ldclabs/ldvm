// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"errors"
	"strconv"

	"github.com/ldclabs/ldvm/util/encoding"
	"golang.org/x/crypto/sha3"
)

type ID32 [32]byte
type ID20 [20]byte

var EmptyID32 = ID32{}
var EmptyID20 = ID20{}

func ID32FromBytes(data []byte) (ID32, error) {
	var id ID32
	if bytesLen := len(data); bytesLen != 32 {
		return id, errors.New("ids.ID32FromBytes: invalid bytes length, expected 32, got " +
			strconv.Itoa(bytesLen))
	}

	copy(id[:], data)
	return id, nil
}

func ID32FromData(data []byte) ID32 {
	return ID32(sha3.Sum256(data))
}

func (id ID32) Bytes() []byte {
	return id[:]
}

func (id ID32) Ptr() *ID32 {
	return &id
}

func (id ID32) Valid() bool {
	return true
}

func (id ID32) String() string {
	return encoding.EncodeToString(id[:])
}

func (id ID32) GoString() string {
	return id.String()
}

func (id ID32) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ID32) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.ID32.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	b, err := encoding.DecodeStringWithLen(str, 32)
	if err != nil {
		return errors.New("ids.ID32.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ID32) MarshalJSON() ([]byte, error) {
	return []byte(encoding.EncodeToQuoteString(id[:])), nil
}

func (id *ID32) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.ID32.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := encoding.DecodeQuoteStringWithLen(str, 32)
	if err != nil {
		return errors.New("ids.ID32.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ID32) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.ID32.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *ID32) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.ID32.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 32)
	if err != nil {
		return errors.New("ids.ID32.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ID20) Bytes() []byte {
	return id[:]
}

func (id ID20) Ptr() *ID20 {
	return &id
}

func (id ID20) Valid() bool {
	return true
}

func (id ID20) String() string {
	return encoding.EncodeToString(id[:])
}

func (id ID20) GoString() string {
	return id.String()
}

func (id ID20) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ID20) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.ID20.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	b, err := encoding.DecodeStringWithLen(str, 20)
	if err != nil {
		return errors.New("ids.ID20.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ID20) MarshalJSON() ([]byte, error) {
	return []byte(encoding.EncodeToQuoteString(id[:])), nil
}

func (id *ID20) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.ID20.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := encoding.DecodeQuoteStringWithLen(str, 20)
	if err != nil {
		return errors.New("ids.ID20.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ID20) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.ID20.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *ID20) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.ID20.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("ids.ID20.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}
