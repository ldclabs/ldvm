// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"

	avaids "github.com/ava-labs/avalanchego/ids"
	"go.uber.org/zap"
)

type TxPool struct {
	mu     sync.Mutex
	cli    *httprpc.CBORClient
	nodeID avaids.NodeID
	cache  map[uint64]ld.Txs // cached processing Txs
}

func NewTxPool(ctx *Context, pdsEndpoint string, rt http.RoundTripper) *TxPool {
	header := http.Header{}
	header.Set("x-node-id", ctx.NodeID.String())
	header.Set("user-agent",
		strings.Replace(ctx.Name(), "@v", "/", 1)+" (TxPool Client, CBOR-RPC)")

	return &TxPool{
		cli:    httprpc.NewCBORClient(pdsEndpoint+"/txs", rt, header),
		nodeID: ctx.NodeID,
		cache:  make(map[uint64]ld.Txs, 10),
	}
}

// TODO: Authorization
type TxReqParams struct {
	NodeID avaids.NodeID `cbor:"n"`
	Height uint64        `cbor:"h"`
	Params interface{}   `cbor:"p,omitempty"`
}

type TxsBuildStatus struct {
	Unknown    ids.IDList[ids.ID32] `cbor:"0,omitempty"`
	Processing ids.IDList[ids.ID32] `cbor:"1,omitempty"`
	Rejected   ids.IDList[ids.ID32] `cbor:"2,omitempty"`
	Accepted   ids.IDList[ids.ID32] `cbor:"3,omitempty"`
}

type TxOrBatch struct {
	Tx           *ld.TxData  `cbor:"tx,omitempty"`
	Signatures   signer.Sigs `cbor:"ss,omitempty"`
	ExSignatures signer.Sigs `cbor:"es,omitempty"`
	Batch        ld.Txs      `cbor:"ba,omitempty"`
}

func (t *TxOrBatch) toTransaction() (*ld.Transaction, error) {
	switch {
	case t.Tx != nil:
		tx := &ld.Transaction{Tx: *t.Tx, Signatures: t.Signatures, ExSignatures: t.ExSignatures}
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
		return tx, nil

	case len(t.Batch) > 1:
		return ld.NewBatchTx(t.Batch...)

	default:
		return nil, fmt.Errorf("invalid TxOrBatch")
	}
}

func (p *TxPool) LoadByIDs(height uint64, txIDs ids.IDList[ids.ID32]) (ld.Txs, error) {
	errp := erring.ErrPrefix("chain.TxPool.LoadByIDs: ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txs, ok := p.loadByIDsFromCache(height, txIDs)
	if !ok {
		txs = make(ld.Txs, 0, len(txIDs))
		params := &TxReqParams{NodeID: p.nodeID, Height: height, Params: txIDs}
		res := p.cli.Request(ctx, "LoadByIDs", params, &txs)
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

func (p *TxPool) SizeToBuild(height uint64) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	params := &TxReqParams{NodeID: p.nodeID, Height: height}
	size := 0
	res := p.cli.Request(ctx, "SizeToBuild", params, &size)
	if res.Error != nil {
		logging.Log.Warn("TxPool.SizeToBuild",
			zap.Uint64("height", height),
			zap.Error(res.Error))
	}
	return size
}

func (p *TxPool) FetchToBuild(height uint64) (ld.Txs, error) {
	errp := erring.ErrPrefix("chain.TxPool.FetchToBuild: ")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make([]TxOrBatch, 0)
	params := &TxReqParams{NodeID: p.nodeID, Height: height, Params: ld.MaxBlockTxsSize}
	res := p.cli.Request(ctx, "FetchToBuild", params, &result)
	if res.Error != nil {
		return nil, errp.ErrorIf(res.Error)
	}

	txs := make(ld.Txs, 0, len(result))
	for _, v := range result {
		tx, err := v.toTransaction()
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

func (p *TxPool) UpdateBuildStatus(height uint64, tbs *TxsBuildStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	params := &TxReqParams{NodeID: p.nodeID, Height: height, Params: tbs}
	res := p.cli.Request(ctx, "UpdateBuildStatus", params, nil)
	if res.Error != nil {
		logging.Log.Warn("TxPool.UpdateBuildStatus",
			zap.Uint64("height", height),
			zap.Error(res.Error))
	}

	if len(tbs.Accepted) > 0 {
		p.freeCache(height)
	}
}

func (p *TxPool) loadByIDsFromCache(height uint64, txIDs ids.IDList[ids.ID32]) (txs ld.Txs, ok bool) {
	p.mu.Lock()
	if txs, ok = p.cache[height]; ok {
		if !txs.IDs().Equal(txIDs) {
			delete(p.cache, height)
			ok = false
		}
	}
	p.mu.Unlock()
	return
}

func (p *TxPool) SetCache(height uint64, txs ld.Txs) {
	p.mu.Lock()
	p.cache[height] = txs
	p.mu.Unlock()
}

func (p *TxPool) freeCache(acceptedHeight uint64) {
	p.mu.Lock()
	for k := range p.cache {
		if k <= acceptedHeight {
			delete(p.cache, k)
		}
	}
	p.mu.Unlock()
}
