// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"context"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/value"
)

type API struct {
	bc            chain.BlockChain
	name, version string
}

func NewAPI(bc chain.BlockChain, name, version string) *API {
	return &API{bc, name, version}
}

// ServeRPC is the main entrypoint for the LDVM.
func (api *API) ServeRPC(ctx context.Context, req *cborrpc.Request) *cborrpc.Response {
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
		return req.Result(blk.ID())

	case "lastAcceptedHeight":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.Height())

	case "nextGasPrice":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.NextGasPrice())

	case "preVerifyTxs":
		return api.preVerifyTxs(req)

	case "getGenesisTxs":
		return api.getGenesisTxs(req)

	case "getBlock":
		return api.getBlock(req)

	case "getBlockAtHeight":
		return api.getBlockAtHeight(req)

	case "getState":
		return api.getState(req)

	case "getAccount":
		return api.getAccount(req)

	case "getLedger":
		return api.getLedger(req)

	case "getModel":
		return api.getModel(req)

	case "getData":
		return api.getData(req)

	case "getPrevData":
		return api.getPrevData(req)

	case "getNameID":
		return api.getNameID(req)

	case "getNameData":
		return api.getNameData(req)

	default:
		return req.InvalidMethod()
	}
}

func (api *API) preVerifyTxs(req *cborrpc.Request) *cborrpc.Response {
	txs := ld.Txs{}

	var err error
	heigit := uint64(0)
	if err = req.DecodeParams(&txs); err == nil {
		heigit, err = api.bc.PreVerifyPdsTxs(txs...)
	}

	if err != nil {
		return req.Error(err)
	}

	return req.Result(heigit)
}

func (api *API) getGenesisTxs(req *cborrpc.Request) *cborrpc.Response {
	txs := api.bc.GetGenesisTxs()

	if len(txs) == 0 {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: "no genesis transactions, blockchain not ready"})
	}

	return req.Result(txs)
}

func (api *API) getBlock(req *cborrpc.Request) *cborrpc.Response {
	var id ids.ID32
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("block", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getBlockAtHeight(req *cborrpc.Request) *cborrpc.Response {
	var height uint64
	if err := req.DecodeParams(&height); err != nil {
		return req.Error(err)
	}

	id, err := api.bc.GetBlockIDAtHeight(height)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("block", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getState(req *cborrpc.Request) *cborrpc.Response {
	var id ids.ID32
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("state", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getAccount(req *cborrpc.Request) *cborrpc.Response {
	var id ids.Address
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("account", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getLedger(req *cborrpc.Request) *cborrpc.Response {
	var id ids.Address
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("ledger", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getModel(req *cborrpc.Request) *cborrpc.Response {
	var id ids.ModelID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("model", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getData(req *cborrpc.Request) *cborrpc.Response {
	var id ids.DataID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("data", id[:])
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

type PrevDataParams struct {
	_       struct{} `cbor:",toarray"`
	ID      ids.DataID
	Version uint64
}

func (api *API) getPrevData(req *cborrpc.Request) *cborrpc.Response {
	params := &PrevDataParams{}
	if err := req.DecodeParams(params); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("prevdata", params.ID.VersionKey(params.Version))
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getNameID(req *cborrpc.Request) *cborrpc.Response {
	var name string
	if err := req.DecodeParams(name); err != nil {
		return req.Error(err)
	}

	dn, err := service.NewDN(name)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("name", []byte(dn.ASCII()))
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getNameData(req *cborrpc.Request) *cborrpc.Response {
	var name string
	if err := req.DecodeParams(name); err != nil {
		return req.Error(err)
	}

	dn, err := service.NewDN(name)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("name", []byte(dn.ASCII()))
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}

	raw, err = api.bc.LoadRawData("data", raw)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}
