// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/cborrpc"
	"go.uber.org/zap"
)

type API struct {
	bc      chain.BlockChain
	version string
}

func NewAPI(bc chain.BlockChain, version string) *API {
	return &API{bc, fmt.Sprintf("ldvm/%s", version)}
}

// RPC is the main entrypoint for the LDVM.
// https://ethereum.org/en/developers/docs/apis/json-rpc
func (api *API) RPC(req *cborrpc.Req) *cborrpc.Res {
	switch req.Method {
	case "chainID":
		chainID := api.bc.Context().ChainConfig().ChainID
		return req.Result(chainID)

	case "lastAccepted":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.ID())

	case "lastAcceptedHeight":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.Height())

	case "nextGasPrice":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(blk.NextGasPrice())

	case "issueTxs":
		return api.issueTxs(req)

	case "getTxStatus":
		return api.getTxStatus(req)

	case "getTx":
		return api.getTx(req)

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

func (api *API) issueTxs(req *cborrpc.Req) *cborrpc.Res {
	txs := ld.Txs{}

	var err error
	var tx *ld.Transaction
	if err = txs.Unmarshal(req.Params); err == nil {
		if tx, err = txs.To(); err == nil {
			err = api.bc.AddRemoteTxs(tx)
		}
	}

	if err != nil {
		return req.Error(err)
	}

	return req.Result(tx.ID)
}

func (api *API) getTxStatus(req *cborrpc.Req) *cborrpc.Res {
	var id util.Hash
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	return req.Result(api.bc.GetTxHeight(id.ToID()))
}

func (api *API) getTx(req *cborrpc.Req) *cborrpc.Res {
	var data []byte
	if err := req.DecodeParams(&data); err != nil {
		return req.Error(err)
	}
	id, err := ids.ToID(data)
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    cborrpc.CodeInvalidRequest,
			Message: err.Error()})
	}

	if height := api.bc.GetTxHeight(id); height >= 0 {
		blk, err := api.bc.GetBlockAtHeight(uint64(height))
		if err != nil {
			return req.Error(&cborrpc.Err{
				Code:    CodeServerError,
				Message: err.Error()})
		}
		if tx := blk.Tx(id); tx != nil {
			return req.Result(tx.LD())
		}
	}

	return req.Error(&cborrpc.Err{
		Code:    CodeServerError,
		Message: "tx not found"})
}

func (api *API) getBlock(req *cborrpc.Req) *cborrpc.Res {
	var id util.Hash
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("block", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getBlockAtHeight(req *cborrpc.Req) *cborrpc.Res {
	var height uint64
	if err := req.DecodeParams(&height); err != nil {
		return req.Error(err)
	}

	id, err := api.bc.GetBlockIDAtHeight(height)
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    cborrpc.CodeServerError,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("block", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getState(req *cborrpc.Req) *cborrpc.Res {
	var id util.Hash
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("state", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getAccount(req *cborrpc.Req) *cborrpc.Res {
	var id util.EthID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("account", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeAccountError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getLedger(req *cborrpc.Req) *cborrpc.Res {
	var id util.EthID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("ledger", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeAccountError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getModel(req *cborrpc.Req) *cborrpc.Res {
	var id util.ModelID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("model", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getData(req *cborrpc.Req) *cborrpc.Res {
	var id util.DataID
	if err := req.DecodeParams(&id); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("data", id[:])
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

type PrevDataParams struct {
	_       struct{} `cbor:",toarray"`
	ID      util.DataID
	Version uint64
}

func (api *API) getPrevData(req *cborrpc.Req) *cborrpc.Res {
	params := &PrevDataParams{}
	if err := req.DecodeParams(params); err != nil {
		return req.Error(err)
	}

	raw, err := api.bc.LoadRawData("prevdata", params.ID.VersionKey(params.Version))
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getNameID(req *cborrpc.Req) *cborrpc.Res {
	var name string
	if err := req.DecodeParams(name); err != nil {
		return req.Error(err)
	}

	dn, err := service.NewDN(name)
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("name", []byte(dn.ASCII()))
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) getNameData(req *cborrpc.Req) *cborrpc.Res {
	var name string
	if err := req.DecodeParams(name); err != nil {
		return req.Error(err)
	}

	dn, err := service.NewDN(name)
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error()})
	}

	raw, err := api.bc.LoadRawData("name", []byte(dn.ASCII()))
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}

	raw, err = api.bc.LoadRawData("data", raw)
	if err != nil {
		return req.Error(&cborrpc.Err{
			Code:    CodeServerError,
			Message: err.Error()})
	}
	return req.ResultRaw(raw)
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeCBORRes(w, http.StatusMethodNotAllowed, &cborrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("POST method required, got %q", r.Method),
		})
		return
	}

	contentType := r.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	if contentType != cborrpc.MIMEApplicationCBOR {
		writeCBORRes(w, http.StatusUnsupportedMediaType, &cborrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("unsupported Content-Type, got %q", contentType),
		})
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		writeCBORRes(w, http.StatusBadRequest, &cborrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("read request body error, %v", err),
		})
	}

	req, err := cborrpc.DecodeReq(buf)
	if err != nil {
		writeCBORRes(w, http.StatusBadRequest, err)
		return
	}

	logging.Debug(func() string {
		return fmt.Sprintf("Request: %s", req.String())
	})
	res := api.RPC(req)
	logging.Debug(func() string {
		return fmt.Sprintf("Response: %s", res.String())
	})
	writeCBORRes(w, http.StatusOK, res)
}

func writeCBORRes(w http.ResponseWriter, code int, val interface{}) {
	w.Header().Set("Content-Type", cborrpc.MIMEApplicationCBORCharsetUTF8)
	data, err := util.MarshalCBOR(val)
	if err != nil {
		code = 500
		val = &cborrpc.Err{Code: -32603, Message: err.Error()}
		data, _ = util.MarshalCBOR(val)
	}

	if code >= 500 {
		logging.Log.Warn("writeCBORRes error",
			zap.Int("code", code),
			zap.String("res", fmt.Sprintf("%v", val)))
	}
	w.WriteHeader(code)
	w.Write(data)
}
