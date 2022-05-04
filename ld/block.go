// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

const (
	futureBound = 10 * time.Second
)

type Block struct {
	Parent    ids.ID `cbor:"p" json:"parent"`     // The genesis block's parent ID is ids.Empty.
	Height    uint64 `cbor:"h" json:"height"`     // The genesis block is at 0.
	Timestamp uint64 `cbor:"ts" json:"timestamp"` // The genesis block is at 0.
	Gas       uint64 `cbor:"g" json:"gas"`        // This block's total gas units.
	GasPrice  uint64 `cbor:"gp" json:"gasPrice"`  // This block's gas price
	// Gas rebate rate received by this block's miners, 0 ~ 1000, equal to 0～10 times.
	GasRebateRate uint64 `cbor:"gr" json:"gasRebateRate"`
	// The address of validator (convert to valid StakeAccount) who build this block.
	// All tips and 20% of total gas rebate are distributed to this stakeAccount.
	// Total gas rebate = Gas * GasRebateRate * GasPrice / 100
	Miner util.EthID `cbor:"mn" json:"miner"`
	// All validators (convert to valid StakeAccounts), sorted by Stake Balance.
	// 80% of total gas rebate are distributed to these stakeAccounts
	Shares []util.EthID   `cbor:"sh" json:"shares"`
	Txs    []*Transaction `cbor:"txs" json:"-"`

	// external assignment
	ID     ids.ID            `cbor:"-" json:"id"`
	RawTxs []json.RawMessage `cbor:"-" json:"txs"`
	raw    []byte            `cbor:"-" json:"-"` // the block's raw bytes
}

func (b *Block) Copy() *Block {
	x := new(Block)
	*x = *b
	x.Shares = make([]util.EthID, len(b.Shares))
	copy(x.Shares, x.Shares)
	x.Txs = make([]*Transaction, len(b.Txs))
	for i := range b.Txs {
		x.Txs[i] = b.Txs[i].Copy()
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *Block is well-formed.
func (b *Block) SyntacticVerify() error {
	if b == nil {
		return fmt.Errorf("invalid Block")
	}

	if b.Timestamp > uint64(time.Now().Add(futureBound).Unix()) {
		return fmt.Errorf("invalid timestamp")
	}
	if b.GasRebateRate > 1000 {
		return fmt.Errorf("invalid gasRebateRate")
	}
	for _, a := range b.Shares {
		if a == util.EthIDEmpty {
			return fmt.Errorf("invalid miner address")
		}
	}
	if len(b.Txs) == 0 {
		return fmt.Errorf("invalid block, no txs")
	}
	for _, tx := range b.Txs {
		if tx == nil {
			return fmt.Errorf("invalid transaction")
		}
		if err := tx.SyntacticVerify(); err != nil {
			return fmt.Errorf("invalid transaction, SyntacticVerify error: %v", err)
		}
	}
	if _, err := b.Marshal(); err != nil {
		return fmt.Errorf("Block marshal error: %v", err)
	}
	return nil
}

func (b *Block) Equal(o *Block) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(b.raw) > 0 {
		return bytes.Equal(o.raw, b.raw)
	}
	if o.Parent != b.Parent {
		return false
	}
	if o.Height != b.Height {
		return false
	}
	if o.Timestamp != b.Timestamp {
		return false
	}
	if o.Gas != b.Gas {
		return false
	}
	if o.GasPrice != b.GasPrice {
		return false
	}
	if o.GasRebateRate != b.GasRebateRate {
		return false
	}
	if o.Miner != b.Miner {
		return false
	}
	if len(o.Shares) != len(b.Shares) {
		return false
	}
	for i := range b.Shares {
		if o.Shares[i] != b.Shares[i] {
			return false
		}
	}
	if len(o.Txs) != len(b.Txs) {
		return false
	}
	for i := range b.Txs {
		if !o.Txs[i].Equal(b.Txs[i]) {
			return false
		}
	}
	return true
}

func (b *Block) Bytes() []byte {
	if len(b.raw) == 0 {
		MustMarshal(b)
	}
	return b.raw
}

func (b *Block) FeeCost() *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(b.Gas), new(big.Int).SetUint64(b.GasPrice))
}

func (b *Block) Unmarshal(data []byte) error {
	b.ID = util.IDFromBytes(data)
	b.raw = data
	return DecMode.Unmarshal(data, b)
}

func (b *Block) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(b)
	if err != nil {
		return nil, err
	}
	b.ID = util.IDFromBytes(data)
	b.raw = data
	return data, nil
}
