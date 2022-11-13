// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/rpc/httpcli"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/rs/xid"
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
			data, _ := encoding.MarshalCBOR(err.Data)
			return fmt.Sprintf("%d, %s, data: %x", err.Code, err.Message, data)
		}
	}
}

func DecodeReq(data []byte) (*Req, error) {
	req := &Req{}
	if err := encoding.UnmarshalCBOR(data, req); err != nil {
		return nil, &Err{Code: CodeParseError, Message: err.Error()}
	}
	if req.Method == "" {
		return nil, &Err{
			Code:    CodeInvalidRequest,
			Message: "invalid request, method required"}
	}
	return req, nil
}

func (req *Req) String() string {
	return fmt.Sprintf("{method: %q, params: %x}", req.Method, req.Params)
}

func (req *Req) DecodeParams(params interface{}) error {
	if err := encoding.UnmarshalCBOR(req.Params, params); err != nil {
		return &Err{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
			Data:    req.Params,
		}
	}
	return nil
}

func (req *Req) Result(result interface{}) *Res {
	data, err := encoding.MarshalCBOR(result)
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

func DecodeRes(data []byte) *Res {
	res := &Res{}
	if err := encoding.UnmarshalCBOR(data, res); err != nil {
		res.Error = &Err{Code: CodeParseError, Message: err.Error()}
	}
	return res
}

func (res *Res) DecodeResult(result interface{}) error {
	if res.Error == nil {
		if err := encoding.UnmarshalCBOR(res.Result, result); err != nil {
			res.Error = &Err{Code: CodeParseError, Message: err.Error()}
		}
	}
	return res.Error
}

func (res *Res) String() string {
	return fmt.Sprintf("{error: %v, result: %x}", res.Error, res.Result)
}

type Client struct {
	cli      *httpcli.Client
	endpoint string
	header   http.Header
}

func NewClient(endpoint string, rt http.RoundTripper, header http.Header) *Client {
	return &Client{
		cli:      httpcli.NewClient(rt),
		endpoint: endpoint,
		header:   header,
	}
}

func (c *Client) Do(ctx context.Context, req *Req) *Res {
	err := ctx.Err()
	if err != nil {
		return req.Error(fmt.Errorf("context.Context error: %v", err))
	}

	if req == nil || req.Method == "" {
		return req.Error(fmt.Errorf("invalid request"))
	}

	data, err := encoding.MarshalCBOR(req)
	if err != nil {
		return req.Error(err)
	}

	r, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return req.Error(err)
	}

	r.Header.Set("accept", MIMEApplicationCBOR)
	r.Header.Set("content-type", MIMEApplicationCBORCharsetUTF8)
	r.Header.Set("x-request-id", xid.New().String())
	httpcli.CopyHeader(r.Header, c.header)
	httpcli.CopyHeader(r.Header, httpcli.HeaderFromCtx(ctx))

	data, err = c.cli.Do(r)
	if err != nil {
		return req.Error(err)
	}

	return DecodeRes(data)
}

func (c *Client) Req(ctx context.Context, method string, params, result interface{}) *Res {
	data, err := encoding.MarshalCBOR(params)
	if err != nil {
		return &Res{Error: &Err{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("MarshalCBOR error, %v", err),
		}}
	}

	res := c.Do(ctx, &Req{Method: method, Params: data})
	if result != nil {
		res.DecodeResult(result)
	}

	return res
}
