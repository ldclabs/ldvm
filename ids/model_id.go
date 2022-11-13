// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"errors"

	"github.com/ldclabs/ldvm/util/encoding"
)

type ModelID ID20

var EmptyModelID = ModelID{}

func ModelIDFromStr(str string) (ModelID, error) {
	var id ModelID
	if str == "" {
		return id, nil
	}

	b, err := encoding.DecodeStringWithLen(str, 20)
	if err != nil {
		return id, errors.New("ids.ModelIDFromStr: " + err.Error())
	}

	copy(id[:], b)
	return id, nil
}

func ModelIDFromHash(id ID32) ModelID {
	var modelID ModelID
	copy(modelID[:], id[:])
	return modelID
}

func (id ModelID) Bytes() []byte {
	return id[:]
}

func (id ModelID) Ptr() *ModelID {
	return &id
}

func (id ModelID) Valid() bool {
	return true
}

func (id ModelID) String() string {
	return encoding.EncodeToString(id[:])
}

func (id ModelID) GoString() string {
	return id.String()
}

func (id ModelID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ModelID) UnmarshalText(b []byte) error {
	if id == nil {
		return errors.New("ids.ModelID.UnmarshalText: nil pointer")
	}

	str := string(b)
	if str == "" {
		return nil
	}

	b, err := encoding.DecodeStringWithLen(str, 20)
	if err != nil {
		return errors.New("ids.ModelID.UnmarshalText: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ModelID) MarshalJSON() ([]byte, error) {
	return []byte(encoding.EncodeToQuoteString(id[:])), nil
}

func (id *ModelID) UnmarshalJSON(b []byte) error {
	if id == nil {
		return errors.New("ids.ModelID.UnmarshalJSON: nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}

	b, err := encoding.DecodeQuoteStringWithLen(str, 20)
	if err != nil {
		return errors.New("ids.ModelID.UnmarshalJSON: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}

func (id ModelID) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(id[:])
	if err != nil {
		return nil, errors.New("ids.ModelID.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (id *ModelID) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return errors.New("ids.ModelID.UnmarshalCBOR: nil pointer")
	}

	b, err := encoding.UnmarshalCBORWithLen(data, 20)
	if err != nil {
		return errors.New("ids.ModelID.UnmarshalCBOR: " + err.Error())
	}

	copy((*id)[:], b)
	return nil
}
