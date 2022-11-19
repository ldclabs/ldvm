// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
)

type CBORService struct {
	host host.Host
	h    cborrpc.Handler
	name string
}

type CBORServiceOptions struct {
	ServiceName string
}

var DefaultCBORServiceOptions = CBORServiceOptions{
	ServiceName: "ldc:cborrpc",
}

func NewCBORService(host host.Host, h cborrpc.Handler, opts *CBORServiceOptions) *CBORService {
	if opts == nil {
		opts = &DefaultCBORServiceOptions
	}

	srv := &CBORService{host: host, h: h, name: opts.ServiceName}
	host.SetStreamHandler(CBORRPCProtocol, srv.handler)
	host.SetStreamHandler(CBORRPCGzipProtocol, srv.handler)
	return srv
}

func (srv *CBORService) handler(s network.Stream) {
	defer s.Close()

	ctx := context.Background()
	compress := s.Protocol() == CBORRPCGzipProtocol

	req := &cborrpc.Request{}
	if srv.name != "" {
		s.Scope().SetService(srv.name)
	}

	var err error
	var rd io.ReadCloser = s
	if compress {
		rd, err = gzip.NewReader(s)
		if err != nil {
			srv.writeCBOR(ctx, s, req.Error(err), compress)
			return
		}

		defer rd.Close()
	}

	_, err = req.ReadFrom(rd)
	if err != nil {
		srv.writeCBOR(ctx, s, req.Error(err), compress)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeCBOR(ctx, s, res, compress)
}

func (srv *CBORService) writeCBOR(
	ctx context.Context, w network.MuxedStream, res *cborrpc.Response, compress bool) {
	data, err := encoding.MarshalCBOR(res)
	if err != nil {
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError, Message: err.Error()}}
		data, err = encoding.MarshalCBOR(res)
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
