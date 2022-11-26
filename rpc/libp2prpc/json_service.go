// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/compress"
	"github.com/ldclabs/ldvm/util/value"
)

type JSONService struct {
	ctx     context.Context
	host    host.Host
	h       jsonrpc.Handler
	name    string
	timeout time.Duration
}

type JSONServiceOptions struct {
	ServiceName string
	Timeout     time.Duration
}

var DefaultJSONServiceOptions = JSONServiceOptions{
	ServiceName: "ldc:jsonrpc",
	Timeout:     5 * time.Second,
}

func NewJSONService(ctx context.Context, host host.Host, h jsonrpc.Handler, opts *JSONServiceOptions) *JSONService {
	if opts == nil {
		opts = &DefaultJSONServiceOptions
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultCBORServiceOptions.Timeout
	}

	srv := &JSONService{ctx: ctx, host: host, h: h, name: opts.ServiceName, timeout: opts.Timeout}
	host.SetStreamHandler(JSONRPCProtocol, srv.handler)
	host.SetStreamHandler(JSONRPCZstdProtocol, srv.handler)
	return srv
}

func (srv *JSONService) handler(s network.Stream) {
	defer s.Close()

	ctx, cancel := context.WithTimeout(srv.ctx, srv.timeout)
	defer cancel()

	log := value.Log{Value: value.NewMap(16)}
	ctx = value.CtxWith(ctx, &log)

	start := time.Now()
	log.Set("start", value.Time(start))
	defer func() {
		log.Set("elapsed", value.Int64(int64(time.Since(start)/time.Millisecond)))
	}()

	log.Set("service", value.String(srv.name))
	log.Set("protocol", value.String(string(s.Protocol())))
	log.Set("streamId", value.String(s.ID()))

	conn := s.Conn()
	log.Set("localPeer", value.String(conn.LocalPeer().String()))
	log.Set("remotePeer", value.String(conn.RemotePeer().String()))
	log.Set("remoteAddr", value.String(conn.RemoteMultiaddr().String()))
	log.Set("connStreams", value.Int(conn.Stat().NumStreams))

	cp := s.Protocol() == JSONRPCZstdProtocol

	req := &jsonrpc.Request{}
	if srv.name != "" {
		s.Scope().SetService(srv.name)
	}

	var rd io.ReadCloser = s
	var zr *compress.ZstdReader
	if cp {
		zr = compress.NewZstdReader(s)
		rd = zr
	}

	n, err := req.ReadFrom(rd)
	if zr != nil {
		zr.Reset()
	}
	log.Set("requestBytes", value.Int64(n))
	if err != nil {
		srv.writeJSON(ctx, s, req.Error(err), cp)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeJSON(ctx, s, res, cp)
}

func (srv *JSONService) writeJSON(
	ctx context.Context, w network.MuxedStream, res *jsonrpc.Response, cp bool) {

	data, err := json.Marshal(res)
	var exception *jsonrpc.Response
	if err != nil {
		exception = &jsonrpc.Response{
			Version: "2.0",
			ID:      res.ID,
			Error:   &jsonrpc.Error{Code: jsonrpc.CodeServerError, Message: err.Error()}}
		data, err = json.Marshal(res)
	}

	if err != nil {
		exception.Error.Message = fmt.Sprintf("impossible error, %v", err)
	}

	switch {
	case cp:
		zw := compress.NewZstdWriter(w)
		zw.Write(data)
		zw.Reset()

	default:
		w.Write(data)
	}

	w.CloseWrite()

	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("responseBytes", value.Int(len(data)))

		if res.Error != nil {
			log.Set("responseError", value.String(res.Error.Error()))
		}

		if exception != nil {
			log.Set("error", value.String(exception.Error.Error()))
		}
	})
}
