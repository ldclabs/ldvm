// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/httpcli"
)

const (
	gzipThreshold = 1024
)

type CBORService struct {
	h    cborrpc.Handler
	name string
}

type CBORServiceOptions struct {
	ServiceName string
}

var DefaultCBORServiceOptions = CBORServiceOptions{
	ServiceName: "ldc:cborrpc",
}

func NewCBORService(h cborrpc.Handler, opts *CBORServiceOptions) *CBORService {
	if opts == nil {
		opts = &DefaultCBORServiceOptions
	}
	return &CBORService{h: h, name: opts.ServiceName}
}

func (s *CBORService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	xid := r.Header.Get("x-request-id")
	req := &cborrpc.Request{ID: xid}

	if r.Method != "POST" {
		s.writeCBOR(ctx, w, http.StatusMethodNotAllowed, req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError - http.StatusMethodNotAllowed,
			Message: fmt.Sprintf("expected POST method, got %s", r.Method),
		}), false)
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
		}), false)
		return
	}

	if r.ContentLength > httpcli.MaxContentLength {
		s.writeCBOR(ctx, w, http.StatusBadRequest, req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError - http.StatusBadRequest,
			Message: fmt.Sprintf("content length too large, expected <= %d", httpcli.MaxContentLength),
		}), false)
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
		s.writeCBOR(ctx, w, http.StatusBadRequest, req.Error(err), false)
		return
	}

	res := s.h.ServeRPC(ctx, req)
	s.writeCBOR(ctx, w, http.StatusOK, res, r.Header.Get("accept-encoding") == "gzip")
}

func (s *CBORService) writeCBOR(
	ctx context.Context, w http.ResponseWriter, code int, res *cborrpc.Response, compress bool) {
	w.Header().Set("content-type", cborrpc.MIMEApplicationCBOR)
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-powered-by", s.name)

	if res.ID != "" {
		w.Header().Set("x-request-id", res.ID)
	}

	data, err := encoding.MarshalCBOR(res)
	if err != nil {
		code = 500
		res = &cborrpc.Response{
			ID:    res.ID,
			Error: &cborrpc.Error{Code: cborrpc.CodeServerError - code, Message: err.Error()},
		}
		data, err = encoding.MarshalCBOR(res)
	}

	if err != nil {
		res.Error.Message = fmt.Sprintf("impossible error, %v", err)
	}

	if res.Error != nil {
		s.h.OnError(ctx, res.Error)
	}

	ol := len(data)
	if compress && ol > gzipThreshold {
		b := &bytes.Buffer{}
		gw := gzip.NewWriter(b)
		gw.Write(data)
		gw.Close()

		if l := b.Len(); l > 0 && l < int(0.8*float64(ol)) {
			data = b.Bytes()
			w.Header().Add("vary", "accept-encoding")
			w.Header().Set("content-encoding", "gzip")
			// set original data length
			w.Header().Set("x-content-length", strconv.Itoa(ol))
		}
	}

	w.WriteHeader(code)
	w.Write(data)
}