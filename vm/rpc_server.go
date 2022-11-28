// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	avalogging "github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ldclabs/ldvm/api"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/util/value"
)

type RPCServer interface {
	Start(h http.Handler, addr string)
	Shutdown(ctx context.Context) error
	Done() <-chan struct{}
	Err() error
}

type mux struct {
	log     avalogging.Logger
	cborrpc http.Handler
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := value.Log{Value: value.NewMap(16)}
	ctx := value.CtxWith(r.Context(), &log)
	r = r.WithContext(ctx)

	switch {
	case r.Method == "POST" && r.URL.Path == "/cborrpc/v1":
		m.cborrpc.ServeHTTP(w, r)

	default:
		w.WriteHeader(http.StatusMisdirectedRequest)
		w.Write([]byte(fmt.Sprintf(`misdirected request %q`,
			r.Method+" "+r.URL.Path)))
	}

	m.log.Info("ServeHTTP", logging.MapToFields(log.ToMap())...)
}

func (v *VM) WithRPCServer(s RPCServer) {
	v.rpc = s
}

func (v *VM) startRPCServer(addr string) error {
	cborAPI := httprpc.NewCBORService(api.NewAPI(v.bc, Name, Version.String()), nil)
	v.rpc.Start(&mux{log: v.Log, cborrpc: cborAPI}, addr)

	time.Sleep(time.Second)
	select {
	case <-v.rpc.Done():
		return v.rpc.Err()
	default:
		return nil
	}
}
