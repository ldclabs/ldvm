// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"

	"github.com/ldclabs/ldvm/chain/transaction"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

var (
	_ snowman.Block            = &Block{}
	_ transaction.BlockContext = &Block{}

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
	ld        *ld.Block
	ctx       *Context
	parent    *Block // the genesis block is the parent of itself
	bs        BlockState
	status    choices.Status
	txs       []transaction.Transaction // txs field will unfold batch tx
	originTxs []*ld.Transaction         // originTxs keep the original txs
	txIDs     []ids.ID
	verified  bool
}

func NewGenesisBlock(ctx *Context, txs ld.Txs) (*Block, error) {
	blk := NewBlock(&ld.Block{
		NextGasPrice:  ctx.ChainConfig().FeeConfig.MinGasPrice,
		GasRebateRate: ctx.ChainConfig().FeeConfig.GasRebateRate,
		Validators:    make([]util.StakeSymbol, 0),
		Txs:           make([]*ld.Transaction, 0, len(txs)),
	}, ctx)
	blk.InitState(blk, ctx.Chain().DB(), false)
	blk.status = choices.Processing
	blk.txs = make([]transaction.Transaction, len(blk.ld.Txs))
	for i := range blk.ld.Txs {
		tx := blk.ld.Txs[i]
		ntx, err := transaction.NewGenesisTx(tx)
		if err != nil {
			return nil, err
		}
		blk.txs[i] = ntx
	}
	if err := blk.VerifyGenesis(); err != nil {
		return nil, err
	}
	return blk, nil
}

func NewBlock(b *ld.Block, ctx *Context) *Block {
	return &Block{ld: b, ctx: ctx}
}

// tx gas, block gas, block miners should be set up !!!
func (b *Block) SyntacticVerify() error {
	if err := b.ld.SyntacticVerify(); err != nil {
		return err
	}

	b.txIDs = make([]ids.ID, len(b.ld.Txs))
	for i, tx := range b.ld.Txs {
		b.txIDs[i] = tx.ID
	}

	b.status = choices.Processing
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
	if err := b.SyntacticVerify(); err != nil {
		return err
	}
	b.txs = make([]transaction.Transaction, len(b.ld.Txs))
	for i := range b.ld.Txs {
		tx := b.ld.Txs[i]
		tx.Height = b.ld.Height
		tx.Timestamp = b.ld.Timestamp
		ntx, err := transaction.NewTx(tx, false)
		if err != nil {
			return err
		}
		b.txs[i] = ntx
	}
	return nil
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
	return json.Marshal(b.ld)
}

func (b *Block) InitState(parent *Block, db database.Database, accepted bool) {
	b.parent = parent
	b.bs = newBlockState(
		b.ctx, b.ld.Height, b.ld.Timestamp, b.ld.ParentState, db)
	if accepted { // history block
		b.status = choices.Accepted
	}
}

func (b *Block) State() BlockState { return b.bs }

// ID implements the snowman.Block choices.Decidable ID interface
// ID returns a unique ID for this element.
func (b *Block) ID() ids.ID { return b.ld.ID }

func (b *Block) TxIDs() []ids.ID { return b.txIDs }

func (b *Block) Miner() util.StakeSymbol { return b.ld.Miner }

func (b *Block) BuildMiner(vbs BlockState) {
	b.ld.Miner, _ = vbs.LoadStakeAccountByNodeID(b.ctx.NodeID)
}

func (b *Block) BuildTxs(vbs BlockState, txs ...*ld.Transaction) choices.Status {
	status, _ := b.tryBuildTxs(vbs, true, txs...)
	return status
}

func (b *Block) TryBuildTxs(txs ...*ld.Transaction) error {
	vbs, err := b.bs.DeriveState()
	if err == nil {
		_, err = b.tryBuildTxs(vbs, false, txs...)
	}
	return err
}

