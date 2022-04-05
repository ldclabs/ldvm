// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/utils/hashing"

	"github.com/ldclabs/ldvm/ld"
)

var (
	_ snowman.Block = &Block{}
)

// Block is a possible decision that dictates the next canonical block.
//
// Blocks are guaranteed to be Verified, Accepted, and Rejected in topological
// order. Specifically, if Verify is called, then the parent has already been
// verified. If Accept is called, then the parent has already been accepted. If
// Reject is called, the parent has already been accepted or rejected.
//
// If the status of the block is Unknown, ID is assumed to be able to be called.
// If the status of the block is Accepted or Rejected; Parent, Verify, Accept,
// and Reject will never be called.
type Block struct {
	ld     *ld.Block
	id     ids.ID
	sb     StateBlock
	status choices.Status
	txs    []Transaction
}

func ParseBlock(data []byte) (*Block, error) {
	blk := &ld.Block{}
	if err := blk.Unmarshal(data); err != nil {
		return nil, err
	}
	return NewBlock(blk)
}

// tx gas, block gas, block miners should be set up !!!
func NewBlock(ld *ld.Block) (*Block, error) {
	blk := &Block{ld: ld, status: choices.Processing}
	if err := blk.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	blk.txs = make([]Transaction, len(ld.Txs))
	for i := range ld.Txs {
		tx, err := NewTx(ld.Txs[i], false)
		if err != nil {
			return nil, err
		}
		blk.txs[i] = tx
	}
	blk.id = hashing.ComputeHash256Array(blk.ld.Bytes())
	return blk, nil
}

func (b *Block) InitState(sb StateBlock, preferred *Block) {
	b.sb = sb
	if preferred != nil && b.Height() <= preferred.Height() { // history block
		b.status = choices.Accepted
	}
}

func (b *Block) State() StateBlock { return b.sb }

// ID implements the snowman.Block choices.Decidable ID interface
// ID returns a unique ID for this element.
func (b *Block) ID() ids.ID { return b.id }

// Accept implements the snowman.Block choices.Decidable Accept interface
// This element will be accepted by every correct node in the network.
// Accept sets this block's status to Accepted and sets lastAccepted to this
// block's ID and saves this info to stateDB,
// but not to underlying state db, and not update last_accepted_key. (SetPreference do it)
func (b *Block) Accept() error {
	for i := range b.txs {
		if err := b.txs[i].Accept(b); err != nil {
			return err
		}
	}
	// TODO: miners fee
	if err := b.sb.SaveBlock(b); err != nil {
		return err
	}
	b.status = choices.Accepted
	return b.sb.SetLastAccepted(b)
}

// Reject implements the snowman.Block choices.Decidable Reject interface
// This element will not be accepted by any correct node in the network.
func (b *Block) Reject() error {
	b.status = choices.Rejected
	b.sb.GivebackTxs(b.ld.Txs...)
	b.clear()
	return nil
}

// Status implements the snowman.Block choices.Decidable Status interface
// Status returns this element's current status.
// If Accept has been called on an element with this ID, Accepted should be
// returned. Similarly, if Reject has been called on an element with this
// ID, Rejected should be returned. If the contents of this element are
// unknown, then Unknown should be returned. Otherwise, Processing should be
// returned.
func (b *Block) Status() choices.Status { return b.status }

// Parent implements the snowman.Block Parent interface
// Parent returns the ID of this block's parent.
func (b *Block) Parent() ids.ID { return b.ld.Parent }

// Verify implements the snowman.Block Verify interface
// Verify that the state transition this block would make if accepted is valid.
// It is guaranteed that the Parent has been successfully verified.
func (b *Block) Verify() error {
	b.status = choices.Processing
	parent := b.sb.PreferredBlock()

	if b.Parent() != parent.ID() {
		return fmt.Errorf("invalid parent block, expected %v, got %v", parent.ID(), b.Parent())
	}

	if b.Height() != parent.Height()+1 {
		return fmt.Errorf("invalid block height, expected %v, got %v",
			parent.Height()+1, b.Height())
	}

	if b.Timestamp().Unix() < parent.Timestamp().Unix() {
		return fmt.Errorf("invalid block timestamp, too early")
	}

	gas := uint64(0)
	for i := range b.ld.Txs {
		if err := b.txs[i].Verify(b); err != nil {
			return err
		}
		gas += b.ld.Txs[i].Gas
	}
	if gas != b.ld.Gas {
		return fmt.Errorf("invalid block gas")
	}
	return nil
}

// Bytes implements the snowman.Block Bytes interface
// Bytes returns the binary representation of this block.
// This is used for sending blocks to peers. The bytes should be able to be
// parsed into the same block on another node.
func (b *Block) Bytes() []byte { return b.ld.Bytes() }

// Height implements the snowman.Block Height interface
// Height returns this block's height. The genesis block has height 0.
func (b *Block) Height() uint64 { return b.ld.Height }

// Timestamp implements the snowman.Block Timestamp interface
// Timestamp returns this block's time. The genesis block has timestamp 0.
func (b *Block) Timestamp() time.Time { return time.Unix(int64(b.ld.Timestamp), 0).UTC() }

func (b *Block) GasPrice() *big.Int {
	return new(big.Int).SetUint64(b.ld.GasPrice)
}

func (b *Block) GasRebateRate() *big.Int {
	return new(big.Int).SetUint64(b.ld.GasRebateRate)
}

func (b *Block) clear() {
	b.ld = nil
	b.sb = nil
}
