// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"time"

	"github.com/ldclabs/ldvm/util"
)

const (
	futureBound = 10 * time.Second
)

type Block struct {
	Parent    util.Hash `cbor:"p" json:"parent"`     // The genesis block's parent ID is ids.Empty.
	Height    uint64    `cbor:"h" json:"height"`     // The genesis block is at 0.
	Timestamp uint64    `cbor:"ts" json:"timestamp"` // The genesis block is at 0.
	State     util.Hash `cbor:"s" json:"state"`
	Gas       uint64    `cbor:"g" json:"gas"`       // This block's total gas units.
	GasPrice  uint64    `cbor:"gp" json:"gasPrice"` // This block's gas price
	// Gas rebate rate received by this block's miners, 0 ~ 1000, equal to 0～10 times.
	GasRebateRate uint64 `cbor:"gr" json:"gasRebateRate"`
	// The address of validator (convert to valid StakeAccount) who build this block.
	// All tips and 20% of total gas rebate are distributed to this stakeAccount.
	// Total gas rebate = Gas * GasRebateRate * GasPrice / 100
	Miner util.StakeSymbol `cbor:"mn" json:"miner"`
	// All validators (convert to valid StakeAccounts), sorted by Stake Balance.
	// 80% of total gas rebate are distributed to these stakeAccounts
	Validators []util.StakeSymbol `cbor:"vs" json:"validators"`
	PCHeight   uint64             `cbor:"ph" json:"pChainHeight"` // AVAX P Chain Height
	Txs        Txs                `cbor:"txs" json:"-"`

	// external assignment fields
	ID     util.Hash         `cbor:"-" json:"id"`
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
	errp := util.ErrPrefix("Block.SyntacticVerify: ")

	switch {
	case b == nil:
		return errp.Errorf("nil pointer")

	case b.Height > 0 && b.Parent == util.HashEmpty:
		return errp.Errorf("invalid parent %s", b.Parent)

	case b.State == util.HashEmpty:
		return errp.Errorf("invalid state %s", b.State)

	case b.Timestamp > uint64(time.Now().Add(futureBound).Unix()):
		return errp.Errorf("invalid timestamp")

	case b.GasPrice < 42:
		return errp.Errorf("invalid gasPrice")

	case b.GasRebateRate > 1000:
		return errp.Errorf("invalid gasRebateRate")

	case b.Miner != util.StakeEmpty && !b.Miner.Valid():
		return errp.Errorf("invalid miner address %s", b.Miner.GoString())

	case len(b.Validators) > 256:
		return errp.Errorf("too many validators")

	case len(b.Txs) == 0:
		return errp.Errorf("no txs")
	}

	for _, s := range b.Validators {
		if !s.Valid() {
			return errp.Errorf("invalid validator address %s", s.GoString())
		}
	}

	var err error
	gas := uint64(0)
	for _, tx := range b.Txs {
		if err = tx.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}
		gas += tx.Gas()
	}

	if b.Gas != gas {
		return errp.Errorf("invalid gas, expected %d, got %d", gas, b.Gas)
	}
	if b.raw, err = b.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}

	b.ID = util.HashFromData(b.raw)
	return nil
}

func (b *Block) Bytes() []byte {
	if len(b.raw) == 0 {
		b.raw = MustMarshal(b)
	}
	return b.raw
}

func (b *Block) Unmarshal(data []byte) error {
	return util.ErrPrefix("Block.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, b))
}

func (b *Block) Marshal() ([]byte, error) {
	return util.ErrPrefix("Block.Marshal: ").
		ErrorMap(util.MarshalCBOR(b))
}
