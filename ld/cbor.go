// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"errors"

	cbor "github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/util"
)

// from CoreDetEncOptions()
var EncOpts = cbor.EncOptions{
	Sort:          cbor.SortBytewiseLexical,
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

type RawData cbor.RawMessage

func (r RawData) MarshalJSON() ([]byte, error) {
	return util.JSONMarshalData(r), nil
}

func (r *RawData) UnmarshalJSON(b []byte) error {
	if r == nil {
		return errors.New("ld.RawData: UnmarshalJSON on nil pointer")
	}
	data := util.JSONUnmarshalData(b)
	*r = make([]byte, len(data))
	copy(*r, data)
	return nil
}
