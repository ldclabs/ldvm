// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"
	"encoding/binary"
	"net/http"
	"strings"
	"time"

	"github.com/ldclabs/cose/cose"
	"github.com/ldclabs/cose/cwt"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/txpool"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/sync"

	"go.uber.org/zap"
)

type TxPool struct {
	mu        sync.Mutex
	ctx       *Context
	cli       *httprpc.CBORClient
	requester string
	cwtExData []byte
	cache     map[uint64]ld.Txs // cached processing Txs
}

func NewTxPool(ctx *Context, posEndpoint string, rt http.RoundTripper) *TxPool {
	header := http.Header{}
	header.Set("x-node-id", ctx.NodeID.String())
	header.Set("user-agent",
		strings.Replace(ctx.Name(), "@v", "/", 1)+" (TxPool Client, CBOR-RPC)")

	opts := &httprpc.CBORClientOptions{RoundTripper: rt, Header: header}
	exData := make([]byte, 8)
	binary.BigEndian.PutUint64(exData, ctx.genesis.Chain.ChainID)
	return &TxPool{
		ctx:       ctx,
		cli:       httprpc.NewCBORClient(posEndpoint+"/txs", opts),
		requester: ctx.Builder().String(),
		cwtExData: exData,
		cache:     make(map[uint64]ld.Txs, 10),
	}
}

func (p *TxPool) genParams(params any, withSigs bool) (*txpool.RequestParams, error) {
	rp := &txpool.RequestParams{}

	var err error
	if params != nil {
		if rp.Payload, err = encoding.MarshalCBOR(params); err != nil {
			return nil, err
		}
	}

	if withSigs {
		now := uint64(time.Now().Unix())
		rp.CWT = &cose.Sign1Message[cwt.Claims]{
			Payload: cwt.Claims{
				Subject:    p.requester,
				Audience:   "ldc:txpool",
				Expiration: now + 10,
				IssuedAt:   now,
				CWTID:      ids.ID32FromData(rp.Payload).Bytes(),
			},
		}
		if err := rp.CWT.WithSign(p.ctx.BuilderSigner(), p.cwtExData); err != nil {
			return nil, err
		}
	}

	return rp, nil
}

func (p *TxPool) LoadByIDs(ctx context.Context, height uint64, txIDs ids.IDList[ids.ID32]) (ld.Txs, error) {
	errp := erring.ErrPrefix("chain.TxPool.LoadByIDs: ")

	txs, ok := p.loadByIDsFromCache(height, txIDs)
	if !ok {
		txs = make(ld.Txs, 0, len(txIDs))
		params, err := p.genParams(txIDs, false)
		if err != nil {
			return nil, errp.ErrorIf(err)
		}

		res := p.cli.Request(ctx, "loadByIDs", params, &txs)
		if res.Error != nil {
			return nil, errp.ErrorIf(res.Error)
		}
	}

	if len(txs) != len(txIDs) {
		return nil, errp.Errorf("invalid txs length, expected %d, got %d", len(txIDs), len(txs))
	}

	for i := range txIDs {
		tx := txs[i]
		if err := tx.SyntacticVerify(); err != nil {
			return nil, errp.ErrorIf(err)
		}

		if tx.ID != txIDs[i] {
			return nil, errp.Errorf("invalid tx id, expected %s, got %s", txIDs[i], tx.ID)
		}
	}
	return txs, nil
}

func (p *TxPool) SizeToBuild(ctx context.Context, height uint64) int {
	size := 0
	params, err := p.genParams(nil, false)
	if err == nil {
		res := p.cli.Request(ctx, "sizeToBuild", params, &size)
		err = res.Error
	}

	if err != nil {
		logging.Log.Warn("TxPool.SizeToBuild",
			zap.Uint64("height", height),
			zap.Error(err))
	}
	return size
}

func (p *TxPool) FetchToBuild(ctx context.Context, height uint64) (ld.Txs, error) {
	errp := erring.ErrPrefix("chain.TxPool.FetchToBuild: ")

	params, err := p.genParams(ld.MaxBlockTxsSize, true)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	result := make([]txpool.TxOrBatch, 0)
	res := p.cli.Request(ctx, "fetchToBuild", params, &result)
	if res.Error != nil {
		return nil, errp.ErrorIf(res.Error)
	}

	txs := make(ld.Txs, 0, len(result))
	for _, v := range result {
		tx, err := v.ToTransaction()
		if err != nil {
			return nil, errp.ErrorIf(err)
		}
		txs = append(txs, tx)
	}

	if txs.Size() > ld.MaxBlockTxsSize {
		return nil, errp.Errorf("invalid txs size, expected <= %d, got %d", ld.MaxBlockTxsSize, txs.Size())
	}

	txIDs := txs.IDs()
	// sort and check
	txs.Sort()
	for i, id := range txs.IDs() {
		if id != txIDs[i] {
			return nil, errp.Errorf("invalid txs order, expected %s, got %s", id, txIDs[i])
		}
	}

	return txs, nil
}

func (p *TxPool) UpdateBuildStatus(ctx context.Context, height uint64, tbs *txpool.TxsBuildStatus) {
	params, err := p.genParams(tbs, true)
	if err == nil {
		res := p.cli.Request(ctx, "updateBuildStatus", params, nil)
		if res.Error != nil {
			err = res.Error
		}
	}

	if err != nil {
		logging.Log.Warn("chain.TxPool.UpdateBuildStatus",
			zap.Uint64("height", height),
			zap.Error(err))
	}
}

func (p *TxPool) AcceptByBlock(ctx context.Context, blk *ld.Block) error {
	params, err := p.genParams(blk, true)
	if err == nil {
		res := p.cli.Request(ctx, "acceptByBlock", params, nil)
		if res.Error != nil {
			err = res.Error
		}
	}

	if err != nil {
		logging.Log.Warn("chain.TxPool.AcceptByBlock",
			zap.Uint64("height", blk.Height),
			zap.Error(err))
	}
	p.freeCache(blk.Height)
	return err
}

func (p *TxPool) loadByIDsFromCache(height uint64, txIDs ids.IDList[ids.ID32]) (txs ld.Txs, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if txs, ok = p.cache[height]; ok {
		if !txs.IDs().Equal(txIDs) {
			delete(p.cache, height)
			ok = false
		}
	}

	return
}

func (p *TxPool) SetCache(height uint64, txs ld.Txs) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cache[height] = txs
}

func (p *TxPool) freeCache(acceptedHeight uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k := range p.cache {
		if k <= acceptedHeight {
			delete(p.cache, k)
		}
	}
}
