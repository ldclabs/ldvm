// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txpool

import (
	"context"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/rpc/httprpc"
)

// Chain is Chain's tx interface.
type Chain interface {
	GetGenesisTxs(context.Context) (ld.Txs, error)
	GetAccount(context.Context, ids.Address) (*ld.Account, error)
	PreVerifyTxs(context.Context, ld.Txs) error
}

type chainAPI struct {
	cli *httprpc.CBORClient
}

func NewChainAPI(endpoint string, opts *httprpc.CBORClientOptions) Chain {
	return &chainAPI{cli: httprpc.NewCBORClient(endpoint, opts)}
}

func (api *chainAPI) GetGenesisTxs(ctx context.Context) (ld.Txs, error) {
	var txs ld.Txs
	res := api.cli.Request(ctx, "getGenesisTxs", nil, &txs)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, tx := range txs {
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
	}
	return txs, nil
}

func (api *chainAPI) GetAccount(ctx context.Context, id ids.Address) (*ld.Account, error) {
	acc := &ld.Account{}
	res := api.cli.Request(ctx, "getAccount", id, acc)
	if res.Error != nil {
		return nil, res.Error
	}
	if err := acc.SyntacticVerify(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (api *chainAPI) PreVerifyTxs(ctx context.Context, txs ld.Txs) error {
	res := api.cli.Request(ctx, "preVerifyTxs", txs, nil)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
