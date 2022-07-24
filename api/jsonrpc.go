// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/json"
	"fmt"
)

type JsonrpcReq struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JsonrpcRes struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   *JsonrpcErr     `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type JsonrpcErr struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// -32000 to -32099	Server error	Reserved for implementation-defined server-errors.
const (
	AccountErrorCode = -32000
)

// NewJsonrpcRequest is a simple implementation of JSON-RPC 2.0.
// https://www.jsonrpc.org/specification
func NewJsonrpcRequest(data []byte) (*JsonrpcReq, error) {
	req := &JsonrpcReq{}
	if err := json.Unmarshal(data, req); err != nil {
		return nil, &JsonrpcErr{Code: -32700, Message: err.Error()}
	}
	if !req.isCall() {
		return nil, &JsonrpcErr{Code: -32600, Message: fmt.Sprintf("invalid request, %q", req.String())}
	}
	return req, nil
}

func (req *JsonrpcReq) isCall() bool {
	return req.Version == "2.0" && req.hasValidID() && req.Method != ""
}

func (req *JsonrpcReq) hasValidID() bool {
	return len(req.ID) > 0 && req.ID[0] != '{' && req.ID[0] != '['
}

func (req *JsonrpcReq) String() string {
	b, _ := json.Marshal(req)
	return string(b)
}

func (req *JsonrpcReq) ParseParams(params interface{}) error {
	if err := json.Unmarshal(req.Params, params); err != nil {
		return &JsonrpcErr{
			Code:    -32602,
			Message: fmt.Sprintf("invalid method parameter(s), %s", err),
			Data:    req.Params,
		}
	}
	return nil
}

func (req *JsonrpcReq) Result(result interface{}) *JsonrpcRes {
	data, err := json.Marshal(result)
	if err != nil {
		return req.Error(&JsonrpcErr{
			Code:    -32603,
			Message: fmt.Sprintf("JsonrpcReq.Result: internal error, %v", err),
		})
	}

	return &JsonrpcRes{
		Version: req.Version,
		ID:      req.ID,
		Result:  data,
	}
}

func (req *JsonrpcReq) Error(err error) *JsonrpcRes {
	var rpcErr *JsonrpcErr

	switch v := err.(type) {
	case *JsonrpcErr:
		rpcErr = v
	default:
		rpcErr = &JsonrpcErr{
			Code:    -32603,
			Message: fmt.Sprintf("JsonrpcReq.Error: internal error, %v", err),
		}
	}

	return &JsonrpcRes{
		Version: req.Version,
		ID:      req.ID,
		Error:   rpcErr,
	}
}

func (req *JsonrpcReq) InvalidParams(msg string) *JsonrpcRes {
	return req.Error(&JsonrpcErr{
		Code:    -32602,
		Message: fmt.Sprintf("invalid method parameter(s), %s", msg),
		Data:    req.Params,
	})
}

func (req *JsonrpcReq) InvalidMethod() *JsonrpcRes {
	return req.Error(&JsonrpcErr{
		Code:    -32601,
		Message: fmt.Sprintf("method not found, %q", req.Method),
	})
}

func (err *JsonrpcErr) Error() string {
	switch {
	case err.Data == nil:
		return fmt.Sprintf("%d, %s", err.Code, err.Message)
	default:
		return fmt.Sprintf("%d, %s, data: %v", err.Code, err.Message, err.Data)
	}
}
