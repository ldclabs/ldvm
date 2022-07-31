// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// This is a simple implementation of JSON-RPC 2.0.
// https://www.jsonrpc.org/specification

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeServerError    = -32000 // -32000 to -32099	Server error, Reserved for implementation-defined server-errors.
)

type Req struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Res struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   *Err            `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type Err struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *Err) Error() string {
	switch {
	case err.Data == nil:
		return fmt.Sprintf("%d, %s", err.Code, err.Message)
	default:
		data, _ := json.Marshal(err.Data)
		return fmt.Sprintf("%d, %s, data: %v", err.Code, err.Message, string(data))
	}
}

func DecodeReq(data []byte) (*Req, error) {
	req := &Req{}
	if err := json.Unmarshal(data, req); err != nil {
		return nil, &Err{Code: CodeParseError, Message: err.Error()}
	}
	if !req.isCall() {
		return nil, &Err{
			Code:    CodeInvalidRequest,
			Message: fmt.Sprintf("invalid request, %q", req.String())}
	}
	return req, nil
}

func (req *Req) isCall() bool {
	return req.Version == "2.0" && req.hasValidID() && req.Method != ""
}

func (req *Req) hasValidID() bool {
	return len(req.ID) > 0 && req.ID[0] != '{' && req.ID[0] != '['
}

func (req *Req) String() string {
	b, _ := json.Marshal(req)
	return string(b)
}

func (req *Req) DecodeParams(params interface{}) error {
	if err := json.Unmarshal(req.Params, params); err != nil {
		return &Err{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
			Data:    req.Params,
		}
	}
	return nil
}

func (req *Req) Result(result interface{}) *Res {
	data, err := json.Marshal(result)
	if err != nil {
		return req.Error(&Err{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Res{
		Version: req.Version,
		ID:      req.ID,
		Result:  data,
	}
}

func (req *Req) Error(err error) *Res {
	var rpcErr *Err

	switch v := err.(type) {
	case *Err:
		rpcErr = v

	default:
		rpcErr = &Err{
			Code:    CodeServerError,
			Message: err.Error(),
		}
	}

	return &Res{
		Version: req.Version,
		ID:      req.ID,
		Error:   rpcErr,
	}
}

func (req *Req) InvalidParams(msg string) *Res {
	return req.Error(&Err{
		Code:    CodeInvalidParams,
		Message: fmt.Sprintf("invalid parameter(s), %s", msg),
		Data:    req.Params,
	})
}

func (req *Req) InvalidMethod() *Res {
	return req.Error(&Err{
		Code:    CodeMethodNotFound,
		Message: fmt.Sprintf("method not found, %q", req.Method),
	})
}

func DecodeRes(data []byte) (*Res, error) {
	res := &Res{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, &Err{Code: CodeParseError, Message: err.Error()}
	}
	return res, nil
}

func (res *Res) DecodeResult(result interface{}) error {
	if res.Error != nil {
		return res.Error
	}

	if err := json.Unmarshal(res.Result, result); err != nil {
		return &Err{Code: CodeParseError, Message: err.Error()}
	}
	return nil
}

func (res *Res) String() string {
	b, _ := json.Marshal(res)
	return string(b)
}