func (b *Block) tryBuildTxs(vbs BlockState, add bool, txs ...*ld.Transaction) (choices.Status, error) {
	feeCfg := b.FeeConfig()
	gas := b.ld.Gas
	ntxs := make([]transaction.Transaction, 0, len(txs))

	for i := range txs {
		tx := txs[i]
		tx.Height = b.ld.Height
		tx.Timestamp = b.ld.Timestamp
		if tx.GasFeeCap < b.parent.ld.NextGasPrice {
			tx.Err = fmt.Errorf("Block.TryBuildTxs error: invalid gasFeeCap, expected >= %d, got %d",
				b.parent.ld.NextGasPrice, tx.GasFeeCap)
			return choices.Unknown, tx.Err
		}
		if tx.Gas() > feeCfg.MaxTxGas {
			tx.Err = fmt.Errorf("Block.TryBuildTxs error: gas too large, expected <= %d, got %d",
				feeCfg.MaxTxGas, tx.Gas())
			return choices.Rejected, tx.Err
		}
		gas += tx.Gas()

		// syntacticVerify again after gas calculation
		ntx, err := transaction.NewTx(tx, true)
		if err != nil {
			tx.Err = err
			return choices.Rejected, tx.Err
		}
		if err := ntx.Apply(b, vbs); err != nil {
			tx.Err = err
			return choices.Rejected, tx.Err
		}
	}

	if len(ntxs) == 0 {
		return choices.Rejected, fmt.Errorf("Block.TryBuildTxs error: no valid transaction")
	}

	if add {
		b.ld.Gas = gas
		for i := range ntxs {
			b.ld.Txs = append(b.ld.Txs, ntxs[i].LD())
			b.txs = append(b.txs, ntxs[i])
		}
	}
	return choices.Processing, nil
}

