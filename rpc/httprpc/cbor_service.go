// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/gzip"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/compress"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/httpcli"
	"github.com/ldclabs/ldvm/util/value"
)

const (
	compressionThreshold = gzhttp.DefaultMinSize
)

type CBORService struct {
	h              cborrpc.Handler
	name           string
	loggingHeaders []string
}

type CBORServiceOptions struct {
	ServiceName    string
	LoggingHeaders []string
}

var DefaultCBORServiceOptions = CBORServiceOptions{
	ServiceName:    "ldc:cborrpc",
	LoggingHeaders: []string{"user-agent", "x-request-id"},
}

func NewCBORService(h cborrpc.Handler, opts *CBORServiceOptions) *CBORService {
	if opts == nil {
		opts = &DefaultCBORServiceOptions
	}
	return &CBORService{h: h, name: opts.ServiceName, loggingHeaders: opts.LoggingHeaders}
}

func (s *CBORService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	xid := r.Header.Get("x-request-id")
	req := &cborrpc.Request{ID: xid}

	if log := value.CtxValue[value.Log](ctx); log.Valid() {
		start := time.Now()
		log.Set("start", value.Time(start))
		defer func() {
			log.Set("elapsed", value.Int64(int64(time.Since(start)/time.Millisecond)))
		}()

		log.Set("proto", value.String(r.Proto))
		log.Set("method", value.String(r.Method))
		log.Set("requestUri", value.String(r.RequestURI))
		log.Set("remoteAddr", value.String(r.RemoteAddr))
		log.Set("requestBytes", value.Int64(r.ContentLength))

		if len(s.loggingHeaders) > 0 {
			for _, h := range s.loggingHeaders {
				if v := r.Header.Get(h); v != "" {
					log.Set(h, value.String(v))
				}
			}
		}
	}

	if r.Method != "POST" {
		s.writeCBOR(ctx, w, http.StatusMethodNotAllowed, req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError - http.StatusMethodNotAllowed,
			Message: fmt.Sprintf("expected POST method, got %s", r.Method),
		}), "")
		return
	}

	contentType := r.Header.Get("content-type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	if contentType != cborrpc.MIMEApplicationCBOR {
		s.writeCBOR(ctx, w, http.StatusUnsupportedMediaType, req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError - http.StatusUnsupportedMediaType,
			Message: fmt.Sprintf("unsupported content-type, got %q", contentType),
		}), "")
		return
	}

	if r.ContentLength > httpcli.MaxContentLength {
		s.writeCBOR(ctx, w, http.StatusBadRequest, req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError - http.StatusBadRequest,
			Message: fmt.Sprintf("content length too large, expected <= %d", httpcli.MaxContentLength),
		}), "")
	}

	if r.ContentLength > 0 {
		req.Grow(int(r.ContentLength))
	}
	_, err := req.ReadFrom(r.Body)
	r.Body.Close()
	if req.ID == "" {
		req.ID = xid
	}

	if err != nil {
		s.writeCBOR(ctx, w, http.StatusBadRequest, req.Error(err), "")
		return
	}

	res := s.h.ServeRPC(ctx, req)
	s.writeCBOR(ctx, w, http.StatusOK, res, r.Header.Get("accept-encoding"))
}

func (s *CBORService) writeCBOR(
	ctx context.Context, w http.ResponseWriter, code int, res *cborrpc.Response, ae string) {
	w.Header().Set("content-type", cborrpc.MIMEApplicationCBOR)
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-powered-by", s.name)

	if res.ID != "" {
		w.Header().Set("x-request-id", res.ID)
	}

	exception := res.Error
	if exception == nil {
		exception = &erring.Error{}
	}

	data, err := encoding.MarshalCBOR(res)
	if exception.CatchIf(err) {
		code = 500
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError - code, Message: err.Error()},
		}
		data, err = encoding.MarshalCBOR(exception)
		if exception.CatchIf(err) {
			data, _ = encoding.MarshalCBOR(erring.RespondError{
				Err: fmt.Sprintf("impossible error: %q", err.Error()),
			})
		}
	}

	var ww io.Writer = w
	if ol := len(data); ol > compressionThreshold && ae != "" {
		switch {
		case strings.Contains(ae, "zstd"):
			zw := compress.NewZstdWriter(w)
			ww = zw
			w.Header().Add("vary", "accept-encoding")
			w.Header().Set("content-encoding", "zstd")
			w.Header().Set("x-content-length", strconv.Itoa(ol))
			defer zw.Reset()

		case strings.Contains(ae, "gzip"):
			gw := gzip.NewWriter(w)
			ww = gw
			w.Header().Add("vary", "accept-encoding")
			w.Header().Set("content-encoding", "gzip")
			w.Header().Set("x-content-length", strconv.Itoa(ol))
			defer gw.Close()
		}
	}

	w.WriteHeader(code)
	_, err = ww.Write(data)
	exception.CatchIf(err)

	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("status", value.Int(code))
		log.Set("responseBytes", value.Int(len(data)))

		if res.Error != nil {
			log.Set("responseError", value.String(res.Error.Error()))
		}

		if exception.HasErrs() {
			log.Set("error", value.String(exception.Error()))
		}
	})
}
