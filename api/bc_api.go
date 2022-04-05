// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
)

type BlockChain interface{}

type BlockChainAPI struct{ bc BlockChain }

func NewBlockChainAPI(bc BlockChain) *BlockChainAPI {
	return &BlockChainAPI{}
}

type IssueTxArgs struct {
	Tx string `json:"tx"`
}

type IssueTxReply struct {
	TxID ids.ID `json:"txID"`
}

// IssueTx
func (api *BlockChainAPI) IssueTx(_ *http.Request, args *IssueTxArgs, reply *IssueTxReply) error {
	return nil
}

type GetTxArgs struct {
	TxID ids.ID `json:"txID"`
}

type GetTxReply struct {
	ID     ids.ID      `json:"id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// GetTx
func (api *BlockChainAPI) GetTx(_ *http.Request, args *GetTxArgs, reply *GetTxReply) error {
	return nil
}
