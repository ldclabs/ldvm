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
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
)

var (
	_ snowman.Block = &Block{}

	emptyBlock = &Block{ld: &ld.Block{}}
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
	ld       *ld.Block
	parent   *ld.Block
	ctx      *Context
	bs       BlockState
	status   choices.Status
	txs      []Transaction
	miner    *Account
	verified bool
}

func NewGenesisBlock(b *ld.Block, ctx *Context) (*Block, error) {
	blk := &Block{ld: b, ctx: ctx}
	if err := b.SyntacticVerify(); err != nil {
		return nil, err
	}
	blk.status = choices.Processing
	blk.txs = make([]Transaction, len(b.Txs))
	for i := range b.Txs {
		tx, err := NewGenesisTx(b.Txs[i])
		if err != nil {
			return nil, err
		}
		blk.txs[i] = tx
	}
	return blk, nil
}

// tx gas, block gas, block miners should be set up !!!
func NewBlock(b *ld.Block, ctx *Context) (*Block, error) {
	blk := &Block{ld: b, ctx: ctx}
	if err := blk.init(); err != nil {
		return nil, err
	}
	return blk, nil
}

func (b *Block) init() error {
	if err := b.ld.SyntacticVerify(); err != nil {
		return err
	}
	b.status = choices.Processing
	b.txs = make([]Transaction, len(b.ld.Txs))
	for i := range b.ld.Txs {
		tx, err := NewTx(b.ld.Txs[i], false)
		if err != nil {
			return err
		}
		b.txs[i] = tx
	}
	return nil
}

func (b *Block) Unmarshal(data []byte) error {
	if b == nil {
		return fmt.Errorf("Block: Unmarshal on nil pointer")
	}
	if b.ld == nil {
		b.ld = &ld.Block{}
	}
	if err := b.ld.Unmarshal(data); err != nil {
		return err
	}
	return b.init()
}

func (b *Block) SetContext(ctx *Context) { b.ctx = ctx }

func (b *Block) Context() *Context { return b.ctx }

func (b *Block) MarshalJSON() ([]byte, error) {
	txs := make([]json.RawMessage, len(b.txs))
	for i := range b.txs {
		d, err := b.txs[i].MarshalJSON()
		if err != nil {
			return nil, err
		}
		txs[i] = d
	}
	b.ld.RawTxs = txs
	return b.ld.MarshalJSON()
}

func (b *Block) InitState(db database.Database, accepted bool) {
	b.bs = newBlockState(b.ctx, b.ID(), db)
	if accepted { // history block
		b.status = choices.Accepted
		for _, tx := range b.txs {
			tx.SetStatus(choices.Accepted)
		}
	}
}

func (b *Block) State() BlockState { return b.bs }

// ID implements the snowman.Block choices.Decidable ID interface
// ID returns a unique ID for this element.
func (b *Block) ID() ids.ID { return b.ld.ID() }

func (b *Block) Miner() (*Account, error) {
	var err error
	if b.miner == nil {
		feeCfg := b.FeeConfig()
		b.miner, err = b.bs.LoadAccount(b.ld.Miner)
		if err != nil {
			return nil, err
		}
		if !b.miner.ValidStake(feeCfg.MinValidatorStake) {
			logging.Log.Warn("Block.Miner %s not valid, used genesis account", b.ld.Miner)
			b.miner, err = b.bs.LoadAccount(constants.GenesisAccount)
		}
	}
	return b.miner, err
}

func (b *Block) VerifyGenesis() error {
	ts := b.Timestamp().Unix()
	for i := range b.ld.Txs {
		tx, ok := b.txs[i].(GenesisTx)
		if !ok {
			return fmt.Errorf("invalid genesis tx")
		}
		if err := tx.VerifyGenesis(b); err != nil {
			return err
		}
		if err := b.txs[i].Accept(b); err != nil {
			return err
		}
		b.ctx.StateDB().AddRecentTx(b.txs[i], choices.Processing)
		b.bs.AddEvent(b.txs[i].Event(ts))
	}

	if err := b.bs.SaveBlock(b); err != nil {
		return fmt.Errorf("save block failed: %v", err)
	}
	b.status = choices.Processing
	b.ctx.StateDB().AddVerifiedBlock(b)
	b.verified = true
	return nil
}