func (b *Block) BuildMinerFee(vbs BlockState) error {
	ldc, err := vbs.LoadAccount(constants.LDCAccount)
	if err != nil {
		return err
	}
	genesisAccount, err := vbs.LoadAccount(constants.GenesisAccount)
	if err != nil {
		return err
	}
	miner := genesisAccount
	if b.ld.Miner != util.StakeEmpty {
		miner, err = vbs.LoadAccount(util.EthID(b.ld.Miner))
		if err != nil {
			return err
		}
	}

	shares := make([]*transaction.Account, 0)
	if b.ctx.ValidatorState != nil {
		var err error
		b.ld.PCHeight, err = b.ctx.ValidatorState.GetCurrentHeight()
		if err != nil {
			return fmt.Errorf("Block.BuildFee ValidatorState.GetCurrentHeight error: %v", err)
		}
		var vs map[ids.NodeID]uint64
		vs, err = b.ctx.ValidatorState.GetValidatorSet(b.ld.PCHeight, b.ctx.SubnetID)
		if err != nil {
			return fmt.Errorf("Block.BuildFee ValidatorState.GetValidatorSet error: %v", err)
		}

		vv := make([]ids.NodeID, 0, len(vs))
		for nid := range vs {
			vv = append(vv, nid)
		}
		sort.SliceStable(vv, func(i, j int) bool {
			if vs[vv[i]] == vs[vv[j]] {
				return bytes.Compare(vv[i][:], vv[j][:]) == 1
			}
			return vs[vv[i]] > vs[vv[j]]
		})

		shares = make([]*transaction.Account, 0, len(vs))
		for _, nid := range vv {
			if _, acc := vbs.LoadStakeAccountByNodeID(nid); acc != nil {
				shares = append(shares, acc)
			}
		}
		sort.SliceStable(shares, func(i, j int) bool {
			return shares[i].Balance().Cmp(shares[j].Balance()) > 0
		})
		if len(shares) > 256 {
			shares = shares[:256]
		}
		b.ld.Validators = make([]util.StakeSymbol, 0, len(shares))
		for _, acc := range shares {
			b.ld.Validators = append(b.ld.Validators, util.EthID(acc.ID()).ToStakeSymbol())
		}
	}

	if len(shares) == 0 {
		shares = append(shares, genesisAccount)
		b.ld.Validators = []util.StakeSymbol{}
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
	if err := ldc.Sub(constants.NativeToken, total); err != nil {
		return fmt.Errorf("Block.BuildFee error: %v", err)
	}

	for _, share := range shares {
		if err = share.Add(constants.NativeToken, fee); err != nil {
			return err
		}
	}
	return miner.Add(constants.NativeToken, b.GasRebate20())
}

func (b *Block) VerifyMinerFee() error {
	return nil
}

func (b *Block) BuildState(vbs BlockState) error {
	// set the state's hash to ld.Block.State
	// set the block's hash to ld.Block.ID
	return vbs.SaveBlock(b.ld)
}

func (b *Block) VerifyGenesis() error {
	for i := range b.ld.Txs {
		tx, ok := b.txs[i].(transaction.GenesisTx)
		if !ok {
			return fmt.Errorf("invalid genesis tx")
		}
		if err := tx.ApplyGenesis(b, b.bs); err != nil {
			return err
		}
	}

	if err := b.bs.SaveBlock(b.ld); err != nil {
		return fmt.Errorf("save block error: %v", err)
	}
	b.status = choices.Processing
	b.ctx.Chain().AddVerifiedBlock(b)
	b.verified = true
	b.ctx.Chain().SetTxsStatus(choices.Processing, b.txIDs...)
	return nil
}

// Verify implements the snowman.Block Verify interface
// Verify that the state transition this block would make if accepted is valid.
// It is guaranteed that the Parent has been successfully verified.
func (b *Block) Verify() error {
	if err := b.verify(); err != nil {
		logging.Log.Warn("Block.Verify %s at %d error: %v", b.ID(), b.Height(), err)
		b.reject()
		return err
	}

	b.ctx.Chain().AddVerifiedBlock(b)
	b.verified = true
	b.ctx.Chain().SetTxsStatus(choices.Processing, b.txIDs...)
	logging.Log.Info("Block.Verify %s at %d", b.ID(), b.Height())
	return nil
}

func (b *Block) TxsSize() int {
	txsSize := 0
	for _, tx := range b.ld.Txs {
		txsSize += tx.BytesSize()
	}
	return txsSize
}

func (b *Block) verify() error {
	b.status = choices.Processing
	id := b.ID()
	if id == ids.Empty {
		return fmt.Errorf("invalid block id")
	}

	if h := b.parent.Height() + 1; b.Height() != h {
		return fmt.Errorf("invalid block height, expected %d, got %d",
			h, b.Height())
	}

	if b.ld.Timestamp < b.parent.ld.Timestamp {
		return fmt.Errorf("invalid block timestamp, too early")
	}

	if nextGasPrice := b.NextGasPrice(); b.ld.NextGasPrice != nextGasPrice {
		return fmt.Errorf("invalid block nextGasPrice, expected %d, got %d",
			nextGasPrice, b.ld.NextGasPrice)
	}

	gas := uint64(0)
	txsSize := 0
	for i := range b.ld.Txs {
		tx := b.txs[i]
		if err := tx.Apply(b, b.bs); err != nil {
			return err
		}
		gas += b.ld.Txs[i].Gas()
		txsSize += len(b.ld.Txs[i].Bytes())
	}

	if gas != b.ld.Gas {
		return fmt.Errorf("invalid block gas, expected %d, got %d", gas, b.ld.Gas)
	}
	if uint64(txsSize) > b.FeeConfig().MaxBlockTxsSize {
		return fmt.Errorf("invalid block txs size: should not more than %d bytes",
			b.FeeConfig().MaxBlockTxsSize)
	}

	// TODO verify miners
	if err := b.VerifyMinerFee(); err != nil {
		return fmt.Errorf("set mint fee error: %v", err)
	}
	if err := b.bs.SaveBlock(b.ld); err != nil {
		return fmt.Errorf("save block error: %v", err)
	}
	if id != b.ld.ID {
		return fmt.Errorf("invalid block id, expected %s, got %s", id, b.ld.ID)
	}
	return nil
}

// Accept implements the snowman.Block choices.Decidable Accept interface
// This element will be accepted by every correct node in the network.
// Accept sets this block's status to Accepted and sets lastAccepted to this
// block's ID and saves this info to stateDB.
func (b *Block) Accept() error {
	logging.Log.Info("Block.Accept  %s at %d, parent: %s", b.ID(), b.Height(), b.Parent())
	if !b.verified {
		b.reject()
		return fmt.Errorf("Block.Accept %s not verified", b.ID())
	}

	if err := b.bs.Commit(); err != nil {
		logging.Log.Warn("Block.Accept %s at %d error: %v", b.ID(), b.Height(), err)
		b.reject()
		return err
	}

	if err := b.ctx.Chain().SetLastAccepted(b); err != nil {
		b.reject()
		return fmt.Errorf("Block.Accept set last accepted error: %v", err)
	}

	b.status = choices.Accepted
	b.ctx.Chain().SetTxsStatus(choices.Accepted, b.txIDs...)
	return nil
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
		b.ctx.Chain().AddLocalTxs(b.originTxs...)
	}
}

