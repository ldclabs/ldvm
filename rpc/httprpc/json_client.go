// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/xid"

	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/httpcli"
)

type JSONClient struct {
	cli      *httpcli.Client
	header   http.Header
	endpoint string
}

type JSONClientOptions struct {
	RoundTripper http.RoundTripper
	Header       http.Header
}

var DefaultJSONClientOptions = JSONClientOptions{
	RoundTripper: httpcli.DefaultTransport,
}

func NewJSONClient(endpoint string, opts *JSONClientOptions) *JSONClient {
	if opts == nil {
		opts = &DefaultJSONClientOptions
	}
	return &JSONClient{
		cli:      httpcli.NewClient(opts.RoundTripper),
		endpoint: endpoint,
		header:   opts.Header,
	}
}

func (c *JSONClient) Request(
	ctx context.Context, method string, params, result interface{}) *jsonrpc.Response {
	var err error

	req := &jsonrpc.Request{Version: "2.0", ID: xid.New().String(), Method: method}
	req.Params, err = json.Marshal(params)
	if err != nil {
		return req.Error(&jsonrpc.Error{
			Code:    jsonrpc.CodeInvalidParams,
			Message: err.Error(),
		})
	}

	res := c.Do(ctx, req)
	if result != nil {
		res.DecodeResult(result)
	}

	return res
}

func (c *JSONClient) Do(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	err := ctx.Err()
	if err != nil {
		return req.Error(err)
	}

	if req.Method == "" {
		return req.InvalidMethod()
	}

	if req.Version == "" {
		req.Version = "2.0"
	}

	if req.ID == "" {
		req.ID = xid.New().String()
	}

	data, err := json.Marshal(req)
	if err != nil {
		return req.Error(err)
	}

	r, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return req.Error(err)
	}

	r.Header.Set("accept", jsonrpc.MIMEApplicationJSON)
	r.Header.Set("content-type", jsonrpc.MIMEApplicationJSONCharsetUTF8)
	r.Header.Set("x-request-id", req.ID)
	httpcli.CopyHeader(r.Header, c.header)
	httpcli.CopyHeader(r.Header, httpcli.HeaderCtxValue(ctx))

	res := &jsonrpc.Response{Version: "2.0", ID: req.ID}
	err = c.cli.DoWithReader(r, res)
	if err != nil {
		return req.Error(err)
	}

	return res
}
