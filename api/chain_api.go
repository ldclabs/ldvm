// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/value"
)

type ChainAPI struct {
	name, version string
	bc            chain.BlockChain
	rpc           http.Handler
}

func NewChainAPI(bc chain.BlockChain, name, version string) *ChainAPI {
	api := &ChainAPI{bc: bc, name: name, version: version}
	api.rpc = httprpc.NewJSONService(api, nil)
	return api
}

func (api *ChainAPI) ServeRPC(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("rpcId", value.String(req.ID))
		log.Set("rpcMethod", value.String(req.Method))
	})

	switch req.Method {
	case "version":
		return req.Result(map[string]string{
			"name":    api.name,
			"version": api.version,
		})

	case "info":
		return req.Result(api.bc.Info())

	case "chainID":
		chainID := api.bc.Context().ChainConfig().ChainID
		return req.Result(chainID)

	case "chainState":
		return req.Result(api.bc.State())

	case "lastAccepted":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.Hash())

	case "lastAcceptedHeight":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.Height())

	default:
		return req.InvalidMethod()
	}
}

func (api *ChainAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := value.Log{Value: value.NewMap(16)}
	ctx := value.CtxWith(r.Context(), &log)
	r = r.WithContext(ctx)

	switch {
	case r.Method == "POST":
		api.rpc.ServeHTTP(w, r)

	default:
		w.WriteHeader(http.StatusMisdirectedRequest)
		w.Write([]byte(fmt.Sprintf(`misdirected request %q`,
			r.Method+" "+r.URL.Path)))
	}

	// api.Log.Info("ServeHTTP", logging.MapToFields(log.ToMap())...)
}
