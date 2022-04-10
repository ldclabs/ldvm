// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"

	"github.com/ldclabs/ldvm/constants"
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
	bs     BlockState
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
func NewBlock(b *ld.Block) (*Block, error) {
	blk := &Block{ld: b, status: choices.Processing}
	if err := blk.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	blk.txs = make([]Transaction, len(b.Txs))
	for i := range b.Txs {
		tx, err := NewTx(b.Txs[i], false)
		if err != nil {
			return nil, err
		}
		blk.txs[i] = tx
	}
	return blk, nil
}

func newGenesisBlock(b *ld.Block) (*Block, error) {
	blk := &Block{ld: b, status: choices.Processing}
	if err := blk.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	blk.txs = make([]Transaction, len(b.Txs))
	for i := range b.Txs {
		tx, err := newGenesisTx(b.Txs[i])
		if err != nil {
			return nil, err
		}
		blk.txs[i] = tx
	}
	return blk, nil
}

func (b *Block) MarshalJSON() ([]byte, error) {
	txs := make([]json.RawMessage, len(b.txs))
	for i := range b.txs {
		d, err := b.txs[i].MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.bs.Log().Info("Tx: %s", string(d))
		txs[i] = d
	}
	b.ld.RawTxs = txs
	return b.ld.MarshalJSON()
}

func (b *Block) InitState(s *stateDB, db database.Database, preferred *Block) {
	b.bs = newBlockState(s, s.db, b.Height())
	if preferred != nil && b.Height() <= preferred.Height() { // history block
		b.status = choices.Accepted
	}
}

func (b *Block) State() BlockState { return b.bs }

// ID implements the snowman.Block choices.Decidable ID interface
// ID returns a unique ID for this element.
func (b *Block) ID() ids.ID { return b.ld.ID() }

func (b *Block) VerifyGenesis() error {
	for i := range b.ld.Txs {
		tx, ok := b.txs[i].(GenesisTx)
		if !ok {
			return fmt.Errorf("invalid genesis tx")
		}
		if err := tx.VerifyGenesis(b); err != nil {
			return err
		}
	}
	return nil
}

// Verify implements the snowman.Block Verify interface
// Verify that the state transition this block would make if accepted is valid.
// It is guaranteed that the Parent has been successfully verified.
func (b *Block) Verify() error {
	err := b.verify()
	if err != nil {
		b.bs.Log().Warn("Block.Verify error: %v", err)
	}
	return err
}

func (b *Block) verify() error {
	b.status = choices.Processing
	parent := b.bs.PreferredBlock()

	if b.ID() == ids.Empty {
		return fmt.Errorf("invalid block id")
	}

	if b.Parent() != parent.ID() {
		return fmt.Errorf("invalid block parent, expected %s, got %s",
			parent.ID(), b.Parent())
	}

	if b.Height() != parent.Height()+1 {
		return fmt.Errorf("invalid block height, expected %d, got %d",
			parent.Height()+1, b.Height())
	}

	if b.Timestamp().Unix() < parent.Timestamp().Unix() {
		return fmt.Errorf("invalid block timestamp, too early")
	}

	if uint64(len(b.Bytes())) > b.State().FeeConfig().MaxBlockSize {
		return fmt.Errorf("invalid block size: should not more than %d bytes",
			b.State().FeeConfig().MaxBlockSize)
	}

	miners := ids.NewShortSet(len(b.ld.Miners))
	miners.Add(b.ld.Miners...)
	if len(b.ld.Miners) != miners.Len() {
		return fmt.Errorf("invalid block miners")
	}

	gas := uint64(0)
	for i := range b.ld.Txs {
		if i > 1 && b.ld.Txs[i].Type == ld.TypeMintFee {
			return fmt.Errorf("invalid TypeMintFee tx: can only appear in the first place")
		}
		if err := b.txs[i].Verify(b); err != nil {
			return err
		}
		gas += b.ld.Txs[i].Gas
	}
	if gas != b.ld.Gas {
		return fmt.Errorf("invalid block gas, expected %d, got %d", gas, b.ld.Gas)
	}
	return nil
}