func (b *Block) ChainConfig() *genesis.ChainConfig {
	return b.ctx.ChainConfig()
}

func (b *Block) FeeConfig() *genesis.FeeConfig {
	return b.ctx.ChainConfig().Fee(b.ld.Height)
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

func (b *Block) AncestorBlocks(ancestorHeight uint64) ([]*Block, error) {
	if ancestorHeight >= b.ld.Height {
		return nil, fmt.Errorf("Block.AncestorBlocks invalid height, should < %d", b.ld.Height)
	}

	blks := make([]*Block, b.ld.Height-ancestorHeight)
	blk := b
	for blk.ld.Height > ancestorHeight {
		pid := blk.ld.Parent
		blk = b.ctx.Chain().GetVerifiedBlock(pid)
		if blk == nil {
			return nil, fmt.Errorf("Block.AncestorBlocks: %s not found", pid)
		}
		blks[blk.ld.Height-ancestorHeight] = blk
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
	return new(big.Int).SetUint64(b.parent.ld.NextGasPrice)
}

func (b *Block) NextGasPrice() uint64 {
	feeCfg := b.FeeConfig()
	nextGasPrice := b.parent.ld.NextGasPrice
	txsSize := b.ld.Txs.BytesSize()
	if uint64(txsSize)*2 <= feeCfg.MaxBlockTxsSize {
		nextGasPrice = uint64(float64(nextGasPrice) / math.SqrtPhi)
		if nextGasPrice < feeCfg.MinGasPrice {
			nextGasPrice = feeCfg.MinGasPrice
		}
	} else if float64(txsSize)*math.SqrtPhi >= float64(feeCfg.MaxBlockTxsSize) {
		nextGasPrice = uint64(float64(nextGasPrice) * math.SqrtPhi)
		if nextGasPrice > feeCfg.MaxGasPrice {
			nextGasPrice = feeCfg.MaxGasPrice
		}
	}
	return nextGasPrice
}

// Regard to pareto 80/20 Rule
// 20% to block builder, 80% to blocker shares,
// GasRebate20 = Gas * GasPrice * (GasRebateRate / 100) * 20%
func (b *Block) GasRebate20() *big.Int {
	gasRebate := new(big.Int).SetUint64(b.ld.GasRebateRate)
	gasRebate = gasRebate.Mul(gasRebate, new(big.Int).SetUint64(b.ld.Gas))
	gasRebate = gasRebate.Mul(gasRebate, b.GasPrice())
	return gasRebate.Quo(gasRebate, big.NewInt(500))
}
