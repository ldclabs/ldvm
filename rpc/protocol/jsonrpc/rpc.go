// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = "application/json; charset=utf-8"
)

type Handler interface {
	// ServeRPC handles a JSON-RPC request.
	ServeRPC(context.Context, *Request) *Response
}

// Result returns a response from the request with the given result.
func (req *Request) Result(result any) *Response {
	data, err := json.Marshal(result)
	if err != nil {
		return req.Error(&Error{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Response{Version: req.Version, ID: req.ID, Result: data}
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

	return &Response{Version: req.Version, ID: req.ID, Error: rpcErr}
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
