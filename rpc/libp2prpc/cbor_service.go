// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
)

type CBORService struct {
	host host.Host
	h    cborrpc.Handler
}

func NewCBORService(host host.Host, h cborrpc.Handler) *CBORService {
	srv := &CBORService{host: host, h: h}
	host.SetStreamHandler(ProtocolID, srv.handler)
	return srv
}

func (srv *CBORService) handler(s network.Stream) {
	defer s.Close()

	ctx := context.Background()
	req := &cborrpc.Request{}
	if err := s.Scope().SetService(ServiceName); err != nil {
		srv.writeCBOR(ctx, s, req.Error(err))
		return
	}

	_, err := req.ReadFrom(s)
	if err != nil {
		srv.writeCBOR(ctx, s, req.Error(err))
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeCBOR(ctx, s, res)
}

func (srv *CBORService) writeCBOR(ctx context.Context, w network.MuxedStream, res *cborrpc.Response) {
	data, err := encoding.MarshalCBOR(res)
	if err != nil {
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError, Message: err.Error()}}
		data, err = encoding.MarshalCBOR(res)
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
