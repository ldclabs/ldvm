// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
	cborpatch "github.com/ldclabs/cbor-patch"

	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/httpcli"
)

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification
type Request struct {
	ID     string          `cbor:"id,omitempty"`
	Method string          `cbor:"method"`
	Params cbor.RawMessage `cbor:"params,omitempty"`

	buf bytes.Buffer `cbor:"-"`
}

func (req *Request) Grow(n int) {
	req.buf.Grow(n)
}

func (req *Request) ReadFrom(r io.Reader) (int64, error) {
	n, err := req.buf.ReadFrom(r)
	if err != nil {
		return n, err
	}

	if err = encoding.UnmarshalCBOR(req.buf.Bytes(), req); err != nil {
		return n, err
	}

	return n, nil
}

func (req *Request) String() string {
	return fmt.Sprintf(`{"id":%q,"method":%q,"params":"%x"}`,
		req.ID, req.Method, encoding.EncodeToQuoteString(req.Params))
}

func (req *Request) DecodeParams(params interface{}) error {
	if err := encoding.UnmarshalCBOR(req.Params, params); err != nil {
		return &Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
		}
	}
	return nil
}

func (req *Request) Result(result interface{}) *Response {
	data, err := encoding.MarshalCBOR(result)
	if err != nil {
		return req.Error(&Error{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Response{ID: req.ID, Result: data}
}

func (req *Request) ResultRaw(result cbor.RawMessage) *Response {
	return &Response{ID: req.ID, Result: result}
}

func (req *Request) Error(err error) *Response {
	var rpcErr *Error

	switch v := err.(type) {
	case *Error:
		rpcErr = v

	case *httpcli.Error:
		rpcErr = &Error{
			Code:    CodeServerError - v.Code,
			Message: v.Message,
			Data: map[string]interface{}{
				"body":   v.Body,
				"header": v.Header,
			},
		}

	default:
		rpcErr = &Error{
			Code:    CodeServerError,
			Message: err.Error(),
		}
	}

	return &Response{ID: req.ID, Error: rpcErr}
}

func (req *Request) InvalidParams(msg string) *Response {
	return req.Error(&Error{
		Code:    CodeInvalidParams,
		Message: fmt.Sprintf("invalid parameter(s), %s", msg),
	})
}

func (req *Request) InvalidMethod() *Response {
	return req.Error(&Error{
		Code:    CodeMethodNotFound,
		Message: fmt.Sprintf("method %q not found", req.Method),
	})
}

func ToJSON(data cborpatch.RawMessage) json.RawMessage {
	if data, err := cborpatch.ToJSON(data, nil); err == nil {
		return data
	}
	return []byte(encoding.EncodeToQuoteString(data))
}
