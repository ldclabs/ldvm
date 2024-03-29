// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"go.uber.org/zap"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/txpool"
	"github.com/ldclabs/ldvm/util/erring"
)

const (
	minTxsWhenBuild = 2
	waitForMoreTxs  = 1 * time.Second
)

// builderStatus denotes the current status of the VM in block production.
type builderStatus uint8

const (
	dontBuild builderStatus = iota // no need to build a block.
	waitBuild                      // has sent a request to the engine to build a block.
	building                       // building a block.
)

type BlockBuilder struct {
	mu              sync.RWMutex
	lastBuildHeight uint64
	status          builderStatus
	ctx             *Context
	txPool          *TxPool
	toEngine        chan<- common.Message
}

func NewBlockBuilder(txPool *TxPool, toEngine chan<- common.Message) *BlockBuilder {
	return &BlockBuilder{
		ctx:      txPool.ctx,
		txPool:   txPool,
		toEngine: toEngine,
	}
}

// HandlePreferenceBlock should be called immediately after [VM.SetPreference].
func (b *BlockBuilder) HandlePreferenceBlock(ctx context.Context, height uint64) {
	size := b.txPool.SizeToBuild(ctx, height+1)

	b.mu.Lock()
	defer b.mu.Unlock()
	if size > 0 {
		b.markBuilding()
	} else {
		b.status = dontBuild
	}
}

// SignalTxsReady should be called immediately when a new tx incoming
func (b *BlockBuilder) SignalTxsReady() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.status == dontBuild {
		b.markBuilding()
	}
}

// signal the avalanchego engine to build a block from pending transactions
func (b *BlockBuilder) markBuilding() {
	select {
	case b.toEngine <- common.PendingTxs:
		b.status = waitBuild
	default:
		logging.Log.Debug("BlockBuilder.markBuilding: dropping message to consensus engine")
	}
}

func (b *BlockBuilder) Build(ctx context.Context, builder ids.Address) (*Block, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	blk, err := b.build(ctx, builder)

	if err != nil {
		b.status = dontBuild
		return nil, erring.ErrPrefix("chain.BlockBuilder.Build: ").ErrorIf(err)
	}

	b.status = building
	return blk, nil
}

func (b *BlockBuilder) build(ctx context.Context, builder ids.Address) (*Block, error) {
	ts := uint64(time.Now().UTC().Unix())
	preferred := b.ctx.Chain().PreferredBlock()
	parentHeight := preferred.Height()

	if b.lastBuildHeight > parentHeight {
		return nil, fmt.Errorf("wait lastBuildHeight %d becoming preferred, current at %d",
			b.lastBuildHeight, parentHeight)
	}

	pts := preferred.Timestamp2()
	if ts < pts {
		ts = pts
	}

	txs, err := b.txPool.FetchToBuild(ctx, parentHeight+1)
	if err != nil {
		return nil, err
	}

	if size := txs.Size(); size == 0 || (ts == pts && size < minTxsWhenBuild) {
		time.AfterFunc(waitForMoreTxs, b.SignalTxsReady)
		return nil, fmt.Errorf("wait txs to build, expected >= %d, got %d", minTxsWhenBuild, size)
	}

	feeCfg := b.ctx.ChainConfig().Fee(parentHeight + 1)
	blk := &ld.Block{
		Parent:        ids.ID32(preferred.ID()),
		Height:        parentHeight + 1,
		Timestamp:     ts,
		GasPrice:      preferred.NextGasPrice(),
		GasRebateRate: feeCfg.GasRebateRate,
		Builder:       builder,
		Txs:           ids.NewIDList[ids.ID32](txs.Size()),
	}

	nblk := NewBlock(blk, preferred.Context())
	nblk.InitState(preferred, preferred.State().VersionDB())
	vbs, err := nblk.State().DeriveState()
	if err != nil {
		return nil, err
	}

	// 1. TryBuildTxs
	var status choices.Status
	tbs := &txpool.TxsBuildStatus{}
	processingTxs := make(ld.Txs, 0, txs.Size())
	for i := range txs {
		nvbs, err := vbs.DeriveState()
		if err != nil {
			return nil, err
		}

		tx := txs[i]
		switch {
		case tx.IsBatched():
			status = nblk.BuildTxs(nvbs, tx.Txs()...)
		case tx.Tx.Type == ld.TypeTest:
			status = choices.Rejected
		default:
			status = nblk.BuildTxs(nvbs, tx)
		}

		switch status {
		case choices.Unknown:
			tbs.Unknown = append(tbs.Unknown, tx.IDs()...)
		case choices.Rejected:
			tbs.Rejected = append(tbs.Rejected, tx.IDs()...)
		default:
			processingTxs = append(processingTxs, tx.Txs()...)
			vbs = nvbs
		}
	}

	go b.txPool.UpdateBuildStatus(ctx, blk.Height, tbs)
	if len(blk.Txs) == 0 {
		return nil, fmt.Errorf("no txs to build")
	}

	// 2. SetBuilderFee
	if err := nblk.SetBuilderFee(vbs); err != nil {
		return nil, err
	}

	// 3. BuildState and Verify block
	if err := nblk.BuildState(vbs); err != nil {
		return nil, err
	}

	b.lastBuildHeight = blk.Height
	b.txPool.SetCache(blk.Height, processingTxs)
	logging.Log.Info("BlockBuilder.Build",
		zap.Stringer("parent", blk.Parent),
		zap.Stringer("id", blk.ID),
		zap.Uint64("height", blk.Height),
		zap.Int("txs", len(blk.Txs)))
	return nblk, nil
}
