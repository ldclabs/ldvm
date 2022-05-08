// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"errors"

	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/util"
)

// from CoreDetEncOptions()
var EncOpts = cbor.EncOptions{
	Sort:          cbor.SortLengthFirst,
	Time:          cbor.TimeUnix,
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
	MaxArrayElements: 10_000,
	MaxMapPairs:      1_000_000,
}
var DecMode, _ = DecOpts.DecMode()

type RawData []byte

func (r RawData) MarshalJSON() ([]byte, error) {
	return util.JSONMarshalData(r), nil
}

func (r *RawData) UnmarshalJSON(b []byte) error {
	if r == nil {
		return errors.New("RawData: UnmarshalJSON on nil pointer")
	}
	data := util.JSONUnmarshalData(b)
	*r = append((*r)[0:0], data...)
	return nil
}

func (r RawData) MarshalCBOR() ([]byte, error) {
	switch {
	case len(r) == 0:
		return []byte{0xf6}, nil
	case DecMode.Valid(r) == nil:
		return r, nil
	}
	return EncMode.Marshal([]byte(r))
}

func (r *RawData) UnmarshalCBOR(data []byte) error {
	if r == nil {
		return errors.New("RawData: UnmarshalCBOR on nil pointer")
	}

	*r = append((*r)[0:0], data...)
	return nil
}
