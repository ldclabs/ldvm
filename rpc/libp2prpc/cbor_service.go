// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/compress"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/value"
)

type CBORService struct {
	ctx       context.Context
	host      host.Host
	h         cborrpc.Handler
	name      string
	timeout   time.Duration
	handleLog func(log *value.Log)
}

type CBORServiceOptions struct {
	ServiceName string
	Timeout     time.Duration
	HandleLog   func(log *value.Log)
}

var DefaultCBORServiceOptions = CBORServiceOptions{
	ServiceName: "ldc:cborrpc",
	Timeout:     5 * time.Second,
	HandleLog:   value.DefaultLogHandler,
}

func NewCBORService(
	ctx context.Context, host host.Host, h cborrpc.Handler, opts *CBORServiceOptions) *CBORService {
	if opts == nil {
		opts = &DefaultCBORServiceOptions
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultCBORServiceOptions.Timeout
	}
	if opts.HandleLog == nil {
		opts.HandleLog = DefaultCBORServiceOptions.HandleLog
	}

	srv := &CBORService{
		ctx: ctx, host: host, h: h,
		name:      opts.ServiceName,
		timeout:   opts.Timeout,
		handleLog: opts.HandleLog,
	}
	host.SetStreamHandler(CBORRPCProtocol, srv.handler)
	host.SetStreamHandler(CBORRPCZstdProtocol, srv.handler)
	return srv
}

func (srv *CBORService) handler(s network.Stream) {
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

	cp := s.Protocol() == CBORRPCZstdProtocol

	req := &cborrpc.Request{}
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
		srv.writeCBOR(ctx, s, req.Error(err), cp)
		return
	}

	res := srv.h.ServeRPC(ctx, req)
	srv.writeCBOR(ctx, s, res, cp)
	go srv.handleLog(&log)
}

func (srv *CBORService) writeCBOR(
	ctx context.Context, w network.MuxedStream, res *cborrpc.Response, cp bool) {
	exception := &erring.Error{}
	if res.Error != nil && res.Error.HasErrs() {
		exception = res.Error
	}

	data, err := encoding.MarshalCBOR(res)
	if exception.CatchIf(err) {
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError, Message: err.Error()},
		}

		data, err = encoding.MarshalCBOR(res)
		if exception.CatchIf(err) {
			data, _ = encoding.MarshalCBOR(erring.RespondError{
				Err: fmt.Sprintf("impossible error: %q", err.Error()),
			})
		}
	}

	switch {
	case cp:
		zw := compress.NewZstdWriter(w)
		_, err = zw.Write(data)
		zw.Reset()

	default:
		_, err = w.Write(data)
	}

	exception.CatchIf(err)
	exception.CatchIf(w.CloseWrite())

	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("responseBytes", value.Int(len(data)))

		if res.Error != nil {
			log.Set("responseError", value.String(res.Error.Error()))
		}

		if exception.HasErrs() {
			log.Err = exception
		}
	})
}
