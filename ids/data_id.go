// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ldclabs/ldvm/util/encoding"
)

type DataID ID32

var EmptyDataID = DataID{}

func DataIDFromStr(str string) (DataID, error) {
	var id DataID
	if str == "" {
		return id, nil
	}

	b, err := encoding.DecodeStringWithLen(str, 32)
	if err != nil {
		return id, errors.New("ids.DataIDFromStr: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

func (id DataID) Bytes() []byte {
	return id[:]
}

func (id DataID) Ptr() *DataID {
	return &id
}

func (id DataID) Valid() bool {
	return true
}

func (id DataID) String() string {
	return encoding.EncodeToString(id[:])
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
		return errors.New("ids.DataID.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	b, err := encoding.DecodeStringWithLen(str, 32)
	if err != nil {
		return errors.New("ids.DataID.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id DataID) MarshalJSON() ([]byte, error) {
	return []byte(encoding.EncodeToQuoteString(id[:])), nil
}

func (id *DataID) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.DataID.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := encoding.DecodeQuoteStringWithLen(str, 32)
	if err != nil {
		return errors.New("ids.DataID.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id DataID) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.DataID.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *DataID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.DataID.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 32)
	if err != nil {
		return errors.New("ids.DataID.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}
