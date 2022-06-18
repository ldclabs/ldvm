// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
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
	Parent      ids.ID `cbor:"p" json:"parent"`     // The genesis block's parent ID is ids.Empty.
	Height      uint64 `cbor:"h" json:"height"`     // The genesis block is at 0.
	Timestamp   uint64 `cbor:"ts" json:"timestamp"` // The genesis block is at 0.
	ParentState ids.ID `cbor:"ps" json:"parentState"`
	State       ids.ID `cbor:"s" json:"state"`
	Gas         uint64 `cbor:"g" json:"gas"`       // This block's total gas units.
	GasPrice    uint64 `cbor:"gp" json:"gasPrice"` // This block's gas price
	// Gas rebate rate received by this block's miners, 0 ~ 1000, equal to 0～10 times.
	GasRebateRate uint64 `cbor:"gr" json:"gasRebateRate"`
	// The address of validator (convert to valid StakeAccount) who build this block.
	// All tips and 20% of total gas rebate are distributed to this stakeAccount.
	// Total gas rebate = Gas * GasRebateRate * GasPrice / 100
	Miner util.StakeSymbol `cbor:"mn" json:"miner"`
	// All validators (convert to valid StakeAccounts), sorted by Stake Balance.
	// 80% of total gas rebate are distributed to these stakeAccounts
	Validators []util.StakeSymbol `cbor:"vs" json:"validators"`
	PCHeight   uint64             `cbor:"ph" json:"PChainHeight"`
	Txs        Txs                `cbor:"txs" json:"-"`

	// external assignment fields
	ID     ids.ID            `cbor:"-" json:"id"`
	RawTxs []json.RawMessage `cbor:"-" json:"txs"`
	raw    []byte            `cbor:"-" json:"-"` // the block's raw bytes
}

func (b *Block) TxsMarshalJSON() error {
	b.RawTxs = make([]json.RawMessage, len(b.Txs))
	for i, tx := range b.Txs {
		d, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		b.RawTxs[i] = d
	}
	return nil
}

// SyntacticVerify verifies that a *Block is well-formed.
func (b *Block) SyntacticVerify() error {
	errPrefix := "Block.SyntacticVerify failed:"

	switch {
	case b == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case b.Height > 0 && b.Parent == ids.Empty:
		return fmt.Errorf("%s invalid parent %s", errPrefix, b.Parent)

	case b.Height > 0 && b.ParentState == ids.Empty:
		return fmt.Errorf("%s invalid parent state %s", errPrefix, b.ParentState)

	case b.State == ids.Empty || b.State == b.ParentState:
		return fmt.Errorf("%s invalid state %s", errPrefix, b.State)

	case b.Timestamp > uint64(time.Now().Add(futureBound).Unix()):
		return fmt.Errorf("%s invalid timestamp", errPrefix)

	case b.GasRebateRate > 1000:
		return fmt.Errorf("%s invalid gasRebateRate", errPrefix)

	case b.Miner != util.StakeEmpty && !b.Miner.Valid():
		return fmt.Errorf("%s invalid miner address %s", errPrefix, b.Miner.GoString())

	case len(b.Validators) > 256:
		return fmt.Errorf("%s too many validators", errPrefix)

	case len(b.Txs) == 0:
		return fmt.Errorf("%s no txs", errPrefix)
	}

	for _, s := range b.Validators {
		if !s.Valid() {
			return fmt.Errorf("%s invalid validator address %s", errPrefix, s.GoString())
		}
	}

	var err error
	for _, tx := range b.Txs {
		if err = tx.SyntacticVerify(); err != nil {
			return fmt.Errorf("%s %v", errPrefix, err)
		}
	}
	if b.raw, err = b.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	b.ID = util.IDFromData(b.raw)
	return nil
}

func (b *Block) FeeCost() *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(b.Gas), new(big.Int).SetUint64(b.GasPrice))
}

func (b *Block) Bytes() []byte {
	if len(b.raw) == 0 {
		b.raw = MustMarshal(b)
	}
	return b.raw
}

func (b *Block) Unmarshal(data []byte) error {
	return UnmarshalCBOR(data, b)
}

func (b *Block) Marshal() ([]byte, error) {
	return MarshalCBOR(b)
}