// Verify implements the snowman.Block Verify interface
// Verify that the state transition this block would make if accepted is valid.
// It is guaranteed that the Parent has been successfully verified.
func (b *Block) Verify() error {
	if b.bs == nil {
		b.InitState(b.ctx.StateDB().PreferredBlock().State().VersionDB(), false)
	}
	if err := b.verify(); err != nil {
		logging.Log.Warn("Block.Verify %s at %d error: %v", b.ID(), b.Height(), err)
		b.reject()
		return err
	}

	logging.Log.Info("Block.Verify %s at %d", b.ID(), b.Height())
	b.ctx.StateDB().AddVerifiedBlock(b)
	b.verified = true
	return nil
}

func (b *Block) verify() error {
	b.status = choices.Processing
	if b.ID() == ids.Empty {
		return fmt.Errorf("invalid block id")
	}

	var parent *ld.Block
	lastAccepted := b.ctx.StateDB().LastAcceptedBlock()
	preferred := b.ctx.StateDB().PreferredBlock()
	switch b.Parent() {
	case lastAccepted.ID():
		parent = lastAccepted
	case preferred.ID():
		parent = preferred.ld
	default:
		return fmt.Errorf("invalid block parent, expected %s, got %s",
			preferred.ID(), b.Parent())
	}

	if b.Height() != parent.Height+1 {
		return fmt.Errorf("invalid block height, expected %d, got %d",
			parent.Height+1, b.Height())
	}

	if b.ld.Timestamp < parent.Timestamp {
		return fmt.Errorf("invalid block timestamp, too early")
	}

	gas := uint64(0)
	txsSize := 0
	ts := b.Timestamp().Unix()
	for i := range b.ld.Txs {
		tx := b.txs[i]
		if err := tx.Verify(b); err != nil {
			return err
		}
		// d, _ := tx.MarshalJSON()
		// logging.Log.Info("Block.Accept %s %d, %v", b.ID(), i, string(d))
		if err := tx.Accept(b); err != nil {
			return err
		}
		b.ctx.StateDB().RemoveTx(tx.ID())
		gas += b.ld.Txs[i].Gas
		txsSize += len(b.ld.Txs[i].Bytes())

		b.ctx.StateDB().AddRecentTx(tx, choices.Processing)
		b.bs.AddEvent(tx.Event(ts))
	}

	if gas != b.ld.Gas {
		return fmt.Errorf("invalid block gas, expected %d, got %d", gas, b.ld.Gas)
	}
	if uint64(txsSize) > b.FeeConfig().MaxBlockTxsSize {
		return fmt.Errorf("invalid block txs size: should not more than %d bytes",
			b.FeeConfig().MaxBlockTxsSize)
	}

	// TODO verify miners
	if err := b.mintFee(); err != nil {
		return fmt.Errorf("set mint fee failed: %v", err)
	}
	if err := b.bs.SaveBlock(b); err != nil {
		return fmt.Errorf("save block failed: %v", err)
	}
	return nil
}

// Accept implements the snowman.Block choices.Decidable Accept interface
// This element will be accepted by every correct node in the network.
// Accept sets this block's status to Accepted and sets lastAccepted to this
// block's ID and saves this info to stateDB.
func (b *Block) Accept() error {
	logging.Log.Info("Block.Accept  %s at %d, parent: %s", b.ID(), b.Height(), b.Parent())
	if err := b.accept(); err != nil {
		logging.Log.Warn("Block.Accept %s at %d error: %v", b.ID(), b.Height(), err)
		b.reject()
		return err
	}

	if err := b.ctx.StateDB().SetLastAccepted(b); err != nil {
		b.reject()
		return fmt.Errorf("Block.Accept set last accepted failed: %v", err)
	}

	for i := range b.txs {
		b.ctx.StateDB().AddRecentTx(b.txs[i], choices.Accepted)
	}
	b.status = choices.Accepted
	return nil
}

func (b *Block) accept() error {
	if !b.verified {
		return fmt.Errorf("Block.Accept %s not verified", b.ID())
	}
	parent := b.ctx.StateDB().LastAcceptedBlock()
	if b.Parent() != parent.ID() {
		return fmt.Errorf("Block.Accept invalid parent, expected %s, got %s", parent.ID(), b.Parent())
	}

	return b.bs.Commit()
}

