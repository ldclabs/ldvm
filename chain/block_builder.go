// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
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
	nodeID          util.EthID
	txPool          txPoolForBuilder
	lastBuildHeight uint64
	status          builderStatus
	toEngine        chan<- common.Message
}

type txPoolForBuilder interface {
	Len() int
	AddLocal(...*ld.Transaction)
	SetTxsStatus(choices.Status, ...ids.ID)
	PopTxsBySize(int) ld.Txs
	Reject(*ld.Transaction)
}

func NewBlockBuilder(nodeID ids.NodeID, txPool txPoolForBuilder, toEngine chan<- common.Message) *BlockBuilder {
	return &BlockBuilder{
		nodeID:   util.EthID(nodeID),
		txPool:   txPool,
		toEngine: toEngine,
	}
}

// HandlePreferenceBlock should be called immediately after [VM.SetPreference].
func (b *BlockBuilder) HandlePreferenceBlock() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.txPool.Len() > 0 {
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
		logging.Log.Debug("dropping message to consensus engine")
	}
}

func (b *BlockBuilder) Build(ctx *Context) (*Block, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	blk, err := b.build(ctx)
	if err != nil {
		b.status = dontBuild
		return nil, err
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

	pts := uint64(preferred.Timestamp().Unix())
	if ts < pts {
		ts = pts
	}
	if ts == pts && b.txPool.Len() < minTxsWhenBuild {
		time.AfterFunc(waitForMoreTxs, b.SignalTxsReady)
		return nil, fmt.Errorf("BlockBuilder.Build error: wait txs to build")
	}

	feeCfg := ctx.ChainConfig().Fee(parentHeight + 1)
	blk := &ld.Block{
		Parent:        preferred.ID(),
		Height:        parentHeight + 1,
		Timestamp:     ts,
		GasPrice:      preferred.NextGasPrice(),
		GasRebateRate: feeCfg.GasRebateRate,
		Txs:           make([]*ld.Transaction, 0, 16),
	}

	nblk := NewBlock(blk, preferred.Context())
	nblk.InitState(preferred, preferred.State().VersionDB())
	vbs, err := nblk.State().DeriveState()
	if err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}

	// 1. BuildMiner
	nblk.BuildMiner(vbs)

	// 2. TryBuildTxs
	var status choices.Status
	txs := b.txPool.PopTxsBySize(int(feeCfg.MaxBlockTxsSize))
	for len(txs) > 0 {
		for i := range txs {
			nvbs, err := vbs.DeriveState()
			if err != nil {
				return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
			}
			tx := txs[i]
			switch {
			case tx.IsBatched():
				status = nblk.BuildTxs(nvbs, tx.Txs()...)
			case tx.Type == ld.TypeTest:
				tx.Err = fmt.Errorf("BlockBuilder.Build error: TextTx should be in Batch Tx")
				status = choices.Rejected
			default:
				status = nblk.BuildTxs(nvbs, tx)
			}

			switch status {
			case choices.Unknown:
				b.txPool.AddLocal(tx)
			case choices.Rejected:
				b.txPool.Reject(tx)
			default:
				vbs = nvbs
				nblk.originTxs = append(nblk.originTxs, tx)
				b.txPool.SetTxsStatus(status, tx.ID)
			}
		}
		txs = b.txPool.PopTxsBySize(int(feeCfg.MaxBlockTxsSize) - nblk.TxsSize())
	}
	if len(blk.Txs) == 0 {
		return nil, fmt.Errorf("BlockBuilder.Build error: no txs to build")
	}

	// 3. BuildMinerFee
	if err := nblk.BuildMinerFee(vbs); err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}

	// 4. BuildState and Verify block
	if err := nblk.BuildState(vbs); err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}

	b.lastBuildHeight = blk.Height
	logging.Log.Info("Build block %s at %d", blk.ID, blk.Height)
	return nblk, nil
}
