// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/jsonrpc"
)

type EthAPI struct {
	bc      chain.BlockChain
	version string
}

func NewEthAPI(bc chain.BlockChain, version string) *EthAPI {
	return &EthAPI{bc, fmt.Sprintf("ldvm/%s", version)}
}

// RPC is the main entrypoint for the ETH wallet client.
// https://ethereum.org/en/developers/docs/apis/json-rpc
func (api *EthAPI) RPC(req *jsonrpc.Req) *jsonrpc.Res {
	switch req.Method {
	case "web3_clientVersion":
		return req.Result(api.version)

	case "web3_sha3":
		return api.web3Sha3(req)

	case "eth_chainId":
		chainID := api.bc.Context().ChainConfig().ChainID
		return req.Result(formatUint64(chainID))

	case "net_version":
		chainID := api.bc.Context().ChainConfig().ChainID
		return req.Result(strconv.FormatUint(chainID, 10))

	case "eth_blockNumber":
		blk := api.bc.LastAcceptedBlock()
		return req.Result(formatUint64(blk.Height()))

	case "eth_getTransactionCount":
		return api.getTransactionCount(req)

	case "eth_getBalance":
		return api.getBalance(req)

	case "eth_gasPrice":
		return api.gasPrice(req)

	case "eth_getCode":
		return req.Result("0x") // no code

	case "eth_estimateGas":
		return api.estimateGas(req)

	case "eth_getBlockByHash":
		return api.getBlockByHash(req)

	case "eth_getBlockByNumber":
		return api.getBlockByNumber(req)

	case "eth_getTransactionByHash":
		return api.getTransactionByHash(req)

	case "eth_getTransactionReceipt":
		return api.getTransactionReceipt(req)

	case "eth_call":
		return api.call(req)

	case "eth_sendRawTransaction":
		return api.sendRawTransaction(req)

	default:
		return req.InvalidMethod()
	}
}

func (api *EthAPI) web3Sha3(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
		return req.Error(err)
	}
	if len(params) != 1 {
		return req.InvalidParams("expected 1 params")
	}

	data, err := decodeBytes(params[0])
	if err != nil {
		return req.Error(err)
	}

	return req.Result(formatBytes(crypto.Keccak256(data)))
}

func (api *EthAPI) getTransactionCount(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
		return req.Error(err)
	}
	if len(params) != 2 {
		return req.InvalidParams("expected 2 params")
	}
	id, err := decodeAddress(params[0])
	if err != nil {
		return req.InvalidParams("parse address error, " + err.Error())
	}

	acc, err := api.bc.LastAcceptedBlock().State().LoadAccount(id)
	if err != nil {
		return req.Error(&jsonrpc.Err{
			Code:    CodeAccountError,
			Message: err.Error(),
		})
	}
	return req.Result(formatUint64(acc.Nonce()))
}

func (api *EthAPI) getBalance(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
		return req.Error(err)
	}
	if len(params) != 2 {
		return req.InvalidParams("expected 2 params")
	}
	id, err := decodeAddress(params[0])
	if err != nil {
		return req.InvalidParams("parse address error, " + err.Error())
	}

	acc, err := api.bc.LastAcceptedBlock().State().LoadAccount(id)
	if err != nil {
		return req.Error(&jsonrpc.Err{
			Code:    CodeAccountError,
			Message: err.Error(),
		})
	}
	return req.Result(formatEthBalance(acc.Balance()))
}

func (api *EthAPI) gasPrice(req *jsonrpc.Req) *jsonrpc.Res {
	nextGasPrice := api.bc.PreferredBlock().NextGasPrice()
	return req.Result(formatEthBalance(new(big.Int).SetUint64(nextGasPrice)))
}

