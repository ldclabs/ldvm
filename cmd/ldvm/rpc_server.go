// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"net/http"

	"github.com/ldclabs/ldvm/util/sync"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type rpcServer struct {
	done chan struct{}
	srv  http.Server
	err  sync.Value[error]
}

func (s *rpcServer) Start(h http.Handler, addr string) {
	s.srv.Handler = h2c.NewHandler(h, &http2.Server{})
	s.srv.Addr = addr
	s.done = make(chan struct{})
	go func() {
		s.err.Store(s.srv.ListenAndServe())
		s.done <- struct{}{}
	}()
}

func (s *rpcServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *rpcServer) Done() <-chan struct{} {
	return s.done
}

func (s *rpcServer) Err() error {
	err, _ := s.err.Load()
	return err
}
