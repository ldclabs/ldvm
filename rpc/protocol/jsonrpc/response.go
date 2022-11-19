// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"encoding/json"
	"errors"
	"io"
)

// This is a simple implementation of JSON-RPC 2.0.
// https://www.jsonrpc.org/specification

type Response struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func (res *Response) Grow(n int) {}

func (res *Response) ReadFrom(r io.Reader) (int64, error) {
	jd := json.NewDecoder(r)

	if err := jd.Decode(res); err != nil {
		return 0, err
	}

	n := jd.InputOffset()
	if jd.More() {
		return n, errors.New("json: unexpected following extraneous data")
	}

	return n, nil
}

func (res *Response) String() string {
	b, err := json.Marshal(res)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func (res *Response) DecodeResult(result interface{}) error {
	if res.Error != nil {
		return res.Error
	}

	if err := json.Unmarshal(res.Result, result); err != nil {
		return &Error{Code: CodeParseError, Message: err.Error()}
	}
	return nil
}