// Reject implements the snowman.Block choices.Decidable Reject interface
// This element will not be accepted by any correct node in the network.
func (b *Block) Reject() error {
	logging.Log.Info("Block.Reject %s at %d", b.ID(), b.Height())
	b.reject()
	return nil
}

func (b *Block) reject() {
	if b.status != choices.Rejected {
		b.status = choices.Rejected
		b.ctx.StateDB().AddTxs(false, b.ld.Txs...)
	}
}

func (b *Block) FeeConfig() *genesis.FeeConfig {
	return b.ctx.Chain().Fee(b.ld.Height)
}

func (b *Block) mintFee() error {
	miner, err := b.Miner()
	if err != nil {
		return err
	}
	ldc, err := b.bs.LoadAccount(constants.LDCAccount)
	if err != nil {
		return err
	}

	shares := make([]*Account, 0, len(b.ld.Shares))
	feeCfg := b.FeeConfig()
	for _, id := range b.ld.Shares {
		sc, err := b.bs.LoadAccount(id)
		if err != nil {
			return err
		}
		if sc.ValidStake(feeCfg.MinValidatorStake) {
			shares = append(shares, sc)
		} else {
			logging.Log.Warn("Block.mintFee stake account %s not valid, skipped", id)
		}
	}
	if len(shares) == 0 {
		shares = append(shares, miner)
	}

	num := big.NewInt(int64(len(shares)))
	// 80%: 20% * 4
	fee := new(big.Int).Mul(b.GasRebate20(), big.NewInt(4))
	fee = fee.Quo(fee, num)
	if fee.Sign() <= 0 {
		return nil
	}

	total := new(big.Int).Mul(fee, num)
	total = total.Add(total, b.GasRebate20())
	if err := ldc.Sub(constants.LDCAccount, total); err != nil {
		return fmt.Errorf("Block.mintFee failed: %v", err)
	}

	for _, share := range shares {
		if err = share.Add(constants.LDCAccount, fee); err != nil {
			return err
		}
	}
	return miner.Add(constants.LDCAccount, b.GasRebate20())
}

// Status implements the snowman.Block choices.Decidable Status interface
// Status returns this element's current status.
// If Accept has been called on an element with this ID, Accepted should be
// returned. Similarly, if Reject has been called on an element with this
// ID, Rejected should be returned. If the contents of this element are
// unknown, then Unknown should be returned. Otherwise, Processing should be
// returned.
func (b *Block) Status() choices.Status { return b.status }

func (b *Block) SetStatus(s choices.Status) { b.status = s }

// Parent implements the snowman.Block Parent interface
// Parent returns the ID of this block's parent.
func (b *Block) Parent() ids.ID { return b.ld.Parent }

func (b *Block) ParentBlock() (*ld.Block, error) {
	if b.parent == nil {
		parent, err := b.ctx.StateDB().GetBlock(b.ld.Parent)
		if err != nil {
			return nil, fmt.Errorf("Block.ParentBlock error: %v", err)
		}
		b.parent = parent.ld
	}
	return b.parent, nil
}

func (b *Block) AncestorBlocks(ancestorHeight uint64) ([]*Block, error) {
	if ancestorHeight >= b.ld.Height {
		return nil, fmt.Errorf("Block.AncestorBlocks invalid height, should < %d", b.ld.Height)
	}
	blks := make([]*Block, 0, b.ld.Height-ancestorHeight)
	var err error
	blk := b
	for blk.ld.Height > ancestorHeight {
		blk, err = b.ctx.StateDB().GetBlock(blk.ld.Parent)
		if err != nil {
			return nil, fmt.Errorf("Block.AncestorBlocks GetBlock error: %v", err)
		}
		blks = append(blks, blk)
	}
	return blks, nil
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

func (b *Block) Gas() *big.Int {
	return new(big.Int).SetUint64(b.ld.Gas)
}

func (b *Block) GasPrice() *big.Int {
	return new(big.Int).SetUint64(b.ld.GasPrice)
}

// Regard to pareto 80/20 Rule
// 20% to block builder, 80% to blocker shares,
// GasRebate20 = Gas * GasPrice * (GasRebateRate / 100) * 20%
func (b *Block) GasRebate20() *big.Int {
	gasRebate := new(big.Int).SetUint64(b.ld.GasRebateRate)
	gasRebate = gasRebate.Mul(gasRebate, b.ld.FeeCost())
	return gasRebate.Quo(gasRebate, big.NewInt(500))
}
