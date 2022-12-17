// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encoding

import (
	"errors"
	"strconv"

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
	DupMapKey:        cbor.DupMapKeyEnforcedAPF,
	IndefLength:      cbor.IndefLengthForbidden,
	MaxArrayElements: 100_000,
	MaxMapPairs:      1_000_000,
	UTF8:             cbor.UTF8DecodeInvalid,
}
var DecMode, _ = DecOpts.DecMode()

func init() {
	cborpatch.SetCBOR(EncMode.Marshal, DecMode.Unmarshal)
}

func MarshalCBOR(v any) ([]byte, error) {
	return EncMode.Marshal(v)
}

func MustMarshalCBOR(v any) []byte {
	data, err := EncMode.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func UnmarshalCBOR(data []byte, v any) error {
	return DecMode.Unmarshal(data, v)
}

func UnmarshalCBORWithLen(data []byte, expectedLen int) ([]byte, error) {
	var buf []byte
	if err := DecMode.Unmarshal(data, &buf); err != nil {
		return nil, errors.New("encoding.UnmarshalCBORWithLen: " + err.Error())
	}

	if bytesLen := len(buf); bytesLen != expectedLen {
		return nil, errors.New("encoding.UnmarshalCBORWithLen: invalid bytes length, expected " +
			strconv.Itoa(expectedLen) + ", got " + strconv.Itoa(bytesLen))
	}
	return buf, nil
}

func ValidCBOR(data []byte) error {
	return DecMode.Valid(data)
}
