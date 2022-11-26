// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
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

type Error = erring.Error

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification
type Response struct {
	ID     string          `cbor:"id,omitempty"`
	Error  *Error          `cbor:"error,omitempty"`
	Result cbor.RawMessage `cbor:"result,omitempty"`

	buf bytes.Buffer `cbor:"-"`
}

func (res *Response) Grow(n int) {
	res.buf.Grow(n)
}

func (res *Response) ReadFrom(r io.Reader) (int64, error) {
	n, err := res.buf.ReadFrom(r)
	if err != nil {
		return n, err
	}

	if err = encoding.UnmarshalCBOR(res.buf.Bytes(), res); err != nil {
		return n, err
	}

	return n, nil
}

func (res *Response) String() string {
	if res.Error != nil {
		return fmt.Sprintf(`{"id":%q,"error":%q}`, res.ID, res.Error.Error())
	}

	return fmt.Sprintf(`{"id":%q,"result":%s}`, res.ID, string(ToJSON(res.Result)))
}

func (res *Response) DecodeResult(result interface{}) error {
	if res.Error == nil {
		if err := encoding.UnmarshalCBOR(res.Result, result); err != nil {
			res.Error = &Error{Code: CodeParseError, Message: err.Error()}
		}
	}
	return res.Error
}
