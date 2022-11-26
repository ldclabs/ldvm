// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// This is a simple implementation of JSON-RPC 2.0.
// https://www.jsonrpc.org/specification

type Request struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (req *Request) ReadFrom(r io.Reader) (int64, error) {
	jd := json.NewDecoder(r)

	if err := jd.Decode(req); err != nil {
		return 0, err
	}

	n := jd.InputOffset()
	if jd.More() {
		return n, errors.New("json: unexpected following extraneous data")
	}

	if !req.isCall() {
		return n, fmt.Errorf("invalid request, %q", req.String())
	}

	return n, nil
}

func (req *Request) isCall() bool {
	return req.Version == "2.0" && req.hasValidID() && req.Method != ""
}

func (req *Request) hasValidID() bool {
	return len(req.ID) > 0 && req.ID[0] != '{' && req.ID[0] != '['
}

func (req *Request) String() string {
	b, err := json.Marshal(req)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func (req *Request) DecodeParams(params interface{}) error {
	if err := json.Unmarshal(req.Params, params); err != nil {
		return &Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
		}
	}
	return nil
}

func (req *Request) Result(result interface{}) *Response {
	data, err := json.Marshal(result)
	if err != nil {
		return req.Error(&Error{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Response{Version: req.Version, ID: req.ID, Result: data}
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

	return &Response{Version: req.Version, ID: req.ID, Error: rpcErr}
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
