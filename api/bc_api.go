// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
)

type BlockChainAPI struct {
	state chain.StateDB
}

func NewBlockChainAPI(state chain.StateDB) *BlockChainAPI {
	return &BlockChainAPI{state}
}

type NoArgs struct{}

// GetChainConfig
func (api *BlockChainAPI) GetChainConfig(_ *http.Request, _ *NoArgs, reply *genesis.ChainConfig) error {
	*reply = *api.state.ChainConfig()
	return nil
}

// GetTotalSupply
func (api *BlockChainAPI) GetTotalSupply(_ *http.Request, _ *NoArgs, reply *GetBalanceReply) error {
	reply.Balance = api.state.PreferredBlock().State().TotalSupply()
	return nil
}

// GetRecentEvents
func (api *BlockChainAPI) GetRecentEvents(_ *http.Request, _ *NoArgs, reply *[]*chain.Event) error {
	*reply = api.state.RecentEvents()
	return nil
}

type GetAccountArgs struct {
	ID ld.EthID `json:"address"`
}

type GetBalanceReply struct {
	Balance *big.Int `json:"balance"`
}

// GetBalance
func (api *BlockChainAPI) GetBalance(_ *http.Request, args *GetAccountArgs, reply *GetBalanceReply) error {
	id := ids.ShortID(args.ID)
	if id == constants.BlackholeAddr {
		return fmt.Errorf("invalid address: %v", args.ID)
	}
	acc, err := api.state.PreferredBlock().State().LoadAccount(id)
	if err != nil {
		return err
	}
	reply.Balance = acc.Balance()
	return nil
}

// GetAccount
func (api *BlockChainAPI) GetAccount(_ *http.Request, args *GetAccountArgs, reply *ld.Account) error {
	id := ids.ShortID(args.ID)
	if id == constants.BlackholeAddr {
		return fmt.Errorf("invalid address: %v", args.ID)
	}
	acc, err := api.state.PreferredBlock().State().LoadAccount(id)
	if err != nil {
		return err
	}
	*reply = *acc.Account().Copy()
	return nil
}

type GetBlockArgs struct {
	ID ids.ID `json:"id"`
}

// GetBlock
func (api *BlockChainAPI) GetBlock(_ *http.Request, args *GetBlockArgs, reply *chain.Block) error {
	if args.ID == ids.Empty {
		return fmt.Errorf("invalid block id: %v", args.ID)
	}
	blk, err := api.state.GetBlock(args.ID)
	if err != nil {
		return err
	}
	*reply = *blk
	return nil
}

type GetModelArgs struct {
	ID ld.ModelID `json:"id"`
}

// GetModel
func (api *BlockChainAPI) GetModel(_ *http.Request, args *GetModelArgs, reply *ld.ModelMeta) error {
	id := ids.ShortID(args.ID)
	if id == ids.ShortEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.state.PreferredBlock().State().LoadModel(id)
	if err != nil {
		return err
	}
	*reply = *data
	return nil
}

type GetDataArgs struct {
	ID ld.DataID `json:"id"`
}

// GetData
func (api *BlockChainAPI) GetData(_ *http.Request, args *GetDataArgs, reply *ld.DataMeta) error {
	id := ids.ShortID(args.ID)
	if id == ids.ShortEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.state.PreferredBlock().State().LoadData(id)
	if err != nil {
		return err
	}
	*reply = *data
	return nil
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
