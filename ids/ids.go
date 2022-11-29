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

func ID32FromStr(str string) (ID32, error) {
	var id ID32
	if str == "" {
		return id, nil
	}

	b, err := encoding.DecodeStringWithLen(str, 32)
	if err != nil {
		return id, errors.New("ids.ID32FromStr: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

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

// type ID []byte

// var (
// 	emptyID20b [20]byte
// 	emptyID32b [32]byte

// 	emptyID20 = string(emptyID20b[:])
// 	emptyID32 = string(emptyID32b[:])
// )

// func (id ID) Valid() bool {
// 	switch len(id) {
// 	case 20:
// 		return string(id) != emptyID20

// 	case 32:
// 		return string(id) != emptyID32

// 	default:
// 		return false
// 	}
// }

// func (id ID) ValidOrEmpty() bool {
// 	switch {
// 	case len(id) == 0:
// 		return true

// 	default:
// 		return id.Valid()
// 	}
// }

// func (id ID) Bytes() []byte {
// 	return id
// }

// func (id ID) Ptr() *ID {
// 	return &id
// }

// func (id ID) Address() (v Address) {
// 	copy(v[:], id)
// 	return
// }

// func (id ID) ID20() (v ID20) {
// 	copy(v[:], id)
// 	return
// }

// func (id ID) ID32() (v ID32) {
// 	copy(v[:], id)
// 	return
// }

// func (id ID) String() string {
// 	if len(id) == 0 {
// 		return ""
// 	}
// 	return encoding.EncodeToString(id)
// }

// func (id ID) GoString() string {
// 	return id.String()
// }

// func (id ID) MarshalText() ([]byte, error) {
// 	return []byte(id.String()), nil
// }

// func (id *ID) UnmarshalText(b []byte) error {
// 	if id == nil {
// 		return errors.New("ids.ID.UnmarshalText: nil pointer")
// 	}

// 	str := string(b)
// 	if str == "" {
// 		return nil
// 	}

// 	b, err := encoding.DecodeString(str)
// 	if err != nil {
// 		return errors.New("ids.ID.UnmarshalText: " + err.Error())
// 	}

// 	if len(b) != 20 && len(b) != 32 {
// 		return errors.New("ids.ID.UnmarshalText: invalid length " + strconv.Itoa(len(b)))
// 	}

// 	copy((*id)[:], b)
// 	return nil
// }

// func (id ID) MarshalJSON() ([]byte, error) {
// 	if len(id) == 0 {
// 		return []byte(""), nil
// 	}

// 	return []byte(encoding.EncodeToQuoteString(id)), nil
// }

// func (id *ID) UnmarshalJSON(b []byte) error {
// 	if id == nil {
// 		return errors.New("ids.ID.UnmarshalJSON: nil pointer")
// 	}

// 	str := string(b)
// 	if str == "null" || str == `""` { // If "null", do nothing
// 		return nil
// 	}

// 	b, err := encoding.DecodeQuoteString(str)
// 	if err != nil {
// 		return errors.New("ids.ID.UnmarshalJSON: " + err.Error())
// 	}

// 	if len(b) != 20 && len(b) != 32 {
// 		return errors.New("ids.ID.UnmarshalText: invalid length " + strconv.Itoa(len(b)))
// 	}

// 	copy((*id)[:], b)
// 	return nil
// }

// func (id ID) MarshalCBOR() ([]byte, error) {
// 	data, err := encoding.MarshalCBOR(id)
// 	if err != nil {
// 		return nil, errors.New("ids.ID.MarshalCBOR: " + err.Error())
// 	}
// 	return data, nil
// }

// func (id *ID) UnmarshalCBOR(data []byte) error {
// 	if id == nil {
// 		return errors.New("ids.ID.UnmarshalCBOR: nil pointer")
// 	}

// 	var b []byte
// 	if err := encoding.UnmarshalCBOR(data, &b); err != nil {
// 		return errors.New("ids.ID20.UnmarshalCBOR: " + err.Error())
// 	}

// 	if len(b) == 0 {
// 		return nil
// 	}

// 	if len(b) != 20 && len(b) != 32 {
// 		return errors.New("ids.ID.UnmarshalCBOR: invalid length " + strconv.Itoa(len(b)))
// 	}

// 	copy((*id)[:], b)
// 	return nil
// }
