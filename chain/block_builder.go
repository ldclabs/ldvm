// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

const maxBuildInterval = 10 * time.Second

type BlockBuilder struct {
	mu              sync.RWMutex
	nodeID          util.EthID
	txPool          TxPool
	lastBuildHeight uint64
	lastBuildTime   time.Time
	timer           *time.Timer
	notifyBuild     func()
}

func NewBlockBuilder(nodeID ids.NodeID, txPool TxPool, notifyBuild func()) *BlockBuilder {
	return &BlockBuilder{
		nodeID:        util.EthID(nodeID),
		txPool:        txPool,
		notifyBuild:   notifyBuild,
		lastBuildTime: time.Now().UTC(),
		timer:         time.AfterFunc(maxBuildInterval, notifyBuild),
	}
}

func (b *BlockBuilder) NeedBuild() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ln := b.txPool.Len()
	du := time.Now().Sub(b.lastBuildTime)

	switch {
	case ln <= 2:
		return du >= maxBuildInterval
	case ln <= 5:
		return du >= maxBuildInterval/2
	case ln <= 10:
		return du >= maxBuildInterval/5
	default:
		return true
	}
}

func (b *BlockBuilder) Build(ctx *Context, preferred *Block) (*Block, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastBuildTime = time.Now().UTC()
	if !b.timer.Reset(maxBuildInterval) {
		b.timer = time.AfterFunc(maxBuildInterval, b.notifyBuild)
	}

	parentHeight := preferred.Height()
	if b.lastBuildHeight > parentHeight {
		return nil, fmt.Errorf("wait lastBuildHeight %d becoming preferred, current at %d",
			b.lastBuildHeight, parentHeight)
	}

	ts := uint64(b.lastBuildTime.Unix())
	if pt := uint64(preferred.Timestamp().Unix()); ts < pt {
		ts = pt
	}

	feeCfg := ctx.Chain().Fee(parentHeight + 1)
	blk := &ld.Block{
		Parent:        preferred.ID(),
		Height:        parentHeight + 1,
		Timestamp:     ts,
		GasRebateRate: feeCfg.GasRebateRate,
		Txs:           make([]*ld.Transaction, 0, 16),
	}

	txs := b.txPool.PopTxsBySize(int(feeCfg.MaxBlockTxsSize), feeCfg.ThresholdGas, ts)
	blk.GasPrice = preferred.GasPrice().Uint64()
	if b.txPool.Len() > len(txs) {
		blk.GasPrice = uint64(float64(blk.GasPrice) * math.SqrtPhi)
		if blk.GasPrice > feeCfg.MaxGasPrice {
			blk.GasPrice = feeCfg.MaxGasPrice
		}
	} else if b.txPool.Len() == 0 {
		blk.GasPrice = uint64(float64(blk.GasPrice) / math.SqrtPhi)
		if blk.GasPrice < feeCfg.MinGasPrice {
			blk.GasPrice = feeCfg.MinGasPrice
		}
	}

	nblk := NewBlock(blk, preferred.Context())
	nblk.InitState(preferred.State().VersionDB(), false)
	vbs, err := nblk.State().DeriveState()
	if err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}

	// 1. BuildMiner
	nblk.BuildMiner(vbs)
	// 2. TryBuildTxs
	var status choices.Status
	for len(txs) > 0 {
		for i := range txs {
			nvbs, err := vbs.DeriveState()
			if err != nil {
				return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
			}
			tx := txs[i]
			switch {
			case tx.Type == ld.TypeTest:
				tx.Err = fmt.Errorf("BlockBuilder.Build error: TextTx should be in Batch Tx")
				status = choices.Rejected
			case tx.IsBatched():
				status = nblk.BuildTxs(nvbs, tx.Txs()...)
			default:
				status = nblk.BuildTxs(nvbs, tx)
			}

			switch status {
			case choices.Unknown:
				b.txPool.Add(tx)
			case choices.Rejected:
				b.txPool.Reject(tx)
			default:
				vbs = nvbs
				nblk.originTxs = append(nblk.originTxs, tx)
			}
		}
		txs = b.txPool.PopTxsBySize(int(feeCfg.MaxBlockTxsSize)-nblk.TxsSize(), feeCfg.ThresholdGas, ts)
	}
	if len(blk.Txs) == 0 {
		return nil, fmt.Errorf("BlockBuilder.Build error: no txs to build")
	}

	// 3. BuildMinerFee
	if err := nblk.BuildMinerFee(vbs); err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}
	// 4. BuildState
	if err := nblk.BuildState(vbs); err != nil {
		return nil, fmt.Errorf("BlockBuilder.Build error: %v", err)
	}
	b.lastBuildHeight = blk.Height
	logging.Log.Info("Build block %s at %d", blk.ID, blk.Height)
	return nblk, nil
}
