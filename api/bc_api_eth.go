// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

type EthAPI struct {
	bc chain.BlockChain
}

func NewEthAPI(bc chain.BlockChain) *EthAPI {
	return &EthAPI{bc}
}

// RPC is the main entrypoint for the ETH wallet client.
// https://ethereum.org/en/developers/docs/apis/json-rpc
func (api *EthAPI) RPC(req *JsonrpcReq) *JsonrpcRes {
	switch req.Method {
	case "eth_chainId":
		chainId := api.bc.Context().ChainConfig().ChainID
		return req.Result("0x" + strconv.FormatUint(chainId, 16))

	case "net_version":
		chainId := api.bc.Context().ChainConfig().ChainID
		return req.Result(strconv.FormatUint(chainId, 10))

	case "eth_blockNumber":
		blk := api.bc.LastAcceptedBlock()
		return req.Result("0x" + strconv.FormatUint(blk.Height(), 16))

	case "eth_getBalance":
		return api.getBalance(req)

	case "eth_gasPrice":
		return api.gasPrice(req)

	case "eth_getCode":
		return req.Result("0x") // no code

	case "eth_estimateGas":
		return api.estimateGas(req)

	case "eth_call":
		return api.call(req)

	case "eth_sendRawTransaction":
		return api.sendRawTransaction(req)

	case "eth_getBlockByHash":
		return api.getBlockByHash(req)

	case "eth_getBlockByNumber":
		return api.getBlockByNumber(req)

	default:
		return req.InvalidMethod()
	}
}

func (api *EthAPI) getBalance(req *JsonrpcReq) *JsonrpcRes {
	var params []string
	if err := req.ParseParams(&params); err != nil {
		return req.Error(err)
	}
	if len(params) != 2 {
		return req.InvalidParams("expected 2 params")
	}
	id, err := util.EthIDFromString(params[0])
	if err != nil {
		return req.InvalidParams("parse address error, " + err.Error())
	}

	acc, err := api.bc.LastAcceptedBlock().State().LoadAccount(id)
	if err != nil {
		return req.Error(&JsonrpcErr{
			Code:    AccountErrorCode,
			Message: err.Error(),
		})
	}
	return req.Result(toEthBalance(acc.Balance()))
}

func (api *EthAPI) gasPrice(req *JsonrpcReq) *JsonrpcRes {
	nextGasPrice := api.bc.PreferredBlock().NextGasPrice()
	return req.Result(toEthBalance(new(big.Int).SetUint64(nextGasPrice)))
}

func (api *EthAPI) estimateGas(req *JsonrpcReq) *JsonrpcRes {
	var txs []*ethTx
	if err := req.ParseParams(&txs); err != nil {
		return req.InvalidParams("parse eth_estimateGas params error, " + err.Error())
	}
	if len(txs) == 0 || txs[0] == nil {
		return req.InvalidParams("parse eth_estimateGas params error, invalid transaction object")
	}

	// mock tx to estimate gas cost
	tx := &ld.TxData{
		Type:       ld.TypeTransfer,
		ChainID:    api.bc.Context().ChainConfig().ChainID,
		Nonce:      10000,
		GasTip:     1000,
		GasFeeCap:  1000000,
		From:       txs[0].From,
		To:         &txs[0].To,
		Amount:     ld.FromEthBalance(txs[0].GetValue()),
		Data:       util.RawData(txs[0].Data),
		Signatures: []util.Signature{util.SignatureEmpty},
	}

	if err := tx.SyntacticVerify(); err != nil {
		return req.InvalidParams("parse eth_estimateGas params error, " + err.Error())
	}
	return req.Result("0x" + strconv.FormatUint(tx.Gas()+100, 16))
}

func (api *EthAPI) getBlockByHash(req *JsonrpcReq) *JsonrpcRes {
	var params []json.RawMessage
	if err := req.ParseParams(&params); err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	if len(params) != 2 || len(params[0]) < 4 {
		return req.InvalidParams("parse eth_getBlockByNumber params error, invalid params")
	}

	fullTxs := string(params[1]) == "true"
	data, err := hex.DecodeString(string(params[0][1 : len(params[0])-1]))
	if err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	id, err := ids.ToID(data)
	if err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}

	blk, err := api.bc.GetBlock(id)
	if err != nil {
		return req.InvalidParams("eth_getBlockByNumber error, " + err.Error())
	}

	return req.Result(toEthBlock(blk.LD(), fullTxs))
}

