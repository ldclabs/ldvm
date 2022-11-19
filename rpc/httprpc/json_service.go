// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/klauspost/compress/gzip"

	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/compress"
	"github.com/ldclabs/ldvm/util/httpcli"
)

type JSONService struct {
	h    jsonrpc.Handler
	name string
}

type JSONServiceOptions struct {
	ServiceName string
}

var DefaultJSONServiceOptions = JSONServiceOptions{
	ServiceName: "ldc:jsonrpc",
}

func NewJSONService(h jsonrpc.Handler, opts *JSONServiceOptions) *JSONService {
	if opts == nil {
		opts = &DefaultJSONServiceOptions
	}
	return &JSONService{h: h, name: opts.ServiceName}
}

func (s *JSONService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &jsonrpc.Request{Version: "2.0", ID: r.Header.Get("x-request-id")}

	if r.Method != "POST" {
		s.writeJSON(ctx, w, http.StatusMethodNotAllowed, req.Error(&jsonrpc.Error{
			Code:    jsonrpc.CodeServerError - http.StatusMethodNotAllowed,
			Message: fmt.Sprintf("expected POST method, got %s", r.Method),
		}), "")
		return
	}

	contentType := r.Header.Get("content-type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	if contentType != jsonrpc.MIMEApplicationJSON {
		s.writeJSON(ctx, w, http.StatusUnsupportedMediaType, req.Error(&jsonrpc.Error{
			Code:    jsonrpc.CodeServerError - http.StatusUnsupportedMediaType,
			Message: fmt.Sprintf("unsupported content-type, got %q", contentType),
		}), "")
		return
	}

	if r.ContentLength > httpcli.MaxContentLength {
		s.writeJSON(ctx, w, http.StatusBadRequest, req.Error(&jsonrpc.Error{
			Code:    jsonrpc.CodeServerError - http.StatusBadRequest,
			Message: fmt.Sprintf("content length too large, expected <= %d", httpcli.MaxContentLength),
		}), "")
	}

	if r.ContentLength > 0 {
		req.Grow(int(r.ContentLength))
	}
	_, err := req.ReadFrom(r.Body)
	r.Body.Close()

	if err != nil {
		s.writeJSON(ctx, w, http.StatusBadRequest, req.Error(err), "")
		return
	}

	res := s.h.ServeRPC(ctx, req)
	s.writeJSON(ctx, w, http.StatusOK, res, r.Header.Get("accept-encoding"))
}

func (s *JSONService) writeJSON(
	ctx context.Context, w http.ResponseWriter, code int, res *jsonrpc.Response, ae string) {
	w.Header().Set("content-type", jsonrpc.MIMEApplicationJSONCharsetUTF8)
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-powered-by", s.name)

	if res.ID != "" {
		w.Header().Set("x-request-id", res.ID)
	}

	data, err := json.Marshal(res)
	if err != nil {
		code = 500
		res = &jsonrpc.Response{
			Version: "2.0",
			ID:      res.ID,
			Error:   &jsonrpc.Error{Code: jsonrpc.CodeServerError - code, Message: err.Error()},
		}
		data, err = json.Marshal(res)
	}

	if err != nil {
		res.Error.Message = fmt.Sprintf("impossible error, %v", err)
	}

	if res.Error != nil {
		s.h.OnError(ctx, res.Error)
	}

	var ww io.Writer = w
	if ol := len(data); ol > compressionThreshold && ae != "" {
		switch {
		case strings.Contains(ae, "zstd"):
			zw := &compress.ZstdWriter{W: w}
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
	ww.Write(data)
}
