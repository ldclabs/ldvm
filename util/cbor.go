// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"errors"

	"github.com/fxamacker/cbor/v2"
	cborpatch "github.com/ldclabs/cbor-patch"
)

var EncOpts = cbor.EncOptions{
	Sort:          cbor.SortLengthFirst,
	Time:          cbor.TimeRFC3339Nano,
	ShortestFloat: cbor.ShortestFloat16,
	NaNConvert:    cbor.NaNConvert7e00,
	InfConvert:    cbor.InfConvertFloat16,
	IndefLength:   cbor.IndefLengthForbidden,
	BigIntConvert: cbor.BigIntConvertNone,
}
var EncMode, _ = EncOpts.EncMode()
var DecOpts = cbor.DecOptions{
	DupMapKey:        cbor.DupMapKeyQuiet,
	IndefLength:      cbor.IndefLengthForbidden,
	MaxArrayElements: 100_000,
	MaxMapPairs:      1_000_000,
	// UTF8:             cbor.UTF8DecodeInvalid,
}
var DecMode, _ = DecOpts.DecMode()

func init() {
	cborpatch.SetCBOR(EncMode.Marshal, DecMode.Unmarshal)
}

func MarshalCBOR(v interface{}) ([]byte, error) {
	return EncMode.Marshal(v)
}

func MustMarshalCBOR(v interface{}) []byte {
	data, err := EncMode.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func UnmarshalCBOR(data []byte, v interface{}) error {
	return DecMode.Unmarshal(data, v)
}

func ValidCBOR(data []byte) error {
	return DecMode.Valid(data)
}

type RawData []byte

func (r RawData) MarshalJSON() ([]byte, error) {
	return MarshalJSONData(r), nil
}

func (r *RawData) UnmarshalJSON(b []byte) error {
	if r == nil {
		return errors.New("RawData: UnmarshalJSON on nil pointer")
	}
	data := UnmarshalJSONData(b)
	*r = append((*r)[0:0], data...)
	return nil
}
