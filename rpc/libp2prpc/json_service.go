// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/compress"
)

type JSONService struct {
	host host.Host
	h    jsonrpc.Handler
	name string
}

type JSONServiceOptions struct {
	ServiceName string
}

var DefaultJSONServiceOptions = JSONServiceOptions{
	ServiceName: "ldc:jsonrpc",
}

func NewJSONService(host host.Host, h jsonrpc.Handler, opts *JSONServiceOptions) *JSONService {
	if opts == nil {
		opts = &DefaultJSONServiceOptions
	}

	srv := &JSONService{host: host, h: h, name: opts.ServiceName}
	host.SetStreamHandler(JSONRPCProtocol, srv.handler)
	host.SetStreamHandler(JSONRPCZstdProtocol, srv.handler)
	return srv
}

func (srv *JSONService) handler(s network.Stream) {
	defer s.Close()

	ctx := context.Background()
	cp := s.Protocol() == JSONRPCZstdProtocol

	req := &jsonrpc.Request{}
	if srv.name != "" {
		s.Scope().SetService(srv.name)
	}

	var rd io.ReadCloser = s
	var zr *compress.ZstdReader
	if cp {
		zr = &compress.ZstdReader{R: s}
		rd = zr
	}

	_, err := req.ReadFrom(rd)
	if zr != nil {
		zr.Reset()
	}
	if err != nil {
		srv.writeJSON(ctx, s, req.Error(err), cp)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeJSON(ctx, s, res, cp)
}

func (srv *JSONService) writeJSON(
	ctx context.Context, w network.MuxedStream, res *jsonrpc.Response, cp bool) {
	if res.Error != nil {
		srv.h.OnError(ctx, res.Error)
	}

	data, err := json.Marshal(res)
	if err != nil {
		res = &jsonrpc.Response{
			Version: "2.0",
			ID:      res.ID,
			Error:   &jsonrpc.Error{Code: jsonrpc.CodeServerError, Message: err.Error()}}
		data, err = json.Marshal(res)
	}

	if err != nil {
		srv.h.OnError(ctx, &jsonrpc.Error{
			Code:    jsonrpc.CodeInternalError,
			Message: fmt.Sprintf("impossible error, %v", err),
		})
	}

	er := res.Error
	if er != nil {
		srv.h.OnError(ctx, er)
	}

	switch {
	case cp:
		zw := &compress.ZstdWriter{W: w}
		_, err = zw.Write(data)
		zw.Reset()

	default:
		_, err = w.Write(data)
	}

	if err != nil {
		srv.h.OnError(ctx, &jsonrpc.Error{
			Code:    jsonrpc.CodeInternalError,
			Message: fmt.Sprintf("impossible error, %v", err),
		})
	}

	w.CloseWrite()
}