func (api *EthAPI) estimateGas(req *jsonrpc.Req) *jsonrpc.Res {
	var txs []*ethTx
	if err := req.DecodeParams(&txs); err != nil {
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
	gas := tx.Gas() + 100
	// make it work with metamask
	//https://github.com/MetaMask/metamask-extension/pull/6625
	if gas < 21000 {
		gas = 21000
	}
	return req.Result(formatUint64(gas))
}

func (api *EthAPI) getBlockByHash(req *jsonrpc.Req) *jsonrpc.Res {
	var params []json.RawMessage
	if err := req.DecodeParams(&params); err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	if len(params) != 2 {
		return req.InvalidParams("parse eth_getBlockByNumber params error, invalid params")
	}

	fullTxs := string(params[1]) == "true"
	id, err := decodeHashByRaw(params[0])
	if err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}

	blk, err := api.bc.GetBlock(id)
	if err != nil {
		return req.InvalidParams("eth_getBlockByNumber error, " + err.Error())
	}

	return req.Result(toEthBlock(blk.LD(), fullTxs))
}

func (api *EthAPI) getBlockByNumber(req *jsonrpc.Req) *jsonrpc.Res {
	var params []json.RawMessage
	if err := req.DecodeParams(&params); err != nil {
		return req.InvalidParams("parse eth_getBlockByNumber params error, " + err.Error())
	}
	if len(params) != 2 {
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

func (api *EthAPI) getTransactionByHash(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
		return req.InvalidParams("parse eth_getTransactionByHash params error, " + err.Error())
	}
	if len(params) != 1 {
		return req.InvalidParams("parse eth_getTransactionByHash params error, invalid params")
	}

	id, err := decodeHash(params[0])
	if err != nil {
		return req.InvalidParams("parse eth_getTransactionByHash params error, " + err.Error())
	}
	height := api.bc.GetTxHeight(id)
	if height < 0 {
		return req.Result(map[string]interface{}{
			"blockHash":   nil,
			"blockNumber": nil,
		})
	}

	blk, err := api.bc.GetBlockAtHeight(uint64(height))
	if err != nil {
		return req.InvalidParams("eth_getTransactionByHash error, " + err.Error())
	}
	txs := toEthTxs(blk.LD(), id)
	if len(txs) == 0 {
		return req.Result(map[string]interface{}{
			"blockHash":   nil,
			"blockNumber": nil,
		})
	}
	return req.Result(txs[0])
}

func (api *EthAPI) getTransactionReceipt(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
		return req.InvalidParams("parse eth_getTransactionReceipt params error, " + err.Error())
	}
	if len(params) != 1 {
		return req.InvalidParams("parse eth_getTransactionReceipt params error, invalid params")
	}

	id, err := decodeHash(params[0])
	if err != nil {
		return req.InvalidParams("parse eth_getTransactionReceipt params error, " + err.Error())
	}
	height := api.bc.GetTxHeight(id)
	if height < 0 {
		return req.Result(map[string]interface{}{
			"blockHash":   nil,
			"blockNumber": nil,
		})
	}

	blk, err := api.bc.GetBlockAtHeight(uint64(height))
	if err != nil {
		return req.InvalidParams("eth_getTransactionByHash error, " + err.Error())
	}

	return req.Result(toEthReceipt(blk.LD(), id))
}

