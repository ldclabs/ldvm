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
)

const maxBuildInterval = 10 * time.Second

type BlockBuilder struct {
	mu              sync.RWMutex
	nodeID          ids.ShortID
	txPool          TxPool
	lastBuildHeight uint64
	lastBuildTime   time.Time
	timer           *time.Timer
	notifyBuild     func()
}

func NewBlockBuilder(nodeID ids.ShortID, txPool TxPool, notifyBuild func()) *BlockBuilder {
	return &BlockBuilder{
		nodeID:        nodeID,
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
	b.txPool.ClearMintTxs(parentHeight)
	mintTx, miners := b.txPool.SelectMiners(b.nodeID, preferred.ID(), parentHeight)
	blk := &ld.Block{
		Parent:        preferred.ID(),
		Height:        parentHeight + 1,
		Timestamp:     ts,
		GasRebateRate: feeCfg.GasRebateRate,
		Miners:        miners,
		Txs:           make([]*ld.Transaction, 0, 16),
	}

	if mintTx != nil {
		blk.Txs = append(blk.Txs, mintTx)
	}

	gas := uint64(0)
	txs := b.txPool.PopTxsBySize(int(feeCfg.MaxBlockTxsSize), feeCfg.ThresholdGas)
	for i := range txs {
		txs[i].Status = choices.Processing
		requireGas := txs[i].RequireGas(feeCfg.ThresholdGas)
		if requireGas > txs[i].GasFeeCap || requireGas > feeCfg.MaxTxGas {
			b.txPool.Rejecte(txs[i])
			continue
		}
		txs[i].Gas = requireGas + txs[i].GasTip
		if txs[i].Gas > feeCfg.MaxTxGas {
			txs[i].Gas = feeCfg.MaxTxGas
		}
		if err := txs[i].SyntacticVerify(); err != nil {
			b.txPool.Rejecte(txs[i])
			continue
		}
		gas += txs[i].Gas
		blk.Txs = append(blk.Txs, txs[i])
	}

	if len(blk.Txs) == 0 {
		return nil, fmt.Errorf("no txs to build")
	}

	blk.Gas = gas
	blk.GasPrice = preferred.GasPrice().Uint64()
	if b.txPool.Len() > len(blk.Txs) {
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

	nblk, err := NewBlock(blk, preferred.Context())
	if err != nil {
		return nil, err
	}

	nblk.InitState(preferred.State().VersionDB(), false)
	b.lastBuildHeight = blk.Height
	logging.Log.Info("Build block %s at %d", blk.ID(), blk.Height)
	return nblk, nil
}
