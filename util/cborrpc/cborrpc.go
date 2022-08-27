// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/util"
)

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeServerError    = -32000 // -32000 to -32099	Server error, Reserved for implementation-defined server-errors.
)

const (
	MIMEApplicationCBOR            = "application/cbor"
	MIMEApplicationCBORCharsetUTF8 = "application/cbor; charset=utf-8"
)

type Req struct {
	Method string          `cbor:"method"`
	Params cbor.RawMessage `cbor:"params,omitempty"`
}

type Res struct {
	Error  *Err            `cbor:"error,omitempty"`
	Result cbor.RawMessage `cbor:"result,omitempty"`
}

type Err struct {
	Code    int         `cbor:"code"`
	Message string      `cbor:"message"`
	Data    interface{} `cbor:"data,omitempty"`
}

func (err *Err) Error() string {
	switch {
	case err.Data == nil:
		return fmt.Sprintf("%d, %s", err.Code, err.Message)

	default:
		switch err.Data.(type) {
		case cbor.RawMessage, []byte:
			return fmt.Sprintf("%d, %s, data: %x", err.Code, err.Message, err.Data)

		default:
			data, _ := util.MarshalCBOR(err.Data)
			return fmt.Sprintf("%d, %s, data: %x", err.Code, err.Message, data)
		}
	}
}

func DecodeReq(data []byte) (*Req, error) {
	req := &Req{}
	if err := util.UnmarshalCBOR(data, req); err != nil {
		return nil, &Err{Code: CodeParseError, Message: err.Error()}
	}
	if req.Method == "" {
		return nil, &Err{
			Code:    CodeInvalidRequest,
			Message: fmt.Sprintf("invalid request, method required")}
	}
	return req, nil
}

func (req *Req) String() string {
	return fmt.Sprintf("{method: %q, params: %x}", req.Method, req.Params)
}

func (req *Req) DecodeParams(params interface{}) error {
	if err := util.UnmarshalCBOR(req.Params, params); err != nil {
		return &Err{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
			Data:    req.Params,
		}
	}
	return nil
}

func (req *Req) Result(result interface{}) *Res {
	data, err := util.MarshalCBOR(result)
	if err != nil {
		return req.Error(&Err{
			Code:    CodeInternalError,
			Message: fmt.Sprintf("internal error, %v", err),
		})
	}

	return &Res{Result: data}
}

func (req *Req) ResultRaw(result cbor.RawMessage) *Res {
	return &Res{Result: result}
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

	return &Res{Error: rpcErr}
}

func (req *Req) InvalidParams(msg string) *Res {
	return req.Error(&Err{
		Code:    CodeInvalidParams,
		Message: fmt.Sprintf("invalid parameter(s), %s", msg),
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
	if err := util.UnmarshalCBOR(data, res); err != nil {
		return nil, &Err{Code: CodeParseError, Message: err.Error()}
	}
	return res, nil
}

func (res *Res) DecodeResult(result interface{}) error {
	if res.Error != nil {
		return res.Error
	}

	if err := util.UnmarshalCBOR(res.Result, result); err != nil {
		return &Err{Code: CodeParseError, Message: err.Error()}
	}
	return nil
}

func (res *Res) String() string {
	return fmt.Sprintf("{error: %v, result: %x}", res.Error, res.Result)
}
