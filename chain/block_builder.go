// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"go.uber.org/zap"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
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
	nodeID          util.Address
	lastBuildHeight uint64
	status          builderStatus
	txPool          *TxPool
	toEngine        chan<- common.Message
}

func NewBlockBuilder(txPool *TxPool, toEngine chan<- common.Message) *BlockBuilder {
	return &BlockBuilder{
		nodeID:   util.Address(txPool.nodeID),
		txPool:   txPool,
		toEngine: toEngine,
	}
}

// HandlePreferenceBlock should be called immediately after [VM.SetPreference].
func (b *BlockBuilder) HandlePreferenceBlock(height uint64) {
	size := b.txPool.SizeToBuild(height + 1)

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

func (b *BlockBuilder) Build(ctx *Context) (*Block, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	blk, err := b.build(ctx)

	if err != nil {
		b.status = dontBuild
		return nil, util.ErrPrefix("chain.BlockBuilder.Build: ").ErrorIf(err)
	}

	b.status = building
	return blk, nil
}

func (b *BlockBuilder) build(ctx *Context) (*Block, error) {
	ts := uint64(time.Now().UTC().Unix())
	preferred := ctx.Chain().PreferredBlock()
	parentHeight := preferred.Height()

	if b.lastBuildHeight > parentHeight {
		return nil, fmt.Errorf("wait lastBuildHeight %d becoming preferred, current at %d",
			b.lastBuildHeight, parentHeight)
	}

	pts := preferred.Timestamp2()
	if ts < pts {
		ts = pts
	}

	txs, err := b.txPool.FetchToBuild(parentHeight + 1)
	if err != nil {
		return nil, err
	}

	if size := txs.Size(); size == 0 || (ts == pts && size < minTxsWhenBuild) {
		time.AfterFunc(waitForMoreTxs, b.SignalTxsReady)
		return nil, fmt.Errorf("wait txs to build, expected >= %d, got %d", minTxsWhenBuild, size)
	}

	feeCfg := ctx.ChainConfig().Fee(parentHeight + 1)
	blk := &ld.Block{
		Parent:        util.Hash(preferred.ID()),
		Height:        parentHeight + 1,
		Timestamp:     ts,
		GasPrice:      preferred.NextGasPrice(),
		GasRebateRate: feeCfg.GasRebateRate,
		Txs:           util.NewIDList[util.Hash](txs.Size()),
	}

	nblk := NewBlock(blk, preferred.Context())
	nblk.InitState(preferred, preferred.State().VersionDB())
	vbs, err := nblk.State().DeriveState()
	if err != nil {
		return nil, err
	}

	// 1. SetBuilder
	nblk.SetBuilder(vbs)

	// 2. TryBuildTxs
	var status choices.Status
	tbs := &TxsBuildStatus{}
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
			tbs.Processing = append(tbs.Processing, tx.IDs()...)
			processingTxs = append(processingTxs, tx.Txs()...)
			vbs = nvbs
		}
	}

	go b.txPool.UpdateBuildStatus(blk.Height, tbs)
	if len(blk.Txs) == 0 {
		return nil, fmt.Errorf("no txs to build")
	}

	// 3. SetBuilderFee
	if err := nblk.SetBuilderFee(vbs); err != nil {
		return nil, err
	}

	// 4. BuildState and Verify block
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
