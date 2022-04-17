// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/app"
)

type BlockChainAPI struct {
	state chain.StateDB
}

func NewBlockChainAPI(state chain.StateDB) *BlockChainAPI {
	return &BlockChainAPI{state}
}

type GetReply struct {
	ID     string      `json:"id"`
	Status string      `json:"status,omitempty"`
	Data   interface{} `json:"data"`
}

// GetChainConfig
func (api *BlockChainAPI) GetChainConfig(_ *http.Request, _ *NoArgs, reply *genesis.ChainConfig) error {
	*reply = *api.state.Context().Chain()
	return nil
}

// GetTotalSupply
func (api *BlockChainAPI) GetTotalSupply(_ *http.Request, _ *NoArgs, reply *GetBalanceReply) error {
	reply.Balance = api.state.TotalSupply()
	return nil
}

// GetRecentEvents
func (api *BlockChainAPI) GetRecentEvents(_ *http.Request, _ *NoArgs, reply *[]*chain.Event) error {
	*reply = api.state.QueryEvents()
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
	acc, err := api.state.LoadAccount(id)
	if err != nil {
		return err
	}
	reply.Balance = acc.Balance
	return nil
}

// GetAccount
func (api *BlockChainAPI) GetAccount(_ *http.Request, args *GetAccountArgs, reply *ld.Account) error {
	id := ids.ShortID(args.ID)
	if id == constants.BlackholeAddr {
		return fmt.Errorf("invalid address: %v", args.ID)
	}
	acc, err := api.state.LoadAccount(id)
	if err != nil {
		return err
	}
	*reply = *acc
	return nil
}

type GetBlockArgs struct {
	ID     ids.ID  `json:"id"`
	Height *uint64 `json:"height"`
}

// GetBlock
func (api *BlockChainAPI) GetBlock(_ *http.Request, args *GetBlockArgs, reply *GetReply) error {
	if args.Height != nil {
		id, err := api.state.GetBlockIDAtHeight(*args.Height)
		if err != nil {
			return fmt.Errorf("invalid block height: %v", *args.Height)
		}
		args.ID = id
	}

	if args.ID == ids.Empty {
		return fmt.Errorf("invalid block id: %v", args.ID)
	}
	blk, err := api.state.GetBlock(args.ID)
	if err != nil {
		return err
	}
	reply.ID = blk.ID().String()
	reply.Data = blk
	return nil
}

// GetTx
func (api *BlockChainAPI) GetTx(_ *http.Request, args *GetBlockArgs, reply *GetReply) error {
	if args.ID == ids.Empty {
		return fmt.Errorf("invalid transaction id: %v", args.ID)
	}

	tx := api.state.GetTx(args.ID)
	if tx == nil {
		return fmt.Errorf("transaction %v not found in cache", args.ID)
	}

	reply.ID = tx.ID().String()
	reply.Status = tx.Status()
	reply.Data = tx
	return nil
}

type GetModelArgs struct {
	ID ld.ModelID `json:"id"`
}

// GetModel
func (api *BlockChainAPI) GetModel(_ *http.Request, args *GetModelArgs, reply *GetReply) error {
	id := ids.ShortID(args.ID)
	if id == ids.ShortEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.state.LoadModel(id)
	if err != nil {
		return err
	}
	reply.ID = args.ID.String()
	reply.Data = data
	return nil
}

type GetDataArgs struct {
	ID      ld.DataID `json:"id"`
	Version uint64    `json:"version"`
}

// GetData
func (api *BlockChainAPI) GetData(_ *http.Request, args *GetDataArgs, reply *GetReply) error {
	id := ids.ShortID(args.ID)
	if id == ids.ShortEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.state.LoadData(id)
	if err != nil {
		return err
	}
	reply.ID = args.ID.String()
	reply.Data = data
	return nil
}

// GetPrevDatas
func (api *BlockChainAPI) GetPrevDatas(_ *http.Request, args *GetDataArgs, reply *GetReply) error {
	id := ids.ShortID(args.ID)
	if id == ids.ShortEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}

	num := 10
	rt := make([]*ld.DataMeta, 0, num)
	ver := args.Version
	for ver > 0 && len(rt) < num {
		data, err := api.state.LoadPrevData(id, ver)
		if err != nil {
			return err
		}
		rt = append(rt, data)
		ver--
	}

	reply.ID = args.ID.String()
	reply.Data = rt
	return nil
}

type ResolveArgs struct {
	Name string `json:"name"`
}

// GetData
func (api *BlockChainAPI) Resolve(_ *http.Request, args *ResolveArgs, reply *GetReply) error {
	if !app.ValidDomainName(args.Name) {
		return fmt.Errorf("invalid name %s to resolve", strconv.Quote(args.Name))
	}
	data, err := api.state.ResolveName(args.Name)
	if err != nil {
		return err
	}
	reply.ID = ld.DataID(data.ID).String()
	reply.Data = data
	return nil
}
