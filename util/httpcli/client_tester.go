//go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	l   net.Listener
	srv *http.Server
}

func NewHTTPServer(h http.Handler) *Server {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s := &Server{l: l, srv: &http.Server{
		Handler: h2c.NewHandler(h, &http2.Server{}),
	}}
	go s.srv.Serve(l)
	return s
}

func (s *Server) Close() error {
	return s.srv.Close()
}

func (s *Server) Addr() net.Addr {
	return s.l.Addr()
}
