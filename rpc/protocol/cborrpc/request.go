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

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification

// Request represents a CBOR-RPC request.
type Request struct {
	ID     string          `cbor:"id"`
	Method string          `cbor:"method"`
	Params cbor.RawMessage `cbor:"params,omitempty"`

	buf bytes.Buffer `cbor:"-"`
}

// Grow grows the underlying buffer's capacity
func (req *Request) Grow(n int) {
	req.buf.Grow(n)
}

// ReadFrom decodes the request from the given reader.
// ReadFrom implements io.ReaderFrom interface.
func (req *Request) ReadFrom(r io.Reader) (int64, error) {
	n, err := req.buf.ReadFrom(r)
	if err != nil {
		return n, err
	}

	if err = encoding.UnmarshalCBOR(req.buf.Bytes(), req); err != nil {
		return n, err
	}

	if req.ID == "" || req.Method == "" {
		return n, fmt.Errorf("invalid request, %q", req.String())
	}

	return n, nil
}

type jsonRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// String returns the string representation of the request.
func (req *Request) String() string {
	b, err := ToJSONRaw(req.Params)
	if err == nil {
		b, err = json.Marshal(jsonRequest{
			ID:     req.ID,
			Method: req.Method,
			Params: b,
		})
	}
	if err != nil {
		b, _ = json.Marshal(erring.RespondError{Err: err.Error()})
	}

	return string(b)
}

// DecodeParams decodes the request parameters into the given value.
func (req *Request) DecodeParams(params any) error {
	if err := encoding.UnmarshalCBOR(req.Params, params); err != nil {
		return &Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
		}
	}
	return nil
}
