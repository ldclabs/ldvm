// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encoding

import (
	"encoding/json"
	"errors"
)

func MarshalJSONData(data []byte) json.RawMessage {
	switch {
	case len(data) == 0 || json.Valid(data):
		return data

	default:
		return []byte(EncodeToQuoteString(data))
	}
}

func UnmarshalJSONData(data json.RawMessage) []byte {
	if d, err := DecodeQuoteString(string(data)); err == nil {
		return d
	}

	return data
}

type RawData []byte

func (r RawData) MarshalJSON() ([]byte, error) {
	return MarshalJSONData(r), nil
}

func (r *RawData) UnmarshalJSON(b []byte) error {
	if r == nil {
		return errors.New("encoding.RawData.UnmarshalJSON: nil pointer")
	}

	data := UnmarshalJSONData(b)
	*r = append((*r)[0:0], data...)
	return nil
}

func MustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