func (api *EthAPI) call(req *jsonrpc.Req) *jsonrpc.Res {
	var params []json.RawMessage
	if err := req.DecodeParams(&params); err != nil {
		return req.InvalidParams("parse eth_call params error, " + err.Error())
	}
	if len(params) == 0 || len(params[0]) == 0 {
		return req.InvalidParams("parse eth_call params error, invalid transaction object")
	}
	etx := &ethTx{}
	if err := json.Unmarshal(params[0], etx); err != nil {
		return req.InvalidParams("parse eth_call params error, " + err.Error())
	}

	token := util.TokenSymbol(etx.To)
	symbol := token.String()
	if symbol == "" {
		return req.Error(&jsonrpc.Err{
			Code:    CodeAccountError,
			Message: fmt.Sprintf("invalid token address %s", etx.To.String()),
		})
	}

	data := etx.GetData()
	funcSig := ""
	if len(data) >= 4 {
		funcSig = hex.EncodeToString(data[:4])
	}

	// ERC20 token
	// "name()": "06fdde03"
	// "symbol()": "95d89b41"
	// "decimals()": "313ce567"
	// "totalSupply()": "18160ddd"
	// "allowance(address,address)": "dd62ed3e"
	// "balanceOf(address)": "70a08231"
	// "supportsInterface(bytes4)": "01ffc9a7"

	// allowance(address,address)
	if funcSig == "dd62ed3e" {
		return req.Result(formatUint64(0))
	}

	switch len(data) {
	case 4:
		switch funcSig {
		case "06fdde03", "95d89b41": // name(), symbol()
			return req.Result(formatBytes([]byte(symbol)))

		case "313ce567": // decimals()
			return req.Result(formatUint64(18))

		case "18160ddd": // totalSupply()
			tokenAcc, err := api.bc.LastAcceptedBlock().State().LoadAccount(etx.To)
			if err != nil {
				return req.Error(&jsonrpc.Err{
					Code:    CodeAccountError,
					Message: err.Error(),
				})
			}

			if !tokenAcc.Valid(ld.TokenAccount) {
				return req.Error(&jsonrpc.Err{
					Code:    CodeAccountError,
					Message: fmt.Sprintf("invalid token address %s", etx.To.String()),
				})
			}
			return req.Result(formatEthBalance(tokenAcc.TotalSupply()))
		}

	case 36:
		switch funcSig {
		case "01ffc9a7": // supportsInterface(bytes4)
			return req.Result("0x")

		case "70a08231": // balanceOf(address)
			id := util.EthID{}
			copy(id[:], data[4:])
			acc, err := api.bc.LastAcceptedBlock().State().LoadAccount(id)
			if err != nil {
				return req.Error(&jsonrpc.Err{
					Code:    CodeAccountError,
					Message: err.Error(),
				})
			}
			return req.Result(formatEthBalance(acc.BalanceOf(token)))
		}
	}

	return req.InvalidParams("invalid eth_call params, only balanceOf(address) is supported")
}

func (api *EthAPI) sendRawTransaction(req *jsonrpc.Req) *jsonrpc.Res {
	var params []string
	if err := req.DecodeParams(&params); err != nil {
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

	// TODO, support ERC20 token interaface
	// "transfer(address,uint256)": "a9059cbb"

	// not support yet
	// "transferFrom(address,address,uint256)": "23b872dd"
	// "approve(address,uint256)": "095ea7b3"

	tx := etx.ToTransaction()
	if err := api.bc.SubmitTx(tx); err != nil {
		return req.InvalidParams("submit eth_sendRawTransaction error, " + err.Error())
	}

	return req.Result(formatBytes(tx.ID[:]))
}

func (api *EthAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSONRes(w, http.StatusMethodNotAllowed, &jsonrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("POST method required, got %q", r.Method),
		})
		return
	}

	contentType := r.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	if contentType != "application/json" {
		writeJSONRes(w, http.StatusUnsupportedMediaType, &jsonrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("unsupported Content-Type, got %q", contentType),
		})
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		writeJSONRes(w, http.StatusBadRequest, &jsonrpc.Err{
			Code:    -32600,
			Message: fmt.Sprintf("read request body error, %v", err),
		})
	}

	req, err := jsonrpc.DecodeReq(buf)
	if err != nil {
		writeJSONRes(w, http.StatusBadRequest, err)
		return
	}

	logging.Debug(func() string {
		return fmt.Sprintf("Request: %s", req.String())
	})
	res := api.RPC(req)
	logging.Debug(func() string {
		return fmt.Sprintf("Response: %s", res.String())
	})
	writeJSONRes(w, http.StatusOK, res)
}

