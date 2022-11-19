// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/rs/xid"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
)

const (
	CBORRPCProtocol     protocol.ID = "/cborrpc/v1"
	CBORRPCGzipProtocol protocol.ID = "/cborrpc/v1gzip"
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

func (c *CBORClient) Request(ctx context.Context, method string, params, result interface{}) *cborrpc.Response {
	var err error

	req := &cborrpc.Request{ID: xid.New().String(), Method: method}
	req.Params, err = encoding.MarshalCBOR(params)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error(),
		})
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
		proto = CBORRPCGzipProtocol
		data, err = tryGzip(data)
		if err != nil {
			return req.Error(err)
		}
	}

	s, err := c.host.NewStream(ctx, c.endpoint, proto)
	if err != nil {
		return req.Error(fmt.Errorf("creating stream to %s, %v", c.endpoint, err))
	}
	defer s.Close()

	res := &cborrpc.Response{ID: req.ID}
	_, err = s.Write(data)
	if err != nil {
		return req.Error(fmt.Errorf("write data failed, %v", err))
	}
	s.CloseWrite()

	var rd io.ReadCloser = s
	if c.compress {
		rd, err = gzip.NewReader(s)
		if err != nil {
			return req.Error(err)
		}
		defer rd.Close()
	}

	_, err = res.ReadFrom(rd)
	if err != nil {
		return req.Error(err)
	}

	return res
}

func tryGzip(data []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	gw := gzip.NewWriter(b)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
