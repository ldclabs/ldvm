// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"

	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type BlockChainAPI struct {
	bc chain.BlockChain
}

func NewBlockChainAPI(bc chain.BlockChain) *BlockChainAPI {
	return &BlockChainAPI{bc}
}

type GetReply struct {
	ID     string      `json:"id"`
	Status string      `json:"status,omitempty"`
	Data   interface{} `json:"data"`
}

// GetChainConfig
func (api *BlockChainAPI) GetChainConfig(_ *http.Request, _ *NoArgs, reply *genesis.ChainConfig) error {
	*reply = *api.bc.Context().ChainConfig()
	return nil
}

// GetTotalSupply
func (api *BlockChainAPI) GetTotalSupply(_ *http.Request, _ *NoArgs, reply *GetBalanceReply) error {
	reply.Balance = api.bc.TotalSupply()
	return nil
}

type GetAccountArgs struct {
	ID util.EthID `json:"address"`
}

type GetBalanceReply struct {
	Balance *big.Int `json:"balance"`
}

// GetBalance
func (api *BlockChainAPI) GetBalance(_ *http.Request, args *GetAccountArgs, reply *GetBalanceReply) error {
	if args.ID == constants.LDCAccount {
		return fmt.Errorf("invalid address: %v", args.ID)
	}
	acc, err := api.bc.LoadAccount(args.ID)
	if err != nil {
		return err
	}
	reply.Balance = acc.Balance
	return nil
}

// GetAccount
func (api *BlockChainAPI) GetAccount(_ *http.Request, args *GetAccountArgs, reply *ld.Account) error {
	if args.ID == constants.LDCAccount {
		return fmt.Errorf("invalid address: %v", args.ID)
	}
	acc, err := api.bc.LoadAccount(args.ID)
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
		id, err := api.bc.GetBlockIDAtHeight(*args.Height)
		if err != nil {
			return fmt.Errorf("invalid block height: %v", *args.Height)
		}
		args.ID = id
	}

	if args.ID == ids.Empty {
		return fmt.Errorf("invalid block id: %v", args.ID)
	}
	blk, err := api.bc.GetBlock(args.ID)
	if err != nil {
		return err
	}
	reply.ID = blk.ID().String()
	reply.Data = blk
	return nil
}

// GetTxStatus
func (api *BlockChainAPI) GetTxStatus(_ *http.Request, args *GetBlockArgs, reply *GetReply) error {
	if args.ID == ids.Empty {
		return fmt.Errorf("invalid transaction id: %v", args.ID)
	}

	switch api.bc.GetTxHeight(args.ID) {
	case -3:
		reply.Status = choices.Unknown.String()
	case -2:
		reply.Status = choices.Rejected.String()
	case -1:
		reply.Status = choices.Processing.String()
	default:
		reply.Status = choices.Accepted.String()
	}

	reply.ID = args.ID.String()
	return nil
}

// GetTx
func (api *BlockChainAPI) GetTx(_ *http.Request, args *GetBlockArgs, reply *GetReply) error {
	if args.ID == ids.Empty {
		return fmt.Errorf("invalid transaction id: %v", args.ID)
	}

	height := api.bc.GetTxHeight(args.ID)
	switch height {
	case -3:
		reply.Status = choices.Unknown.String()
	case -2:
		reply.Status = choices.Rejected.String()
	case -1:
		reply.Status = choices.Processing.String()
	default:
		reply.Status = choices.Accepted.String()
	}

	if height >= 0 {
		blk, err := api.bc.GetBlockAtHeight(uint64(height))
		if err != nil {
			return err
		}
		reply.Data = blk.Tx(args.ID)
	}

	reply.ID = args.ID.String()
	return nil
}

type GetModelArgs struct {
	ID util.ModelID `json:"id"`
}

// GetModel
func (api *BlockChainAPI) GetModel(_ *http.Request, args *GetModelArgs, reply *GetReply) error {
	if args.ID == util.ModelIDEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.bc.LoadModel(args.ID)
	if err != nil {
		return err
	}
	reply.ID = args.ID.String()
	reply.Data = data
	return nil
}

type GetDataArgs struct {
	ID      util.DataID `json:"id"`
	Version uint64      `json:"version"`
}

// GetData
func (api *BlockChainAPI) GetData(_ *http.Request, args *GetDataArgs, reply *GetReply) error {
	if args.ID == util.DataIDEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}
	data, err := api.bc.LoadData(args.ID)
	if err != nil {
		return err
	}
	reply.ID = args.ID.String()
	reply.Data = data
	return nil
}

// GetPrevDatas
func (api *BlockChainAPI) GetPrevDatas(_ *http.Request, args *GetDataArgs, reply *GetReply) error {
	if args.ID == util.DataIDEmpty {
		return fmt.Errorf("invalid data id: %s", args.ID)
	}

	num := 10
	rt := make([]*ld.DataInfo, 0, num)
	ver := args.Version
	for ver > 0 && len(rt) < num {
		data, err := api.bc.LoadPrevData(args.ID, ver)
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
	data, err := api.bc.ResolveName(args.Name)
	if err != nil {
		return err
	}
	reply.ID = data.DataID.String()
	reply.Data = data
	return nil
}