// Accept implements the snowman.Block choices.Decidable Accept interface
// This element will be accepted by every correct node in the network.
// Accept sets this block's status to Accepted and sets lastAccepted to this
// block's ID and saves this info to stateDB,
// but not to underlying state db, and not update last_accepted_key. (SetPreference do it)
func (b *Block) Accept() error {
	err := b.accept()
	if err != nil {
		b.bs.Log().Warn("Block.Accept error: %v", err)
	}
	return err
}

func (b *Block) accept() error {
	log := b.bs.Log()
	ts := b.Timestamp().Unix()
	for i := range b.txs {
		log.Info("Accept Block txs %d, %s", i, b.txs[i].ID())
		if err := b.txs[i].Accept(b); err != nil {
			return err
		}
		b.bs.AddEvent(b.txs[i].Event(ts))
	}
	log.Info("Accept Block mintFee")

	if err := b.mintFee(); err != nil {
		return fmt.Errorf("set mint fee failed: %v", err)
	}
	log.Info("Accept Block SaveBlock")
	if err := b.bs.SaveBlock(b); err != nil {
		return fmt.Errorf("save block failed: %v", err)
	}

	log.Info("Accept Block SetLastAccepted")
	b.status = choices.Accepted
	if err := b.bs.SetLastAccepted(b); err != nil {
		return fmt.Errorf("set last accepted failed: %v", err)
	}
	log.Info("Accept Block ProposeMintFeeTx")
	b.bs.ProposeMintFeeTx(b.Height(), b.ID(), b.Gas())
	return nil
}

func (b *Block) mintFee() error {
	num := len(b.ld.Miners)
	if num == 0 {
		return nil
	}
	// 80%: 20% * 4
	mintFee := new(big.Int).Mul(b.GasRebate20(), big.NewInt(4))
	mintFee = mintFee.Div(mintFee, big.NewInt(int64(num)))
	if mintFee.Sign() <= 0 {
		return nil
	}
	genesisAcc, err := b.State().LoadAccount(constants.GenesisAddr)
	if err != nil {
		return err
	}
	total := new(big.Int).Mul(mintFee, big.NewInt(int64(num)))
	if err := genesisAcc.Sub(total); err != nil {
		return fmt.Errorf("Block mintFee failed: %v", err)
	}
	for _, miner := range b.ld.Miners {
		acc, err := b.State().LoadAccount(miner)
		if err != nil {
			return err
		}
		if err := acc.Add(mintFee); err != nil {
			return err
		}
	}
	return nil
}

// Reject implements the snowman.Block choices.Decidable Reject interface
// This element will not be accepted by any correct node in the network.
func (b *Block) Reject() error {
	b.status = choices.Rejected
	b.bs.GivebackTxs(b.ld.Txs...)
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

func (b *Block) Gas() *big.Int {
	return new(big.Int).SetUint64(b.ld.Gas)
}

func (b *Block) GasPrice() *big.Int {
	return new(big.Int).SetUint64(b.ld.GasPrice)
}

func (b *Block) FeeCost() *big.Int {
	return new(big.Int).Mul(b.Gas(), b.GasPrice())
}

// Regard to pareto 80/20 Rule
// 20% to block builder, 80% to blocker miners,
// GasRebate20 = Gas * GasPrice * (GasRebateRate / 100) * 20%
func (b *Block) GasRebate20() *big.Int {
	gasRebate := new(big.Int).SetUint64(b.ld.GasRebateRate)
	gasRebate = gasRebate.Mul(gasRebate, b.FeeCost())
	return gasRebate.Div(gasRebate, big.NewInt(500))
}

func (b *Block) clear() {
	b.ld = nil
	b.bs = nil
	b.txs = nil
}