func (api *EthAPI) getBlockByNumber(req *JsonrpcReq) *JsonrpcRes {
	var params []json.RawMessage
	if err := req.ParseParams(&params); err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	if len(params) != 2 || len(params[0]) < 4 {
		return req.InvalidParams("parse eth_getBlockByNumber params error, invalid params")
	}
	fullTxs := string(params[1]) == "true"
	number, err := strconv.ParseUint(string(params[0][1:len(params[0])-1]), 0, 64)
	if err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	id, err := api.bc.GetBlockIDAtHeight(number)
	if err != nil {
		return req.InvalidParams("eth_getBlockByNumber error, " + err.Error())
	}
	blk, err := api.bc.GetBlock(id)
	if err != nil {
		return req.InvalidParams("eth_getBlockByNumber error, " + err.Error())
	}

	return req.Result(toEthBlock(blk.LD(), fullTxs))
}

func (api *EthAPI) call(req *JsonrpcReq) *JsonrpcRes {
	var params []json.RawMessage
	if err := req.ParseParams(&params); err != nil {
		return req.InvalidParams("parse eth_call params error, " + err.Error())
	}
	if len(params) == 0 || len(params[0]) == 0 {
		return req.InvalidParams("parse eth_call params error, invalid transaction object")
	}
	etx := &ethTx{}
	if err := json.Unmarshal(params[0], etx); err != nil {
		return req.InvalidParams("parse eth_call params error, " + err.Error())
	}
	if len(etx.Data) != 76 || string(etx.Data[1:11]) != "0x70a08231" {
		return req.InvalidParams("invalid eth_call params, only balanceOf(address) is supported")
	}

	id, err := util.EthIDFromString(string(etx.Data[35:75]))
	if err != nil {
		return req.InvalidParams("parse address error, " + err.Error())
	}

	acc, err := api.bc.LastAcceptedBlock().State().LoadAccount(id)
	if err != nil {
		return req.Error(&JsonrpcErr{
			Code:    AccountErrorCode,
			Message: err.Error(),
		})
	}
	return req.Result(toEthBalance(acc.BalanceOf(util.TokenSymbol(etx.To))))
}

func (api *EthAPI) sendRawTransaction(req *JsonrpcReq) *JsonrpcRes {
	var params []string
	if err := req.ParseParams(&params); err != nil {
		return req.InvalidParams("parse eth_call params error, " + err.Error())
	}
	if len(params) != 1 || len(params[0]) < 100 {
		return req.InvalidParams("parse eth_sendRawTransaction params error, invalid raw transaction")
	}

	data, err := hex.DecodeString(params[0][2:])
	if err != nil {
		return req.InvalidParams("parse eth_sendRawTransaction params error, invalid raw transaction, " + err.Error())
	}

	etx := &ld.TxEth{}
	if err := etx.Unmarshal(data); err != nil {
		return req.InvalidParams("parse eth_sendRawTransaction params error, " + err.Error())
	}
	if err := etx.SyntacticVerify(); err != nil {
		return req.InvalidParams("submit eth_sendRawTransaction error, " + err.Error())
	}

	tx := etx.ToTransaction()
	if err := api.bc.SubmitTx(tx); err != nil {
		return req.InvalidParams("submit eth_sendRawTransaction error, " + err.Error())
	}

	return req.Result("0x" + hex.EncodeToString(tx.ID[:]))
}

func (api *EthAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeRes(w, http.StatusMethodNotAllowed, &JsonrpcErr{
			Code:    -32600,
			Message: "EthAPI: POST method required, received " + r.Method,
		})
		return
	}

	contentType := r.Header.Get("Content-Type")
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}
	if contentType != "application/json" {
		writeRes(w, http.StatusUnsupportedMediaType, &JsonrpcErr{
			Code:    -32600,
			Message: "EthAPI: unsupported Content-Type, " + contentType,
		})
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		writeRes(w, http.StatusBadRequest, &JsonrpcErr{
			Code:    -32600,
			Message: "EthAPI: invalid request, " + err.Error(),
		})
	}

	req, err := NewJsonrpcRequest(buf)
	if err != nil {
		writeRes(w, http.StatusBadRequest, err)
		return
	}

	logging.Log.Info("ETH Req: %s", req.String())
	writeRes(w, http.StatusOK, api.RPC(req))
}

