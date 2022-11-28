// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/rs/xid"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/compress"
	"github.com/ldclabs/ldvm/util/encoding"
)

const (
	CBORRPCProtocol     protocol.ID = "/cborrpc/v1"
	CBORRPCZstdProtocol protocol.ID = "/cborrpc/v1zstd"
)

type CBORClient struct {
	host     host.Host
	endpoint peer.ID
	compress bool
}

type CBORClientOptions struct {
	Compress bool
}

var DefaultCBORClientOptions = CBORClientOptions{
	Compress: true,
}

func NewCBORClient(host host.Host, endpoint peer.ID, opts *CBORClientOptions) *CBORClient {
	if opts == nil {
		opts = &DefaultCBORClientOptions
	}

	return &CBORClient{
		host:     host,
		endpoint: endpoint,
		compress: opts.Compress,
	}
}

func (c *CBORClient) Request(ctx context.Context, method string, params, result any) *cborrpc.Response {
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

	proto := CBORRPCProtocol
	if c.compress {
		proto = CBORRPCZstdProtocol
	}

	s, err := c.host.NewStream(ctx, c.endpoint, proto)
	if err != nil {
		return req.Error(fmt.Errorf("creating stream to %s, %v", c.endpoint, err))
	}
	defer s.Close()

	switch {
	case c.compress:
		zw := compress.NewZstdWriter(s)
		_, err = zw.Write(data)
		zw.Reset()

	default:
		_, err = s.Write(data)
	}

	res := &cborrpc.Response{ID: req.ID}
	if err != nil {
		return req.Error(fmt.Errorf("write data failed, %v", err))
	}
	s.CloseWrite()

	var rd io.ReadCloser = s
	if c.compress {
		zr := compress.NewZstdReader(s)
		rd = zr
		defer zr.Reset() // `defer s.Close()` exists above, just reset the zstd reader
	}

	_, err = res.ReadFrom(rd)
	if err != nil {
		return req.Error(err)
	}

	return res
}
