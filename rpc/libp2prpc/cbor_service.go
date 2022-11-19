// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/compress"
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
	host.SetStreamHandler(CBORRPCZstdProtocol, srv.handler)
	return srv
}

func (srv *CBORService) handler(s network.Stream) {
	defer s.Close()

	ctx := context.Background()
	cp := s.Protocol() == CBORRPCZstdProtocol

	req := &cborrpc.Request{}
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
		srv.writeCBOR(ctx, s, req.Error(err), cp)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeCBOR(ctx, s, res, cp)
}

func (srv *CBORService) writeCBOR(
	ctx context.Context, w network.MuxedStream, res *cborrpc.Response, cp bool) {
	if res.Error != nil {
		srv.h.OnError(ctx, res.Error)
	}

	data, err := encoding.MarshalCBOR(res)
	if err != nil {
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError, Message: err.Error()}}
		data, err = encoding.MarshalCBOR(res)
	}

	if err != nil {
		srv.h.OnError(ctx, &cborrpc.Error{
			Code:    cborrpc.CodeInternalError,
			Message: fmt.Sprintf("impossible error, %v", err),
		})
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
		srv.h.OnError(ctx, &cborrpc.Error{
			Code:    cborrpc.CodeInternalError,
			Message: fmt.Sprintf("impossible error, %v", err),
		})
	}

	w.CloseWrite()
}
