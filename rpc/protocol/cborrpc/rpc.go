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
	// ServeRPC handles a CBOR-RPC request.
	ServeRPC(context.Context, *Request) *Response
}

// Result returns a response from the request with the given result.
func (req *Request) Result(result any) *Response {
	data, err := encoding.MarshalCBOR(result)
	if err != nil {
		return req.Error(&Error{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Response{ID: req.ID, Result: data}
}

// Result returns a response from the request with the given raw result.
func (req *Request) ResultRaw(result cbor.RawMessage) *Response {
	return &Response{ID: req.ID, Result: result}
}

// Error returns a response from the request with the given error.
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

// InvalidParams returns a invalid params response from the request with the given message.
func (req *Request) InvalidParams(msg string) *Response {
	return req.Error(&Error{
		Code:    CodeInvalidParams,
		Message: fmt.Sprintf("invalid parameter(s), %s", msg),
	})
}

// InvalidMethod returns a invalid method response from the request.
func (req *Request) InvalidMethod() *Response {
	return req.Error(&Error{
		Code:    CodeMethodNotFound,
		Message: fmt.Sprintf("method %q not found", req.Method),
	})
}

// ToJSONRaw try to convert a CBOR RawMessage to JSON RawMessage.
// If the CBOR RawMessage is invalid, it returns an error.
// If conversion fails, it returns a base64&checksum JSON RawMessage converting by encoding.EncodeToQuoteString.
func ToJSONRaw(data cborpatch.RawMessage) (json.RawMessage, error) {
	if len(data) == 0 {
		return []byte(data), nil
	}

	if err := encoding.ValidCBOR(data); err != nil {
		return nil, err
	}
	if data, err := cborpatch.ToJSON(data, nil); err == nil {
		return data, nil
	}
	return []byte(encoding.EncodeToQuoteString(data)), nil
}
