// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
)

const futureBound = 10 * time.Second

type Block struct {
	Parent    ids.ID // The genesis block's parent ID is ids.Empty.
	Height    uint64 // The genesis block is at 0.
	Timestamp uint64 // The genesis block is at 0.
	Gas       uint64 // This block's total gas.
	GasPrice  uint64 // This block's gas price
	// Gas rebate rate received by this block's miners, 0 ~ 1000, equal to 0～10 times.
	GasRebateRate uint64
	// The address of miners awarded in this block.
	// Miners can issue a TxMintFee transaction to apply for mining awards
	// after the completion and consensus of the parent block.
	// The first miners to reach new block can be entered, up to 128, sorted by ID.
	// Total gas rebate = Gas * GasRebateRate * GasPrice / 100
	Miners []ids.ShortID
	Txs    []*Transaction
	RawTxs []json.RawMessage
	id     ids.ID
	raw    []byte // the block's raw bytes
}

type jsonBlock struct {
	ID            string            `json:"id"`
	Parent        string            `json:"parent"`
	Height        uint64            `json:"height"`
	Timestamp     uint64            `json:"timestamp"`
	Gas           uint64            `json:"gas"`
	GasPrice      uint64            `json:"gasPrice"`
	GasRebateRate uint64            `json:"gasRebateRate"`
	Miners        []string          `json:"miners"`
	Txs           []json.RawMessage `json:"txs"`
}

func (b *Block) MarshalJSON() ([]byte, error) {
	if b == nil {
		return Null, nil
	}
	v := &jsonBlock{
		ID:            b.id.String(),
		Parent:        b.Parent.String(),
		Height:        b.Height,
		Timestamp:     b.Timestamp,
		Gas:           b.Gas,
		GasPrice:      b.GasPrice,
		GasRebateRate: b.GasRebateRate,
		Miners:        make([]string, len(b.Miners)),
		Txs:           b.RawTxs,
	}
	for i := range b.Miners {
		v.Miners[i] = EthID(b.Miners[i]).String()
	}
	if b.RawTxs == nil {
		v.Txs = make([]json.RawMessage, len(b.Txs))
		for i := range b.Txs {
			d, err := b.Txs[i].MarshalJSON()
			if err != nil {
				return nil, err
			}
			v.Txs[i] = d
		}
	}
	return json.Marshal(v)
}

func (b *Block) ID() ids.ID {
	return b.id
}

func (b *Block) Copy() *Block {
	x := new(Block)
	*x = *b
	x.Miners = make([]ids.ShortID, len(b.Miners))
	copy(x.Miners, x.Miners)
	x.Txs = make([]*Transaction, len(b.Txs))
	for i := range b.Txs {
		x.Txs[i] = b.Txs[i].Copy()
	}
	x.raw = make([]byte, len(b.raw))
	copy(x.raw, b.raw)
	return x
}

// SyntacticVerify verifies that a *Block is well-formed.
func (b *Block) SyntacticVerify() error {
	if b.Timestamp > uint64(time.Now().Add(futureBound).Unix()) {
		return fmt.Errorf("invalid block timestamp")
	}
	if b.GasRebateRate > 1000 {
		return fmt.Errorf("invalid block gasRebateRate")
	}
	for _, a := range b.Miners {
		if a == ids.ShortEmpty {
			return fmt.Errorf("invalid block miner address")
		}
	}
	if len(b.Txs) == 0 {
		return fmt.Errorf("invalid block no txs")
	}
	for _, tx := range b.Txs {
		if tx == nil {
			return fmt.Errorf("invalid block transaction")
		}
		if err := tx.SyntacticVerify(); err != nil {
			return fmt.Errorf("invalid block transaction: %v", err)
		}
	}
	if _, err := b.Marshal(); err != nil {
		return fmt.Errorf("block marshal error: %v", err)
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
	if len(o.Miners) != len(b.Miners) {
		return false
	}
	for i := range b.Miners {
		if o.Miners[i] != b.Miners[i] {
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
		if _, err := b.Marshal(); err != nil {
			panic(err)
		}
	}

	return b.raw
}

func (b *Block) Unmarshal(data []byte) error {
	p, err := blockLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindBlock); ok {
		b.Height = v.Height.Value()
		b.Timestamp = v.Timestamp.Value()
		b.Gas = v.Gas.Value()
		b.GasPrice = v.GasPrice.Value()
		b.GasRebateRate = v.GasRebateRate.Value()
		b.Txs = make([]*Transaction, len(v.Txs))
		for i := range v.Txs {
			tx := &Transaction{}
			if err := tx.Unmarshal(v.Txs[i]); err != nil {
				return err
			}
			b.Txs[i] = tx
		}
		if b.Parent, err = ToID(v.Parent); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if b.Miners, err = ToShortIDs(v.Miners); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		b.raw = data
		b.id = IDFromBytes(data)
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindBlock")
}

func (b *Block) Marshal() ([]byte, error) {
	v := &bindBlock{
		Parent:        FromID(b.Parent),
		Height:        FromUint64(b.Height),
		Timestamp:     FromUint64(b.Timestamp),
		Gas:           FromUint64(b.Gas),
		GasPrice:      FromUint64(b.GasPrice),
		GasRebateRate: FromUint64(b.GasRebateRate),
		Miners:        FromShortIDs(b.Miners),
		Txs:           make([][]byte, len(b.Txs)),
	}
	for i := range b.Txs {
		data, err := b.Txs[i].Marshal()
		if err != nil {
			return nil, err
		}
		v.Txs[i] = data
	}
	data, err := blockLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	b.raw = data
	b.id = IDFromBytes(data)
	return data, nil
}

type bindBlock struct {
	Parent        []byte
	Height        Uint64
	Timestamp     Uint64
	Gas           Uint64
	GasPrice      Uint64
	GasRebateRate Uint64
	Miners        [][]byte
	Txs           [][]byte
}

var blockLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint64 bytes
	type ID20 bytes
	type ID32 bytes
	type Sig65 bytes
	type Block struct {
		Parent        ID32    (rename "p")
		Height        Uint64  (rename "h")
		Timestamp     Uint64  (rename "ts")
		Gas           Uint64  (rename "g")
		GasPrice      Uint64  (rename "gp")
		GasRebateRate Uint64  (rename "gr")
		Miners        [ID20]  (rename "ms")
		Txs           [Bytes] (rename "txs")
	}
`
	builder, err := NewLDBuilder("Block", []byte(sch), (*bindBlock)(nil))
	if err != nil {
		panic(err)
	}
	blockLDBuilder = builder
}