func writeRes(w http.ResponseWriter, code int, val interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	data, err := json.Marshal(val)
	if err != nil {
		val = &JsonrpcErr{Code: -32603, Message: err.Error()}
		data, _ = json.Marshal(val)
	}

	logging.Log.Info("ETH Res: %d, %s", code, string(data))
	w.Write(data)
}

func toEthBalance(amount *big.Int) string {
	return "0x" + ld.ToEthBalance(amount).Text(16)
}

type ethTx struct {
	From     util.EthID      `json:"from,omitempty"`
	To       util.EthID      `json:"to,omitempty"`
	Nonce    string          `json:"nonce,omitempty"`
	Gas      string          `json:"gas,omitempty"`
	GasPrice string          `json:"gasPrice,omitempty"`
	Value    string          `json:"value,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"` // null or hex string
}

func (etx *ethTx) GetValue() *big.Int {
	v := new(big.Int)
	v.SetString(etx.Value, 0)
	return v
}

func toEthTxIDs(txs ld.Txs) []string {
	ids := make([]string, len(txs))
	for i, tx := range txs {
		ids[i] = "0x" + hex.EncodeToString(tx.ID[:])
	}
	return ids
}

func toEthTxs(txs ld.Txs, blk *ld.Block) []map[string]interface{} {
	etxs := make([]map[string]interface{}, len(txs))
	blkID := "0x" + hex.EncodeToString(blk.ID[:])
	blkNumber := "0x" + strconv.FormatUint(blk.Height, 16)
	gasPrice := new(big.Int).SetUint64(blk.GasPrice)
	for i, tx := range txs {
		etxs[i] = map[string]interface{}{
			"blockHash":            blkID,
			"blockNumber":          blkNumber,
			"from":                 tx.From.String(),
			"gas":                  "0x" + strconv.FormatUint(tx.Gas(), 16),
			"gasPrice":             "0x" + ld.ToEthBalance(gasPrice).Text(16),
			"maxFeePerGas":         "0x" + strconv.FormatUint(tx.GasFeeCap, 16),
			"maxPriorityFeePerGas": "0x" + strconv.FormatUint(tx.GasTip, 16),
			"hash":                 "0x" + hex.EncodeToString(tx.ID[:]),
			"input":                "0x",
			"nonce":                "0x" + strconv.FormatUint(tx.Nonce, 16),
			"to":                   tx.From.String(),
			"value":                "0x" + ld.ToEthBalance(tx.Amount).Text(16),
			"type":                 "0x2",
			"accessList":           []string{},
			"chainId":              "0x" + strconv.FormatUint(tx.ChainID, 16),
		}
	}
	return etxs
}

func toEthBlock(blk *ld.Block, fullTxs bool) map[string]interface{} {
	cost := new(big.Int).SetUint64(blk.GasPrice)
	cost = cost.Mul(cost, new(big.Int).SetUint64(blk.Gas))
	res := map[string]interface{}{
		"baseFeePerGas": "0x" + strconv.FormatUint(blk.GasPrice, 16),
		"blockGasCost":  "0x" + ld.ToEthBalance(cost).Text(16),
		"difficulty":    "0x1",
		"gasUsed":       "0x" + strconv.FormatUint(blk.Gas, 16),
		"hash":          "0x" + hex.EncodeToString(blk.ID[:]),
		"miner":         util.EthID(blk.Miner).String(),
		"number":        "0x" + strconv.FormatUint(blk.Height, 16),
		"parentHash":    "0x" + hex.EncodeToString(blk.Parent[:]),
		"size":          "0x" + strconv.FormatUint(uint64(len(blk.Bytes())), 16),
		"timestamp":     "0x" + strconv.FormatUint(blk.Timestamp, 16),
		"uncles":        []string{},
	}

	if fullTxs {
		res["transactions"] = toEthTxs(blk.Txs, blk)
	} else {
		res["transactions"] = toEthTxIDs(blk.Txs)
	}
	return res
}
