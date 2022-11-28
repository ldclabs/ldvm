// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

// refer to JSON-RPC 2.0. https://www.jsonrpc.org/specification
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	// -32000 to -32599	Server error, Reserved for implementation-defined server-errors.
	CodeServerError = -32000
)

// Error is a CBOR-RPC error.
type Error = erring.Error

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification
// Response represents a CBOR-RPC response.
type Response struct {
	ID     string          `cbor:"id"`
	Error  *Error          `cbor:"error,omitempty"`
	Result cbor.RawMessage `cbor:"result,omitempty"`

	buf bytes.Buffer `cbor:"-"`
}

// Grow grows the underlying buffer's capacity
func (res *Response) Grow(n int) {
	res.buf.Grow(n)
}

// ReadFrom decodes the CBOR-RPC response from the given reader.
// ReadFrom implements io.ReaderFrom interface.
func (res *Response) ReadFrom(r io.Reader) (int64, error) {
	n, err := res.buf.ReadFrom(r)
	if err != nil {
		return n, err
	}

	if err = encoding.UnmarshalCBOR(res.buf.Bytes(), res); err != nil {
		return n, err
	}

	if res.ID == "" || (res.Error == nil && len(res.Result) == 0) {
		return n, fmt.Errorf("invalid response, %q", res.String())
	}

	return n, nil
}

type jsonResponse struct {
	ID     string          `json:"id"`
	Error  *Error          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

// String returns the string representation of the response.
func (res *Response) String() string {
	var err error
	var b []byte

	if res.Error == nil {
		b, err = ToJSONRaw(res.Result)
	}
	if err == nil {
		b, err = json.Marshal(jsonResponse{
			ID:     res.ID,
			Error:  res.Error,
			Result: b,
		})
	}
	if err != nil {
		b, _ = json.Marshal(erring.RespondError{Err: err.Error()})
	}

	return string(b)
}

// DecodeResult decodes the result into the given value.
func (res *Response) DecodeResult(result interface{}) error {
	if res.Error != nil {
		return res.Error
	}

	if err := encoding.UnmarshalCBOR(res.Result, result); err != nil {
		return &Error{Code: CodeParseError, Message: err.Error()}
	}
	return nil
}
