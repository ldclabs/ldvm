// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"

// 	"github.com/libp2p/go-libp2p/core/host"
// 	"github.com/libp2p/go-libp2p/core/peer"
// 	"github.com/libp2p/go-libp2p/core/protocol"
// 	"github.com/rs/xid"

// 	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
// 	"github.com/ldclabs/ldvm/util/compress"
// )

// const (
// 	JSONRPCProtocol     protocol.ID = "/jsonrpc/v1"
// 	JSONRPCZstdProtocol protocol.ID = "/jsonrpc/v1zstd"
// )

// type JSONClient struct {
// 	host     host.Host
// 	endpoint peer.ID
// 	compress bool
// }

// type JSONClientOptions struct {
// 	Compress bool
// }

// var DefaultJSONClientOptions = JSONClientOptions{
// 	Compress: true,
// }

// func NewJSONClient(host host.Host, endpoint peer.ID, opts *JSONClientOptions) *JSONClient {
// 	if opts == nil {
// 		opts = &DefaultJSONClientOptions
// 	}

// 	return &JSONClient{
// 		host:     host,
// 		endpoint: endpoint,
// 		compress: opts.Compress,
// 	}
// }

// func (c *JSONClient) Request(
// 	ctx context.Context, method string, params, result any) *jsonrpc.Response {
// 	var err error

// 	req := &jsonrpc.Request{Version: "2.0", ID: xid.New().String(), Method: method}
// 	if params != nil {
// 		req.Params, err = json.Marshal(params)
// 		if err != nil {
// 			return req.Error(&jsonrpc.Error{
// 				Code:    jsonrpc.CodeInvalidParams,
// 				Message: err.Error(),
// 			})
// 		}
// 	}

// 	res := c.Do(ctx, req)
// 	if result != nil {
// 		res.DecodeResult(result)
// 	}

// 	return res
// }

// func (c *JSONClient) Do(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
// 	err := ctx.Err()
// 	if err != nil {
// 		return req.Error(err)
// 	}

// 	if req.Method == "" {
// 		return req.InvalidMethod()
// 	}

// 	if req.Version == "" {
// 		req.Version = "2.0"
// 	}

// 	if req.ID == "" {
// 		req.ID = xid.New().String()
// 	}

// 	data, err := json.Marshal(req)
// 	if err != nil {
// 		return req.Error(err)
// 	}

// 	proto := JSONRPCProtocol
// 	if c.compress {
// 		proto = JSONRPCZstdProtocol
// 	}

// 	s, err := c.host.NewStream(ctx, c.endpoint, proto)
// 	if err != nil {
// 		return req.Error(fmt.Errorf("creating stream to %s, %v", c.endpoint, err))
// 	}
// 	defer s.Close()

// 	switch {
// 	case c.compress:
// 		zw := compress.NewZstdWriter(s)
// 		_, err = zw.Write(data)
// 		zw.Reset()

// 	default:
// 		_, err = s.Write(data)
// 	}

// 	res := &jsonrpc.Response{Version: "2.0", ID: req.ID}
// 	if err != nil {
// 		return req.Error(fmt.Errorf("write data failed, %v", err))
// 	}
// 	s.CloseWrite()

// 	var rd io.ReadCloser = s
// 	if c.compress {
// 		zr := compress.NewZstdReader(s)
// 		rd = zr
// 		defer zr.Reset() // `defer s.Close()` exists above, just reset the zstd reader
// 	}

// 	_, err = res.ReadFrom(rd)
// 	if err != nil {
// 		return req.Error(err)
// 	}

// 	return res
// }