func writeJSONRes(w http.ResponseWriter, code int, val interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data, err := json.Marshal(val)
	if err != nil {
		code = 500
		val = &jsonrpc.Err{Code: -32603, Message: err.Error()}
		data, _ = json.Marshal(val)
	}

	if code >= 500 {
		logging.Log.Warn("write response %d, %s", code, string(data))
	}
	w.WriteHeader(code)
	w.Write(data)
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

func (etx *ethTx) GetData() []byte {
	if len(etx.Data) < 4 || string(etx.Data[:3]) != `"0x` || etx.Data[len(etx.Data)-1] != '"' {
		return nil
	}
	data, err := hex.DecodeString(string(etx.Data[3 : len(etx.Data)-1]))
	if err != nil {
		return nil
	}
	return data
}

func toEthTxIDs(blk *ld.Block) []string {
	ids := make([]string, len(blk.Txs))
	for i, tx := range blk.Txs {
		ids[i] = formatBytes(tx.ID[:])
	}
	return ids
}

func toEthTxs(blk *ld.Block, id ids.ID) []map[string]interface{} {
	txs := make([]map[string]interface{}, 0, len(blk.Txs))
	blkID := formatBytes(blk.ID[:])
	blkNumber := formatUint64(blk.Height)
	gasPrice := formatEthBalance(new(big.Int).SetUint64(blk.GasPrice))
	for i, tx := range blk.Txs {
		if id != ids.Empty && tx.ID != id {
			continue
		}

		etx := map[string]interface{}{
			"blockHash":        blkID,
			"blockNumber":      blkNumber,
			"from":             tx.From.String(),
			"gas":              formatUint64(tx.Gas()),
			"gasPrice":         gasPrice,
			"hash":             formatBytes(tx.ID[:]),
			"input":            "0x",
			"nonce":            formatUint64(tx.Nonce),
			"to":               tx.To.String(),
			"transactionIndex": formatUint64(uint64(i)),
			"value":            formatEthBalance(tx.Amount),
			"v":                "0x0",
			"r":                "0x",
			"s":                "0x",
		}

		if eth := tx.Eth(); eth != nil {
			etx["input"] = formatBytes(eth.Data())
			v, r, s := eth.RawSignatureValues()
			etx["v"] = formatUint64(v.Uint64())

			data := make([]byte, 32)
			copy(data[:], r.Bytes())
			etx["r"] = formatBytes(data)
			copy(data[:], s.Bytes())
			etx["s"] = formatBytes(data)
		}
		txs = append(txs, etx)
	}
	return txs
}

const logsBloom = "0x00000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000080000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000"

const emptyHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

func toEthReceipt(blk *ld.Block, id ids.ID) map[string]interface{} {
	for i, tx := range blk.Txs {
		if tx.ID != id {
			continue
		}

		return map[string]interface{}{
			"transactionHash":   formatBytes(tx.ID[:]),
			"transactionIndex":  formatUint64(uint64(i)),
			"blockHash":         formatBytes(blk.ID[:]),
			"blockNumber":       formatUint64(blk.Height),
			"from":              tx.From.String(),
			"to":                tx.To.String(),
			"cumulativeGasUsed": formatUint64(blk.Gas),
			"gasUsed":           formatUint64(tx.Gas()),
			"contractAddress":   nil,
			"logs":              []string{},
			"logsBloom":         logsBloom,
			"root":              formatBytes(blk.State[:]),
			"status":            "0x1",
		}
	}

	return map[string]interface{}{
		"blockHash":   nil,
		"blockNumber": nil,
	}
}

func toEthBlock(blk *ld.Block, fullTxs bool) map[string]interface{} {
	height := formatUint64(blk.Height)
	res := map[string]interface{}{
		"number":           height,
		"hash":             formatBytes(blk.ID[:]),
		"parentHash":       formatBytes(blk.Parent[:]),
		"nonce":            height,
		"sha3Uncles":       emptyHash,
		"logsBloom":        logsBloom,
		"transactionsRoot": emptyHash,
		"stateRoot":        formatBytes(blk.State[:]),
		"receiptsRoot":     emptyHash,
		"miner":            util.EthID(blk.Miner).String(),
		"difficulty":       "0x1",
		"totalDifficulty":  height,
		"extraData":        "0x",
		"size":             formatUint64(uint64(len(blk.Bytes()))),
		"gasLimit":         formatUint64(blk.Gas * 10),
		"gasUsed":          formatUint64(blk.Gas),
		"timestamp":        formatUint64(blk.Timestamp),
		"uncles":           []string{},
		"gasPrice":         formatEthBalance(new(big.Int).SetUint64(blk.GasPrice)),
	}

	if fullTxs {
		res["transactions"] = toEthTxs(blk, ids.Empty)
	} else {
		res["transactions"] = toEthTxIDs(blk)
	}
	return res
}
