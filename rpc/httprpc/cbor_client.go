// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"bytes"
	"context"
	"net/http"

	"github.com/rs/xid"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/httpcli"
)

type CBORClient struct {
	cli      *httpcli.Client
	header   http.Header
	endpoint string
}

type CBORClientOptions struct {
	RoundTripper http.RoundTripper
	Header       http.Header
}

var DefaultCBORClientOptions = CBORClientOptions{
	RoundTripper: httpcli.DefaultTransport,
}

func NewCBORClient(endpoint string, opts *CBORClientOptions) *CBORClient {
	if opts == nil {
		opts = &DefaultCBORClientOptions
	}
	return &CBORClient{
		cli:      httpcli.NewClient(opts.RoundTripper),
		endpoint: endpoint,
		header:   opts.Header,
	}
}

func (c *CBORClient) Request(
	ctx context.Context, method string, params, result any) *cborrpc.Response {
	var err error

	req := &cborrpc.Request{ID: xid.New().String(), Method: method}
	if params != nil {
		req.Params, err = encoding.MarshalCBOR(params)
		if err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}
	}

	res := c.Do(ctx, req)
	if result != nil {
		res.DecodeResult(result)
	}

	return res
}

func (c *CBORClient) Do(ctx context.Context, req *cborrpc.Request) *cborrpc.Response {
	err := ctx.Err()
	if err != nil {
		return req.Error(err)
	}

	if req.Method == "" {
		return req.InvalidMethod()
	}

	if req.ID == "" {
		req.ID = xid.New().String()
	}

	data, err := encoding.MarshalCBOR(req)
	if err != nil {
		return req.Error(err)
	}

	r, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return req.Error(err)
	}

	r.Header.Set("accept", cborrpc.MIMEApplicationCBOR)
	r.Header.Set("content-type", cborrpc.MIMEApplicationCBOR)
	r.Header.Set("x-request-id", req.ID)
	httpcli.CopyHeader(r.Header, c.header)
	httpcli.CopyHeader(r.Header, httpcli.HeaderCtxValue(ctx))

	res := &cborrpc.Response{ID: req.ID}
	err = c.cli.DoWithReader(r, res)
	if err != nil {
		return req.Error(err)
	}

	return res
}
