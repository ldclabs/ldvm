// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
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
	host.SetStreamHandler(JSONRPCGzipProtocol, srv.handler)
	return srv
}

func (srv *JSONService) handler(s network.Stream) {
	defer s.Close()

	ctx := context.Background()
	compress := s.Protocol() == JSONRPCGzipProtocol

	req := &jsonrpc.Request{}
	if srv.name != "" {
		s.Scope().SetService(srv.name)
	}

	var err error
	var rd io.ReadCloser = s
	if compress {
		rd, err = gzip.NewReader(s)
		if err != nil {
			srv.writeJSON(ctx, s, req.Error(err), compress)
			return
		}

		defer rd.Close()
	}

	_, err = req.ReadFrom(rd)
	if err != nil {
		srv.writeJSON(ctx, s, req.Error(err), compress)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeJSON(ctx, s, res, compress)
}

func (srv *JSONService) writeJSON(
	ctx context.Context, w network.MuxedStream, res *jsonrpc.Response, compress bool) {
	data, err := json.Marshal(res)
	if err != nil {
		res = &jsonrpc.Response{
			Version: "2.0",
			ID:      res.ID,
			Error:   &jsonrpc.Error{Code: jsonrpc.CodeServerError, Message: err.Error()}}
		data, err = json.Marshal(res)
	}

	if compress {
		data, err = tryGzip(data)
	}

	if err != nil {
		res.Error.Message = fmt.Sprintf("impossible error, %v", err)
	}

	if res.Error != nil {
		srv.h.OnError(ctx, res.Error)
	}

	w.Write(data)
	w.CloseWrite()
}
