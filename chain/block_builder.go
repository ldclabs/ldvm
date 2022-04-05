// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type BlockBuiler struct {
	ld *ld.Block
}

func NewBlockBuiler(preferred *Block) *BlockBuiler {
	ts := time.Now().UTC()
	if ts.Before(preferred.Timestamp()) {
		ts = preferred.Timestamp()
	}

	blk := &BlockBuiler{
		ld: &ld.Block{
			Parent:    preferred.ID(),
			Height:    preferred.Height() + 1,
			Timestamp: uint64(ts.Unix()),
			Miners:    make([]ids.ShortID, 0, 16),
			Txs:       make([]*ld.Transaction, 0, 16),
		},
	}
	blk.initGas(preferred)
	return blk
}

func (bb *BlockBuiler) initGas(preferred *Block) {

}

func (bb *BlockBuiler) AddTxs(txs ...*ld.Transaction) (bool, error) {
	return false, nil
}

func (bb *BlockBuiler) Build() (*Block, error) {
	return NewBlock(bb.ld)
}
