// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type Genesis struct {
	ChainConfig ChainConfig `json:"chainConfig"`
	FeeConfig   FeeConfig   `json:"feeConfig"`
	Alloc       map[ld.EthID]struct {
		Balance   *big.Int   `json:"balance"`
		Threshold uint8      `json:"threshold"`
		Guardians []ld.EthID `json:"guardians"`
	} `json:"alloc"`
	Block struct {
		GasRebateRate uint64 `json:"gasRebateRate"`
		GasPrice      uint64 `json:"gasPrice"`
	} `json:"block"`
}

type ChainConfig struct {
	ChainID        uint64    `json:"chainId"`
	MaxTotalSupply *big.Int  `json:"maxTotalSupply"`
	ConfigID       ld.DataID `json:"ConfigID"`
}

type FeeConfig struct {
	MinGasFee      uint64   `json:"minGasFee"`
	MinGasPrice    uint64   `json:"minGasPrice"`
	GasRebateRate  uint64   `json:"gasRebateRate"`
	MaxBlockGas    uint64   `json:"maxBlockGas"`
	MaxTxGas       uint64   `json:"maxTxGas"`
	MaxBlockMiners uint64   `json:"maxBlockMiners"`
	MinMinerStake  *big.Int `json:"minMinerStake"`
}

func FromJSON(data []byte) (*Genesis, error) {
	g := new(Genesis)
	if err := json.Unmarshal(data, g); err != nil {
		return nil, err
	}
	g.ChainConfig.ConfigID = ld.DataIDFromData(
		[]byte("LDVM"), ld.FromUint64(g.ChainConfig.ChainID))
	return g, nil
}

func (g *Genesis) ToBlock() (*ld.Block, error) {
	txs := make([]*ld.Transaction, 0)

	// Alloc Txs
	for k, v := range g.Alloc {
		tx := &ld.Transaction{
			Type:         ld.TypeTransfer,
			ChainID:      g.ChainConfig.ChainID,
			AccountNonce: uint64(0),
			Amount:       v.Balance,
			To:           ids.ShortID(k),
		}
		txs = append(txs, tx)

		if le := len(v.Guardians); le > 0 {
			update := &ld.TxUpdateAccountGuardians{
				Threshold: v.Threshold,
				Guardians: make([]ids.ShortID, len(v.Guardians)),
			}

			for i, id := range v.Guardians {
				if id == k {
					return nil, fmt.Errorf("address %s exists", id)
				}
				update.Guardians[i] = ids.ShortID(id)
			}

			if err := update.SyntacticVerify(); err != nil {
				return nil, err
			}

			tx := &ld.Transaction{
				Type:         ld.TypeUpdateAccountGuardians,
				ChainID:      g.ChainConfig.ChainID,
				AccountNonce: uint64(1),
				From:         ids.ShortID(k),
				Data:         update.Bytes(),
			}
			txs = append(txs, tx)
		}
	}
	// Config Txs

	// build genesis block
	blk := &ld.Block{
		Parent:        ids.Empty,
		Height:        0,
		Timestamp:     0,
		Gas:           0,
		GasPrice:      g.Block.GasPrice,
		GasRebateRate: g.Block.GasRebateRate,
		Txs:           txs,
	}
	if err := blk.SyntacticVerify(); err != nil {
		return nil, err
	}
	return blk, nil
}
