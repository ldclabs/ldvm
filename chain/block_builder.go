// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math"
	"time"

	"github.com/ldclabs/ldvm/ld"
)

func BuildBlock(txPool TxPool, preferred *Block) (*ld.Block, error) {
	ts := time.Now().UTC()
	if ts.Before(preferred.Timestamp()) {
		ts = preferred.Timestamp()
	}

	blk := &ld.Block{
		Parent:    preferred.ID(),
		Height:    preferred.Height() + 1,
		Timestamp: uint64(ts.Unix()),
		Miners:    txPool.PopMiners(preferred.ID(), preferred.Height()),
		Txs:       make([]*ld.Transaction, 0, 16),
	}

	if mintTx := txPool.DecisionMintTx(preferred.ID()); mintTx != nil {
		blk.Txs = append(blk.Txs, mintTx)
	}

	gas := uint64(0)
	feeConfig := preferred.State().FeeConfig()
	txs := txPool.PopBySize(feeConfig.MaxBlockSize - 1000)
	for i := range txs {
		requireGas := txs[i].RequireGas(feeConfig.ThresholdGas)
		if requireGas > txs[i].GasFeeCap || requireGas > feeConfig.MaxTxGas {
			txPool.MarkDropped(txs[i].ID())
			continue
		}
		txs[i].Gas = requireGas + txs[i].GasTip
		if txs[i].Gas > feeConfig.MaxTxGas {
			txs[i].Gas = feeConfig.MaxTxGas
		}
		if err := txs[i].SyntacticVerify(); err != nil {
			txPool.MarkDropped(txs[i].ID())
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
	if txPool.Len() > len(blk.Txs) {
		blk.GasPrice = uint64(float64(blk.GasPrice) * math.SqrtPhi)
		if blk.GasPrice > feeConfig.MaxGasPrice {
			blk.GasPrice = feeConfig.MaxGasPrice
		}
	} else if txPool.Len() == 0 {
		blk.GasPrice = uint64(float64(blk.GasPrice) / math.SqrtPhi)
		if blk.GasPrice < feeConfig.MinGasPrice {
			blk.GasPrice = feeConfig.MinGasPrice
		}
	}

	return blk, nil
}
