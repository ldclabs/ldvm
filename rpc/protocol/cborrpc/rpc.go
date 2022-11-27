// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/util/encoding"
)

const (
	MIMEApplicationCBOR = "application/cbor"
)

type Handler interface {
	ServeRPC(context.Context, *Request) *Response
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

	default:
		if !errors.As(err, &rpcErr) {
			rpcErr = &Error{
				Code:    CodeServerError,
				Message: err.Error(),
			}
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
