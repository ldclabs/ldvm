// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ava-labs/avalanchego/database"
	avaids "github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/sync"
)

var _ BlockChain = &blockChain{}

var (
	dbPrefix             = []byte("LDVM")
	lastAcceptedDBPrefix = []byte{'K'}
	heightDBPrefix       = []byte{'H'}
	blockDBPrefix        = []byte{'B'}
	accountDBPrefix      = []byte{'A'}
	ledgerDBPrefix       = []byte{'L'}
	modelDBPrefix        = []byte{'M'}
	dataDBPrefix         = []byte{'D'}
	prevDataDBPrefix     = []byte{'P'}
	stateDBPrefix        = []byte{'S'}
	nameDBPrefix         = []byte{'N'} // inverted index

	lastAcceptedKey = []byte("last_accepted_key")
)

// BlockChain defines methods to manage state with Blocks and LastAcceptedIDs.
type BlockChain interface {
	// global context
	Context() *Context
	Info() map[string]any
	DB() database.Database

	// global state
	HealthCheck(context.Context) (any, error)
	Bootstrap(context.Context) error
	State() snow.State
	SetState(context.Context, snow.State) error
	TotalSupply(context.Context) *big.Int

	// blocks state
	IsBuilder() bool
	BuildBlock(context.Context) (*Block, error)
	ParseBlock(context.Context, []byte) (*Block, error)
	GetBlockIDAtHeight(context.Context, uint64) (ids.ID32, error)
	GetBlockAtHeight(context.Context, uint64) (*Block, error)
	GetBlock(context.Context, ids.ID32) (*Block, error)
	LastAcceptedBlock(context.Context) *Block
	SetLastAccepted(context.Context, *Block) error
	PreferredBlock() *Block
	SetPreference(context.Context, ids.ID32) error
	AddVerifiedBlock(*Block)
	GetVerifiedBlock(ids.ID32) *Block

	// txs
	GetGenesisTxs() ld.Txs
	PreVerifyPOSTxs(context.Context, ...*ld.Transaction) error
	LoadTxsByIDsFromPOS(context.Context, uint64, []ids.ID32) (ld.Txs, error)

	LoadAccount(context.Context, ids.Address) (*ld.Account, error)
	LoadModel(context.Context, ids.ModelID) (*ld.ModelInfo, error)
	LoadData(context.Context, ids.DataID) (*ld.DataInfo, error)
	LoadPrevData(context.Context, ids.DataID, uint64) (*ld.DataInfo, error)
	LoadRawData(context.Context, string, []byte) ([]byte, error)
}

type blockChain struct {
	ctx          *Context
	config       *config.Config
	genesis      *genesis.Genesis
	genesisBlock *Block
	genesisTxs   ld.Txs

	db             database.Database
	blockDB        *db.PrefixDB
	heightDB       *db.PrefixDB
	lastAcceptedDB *db.PrefixDB
	accountDB      *db.PrefixDB
	ledgerDB       *db.PrefixDB
	modelDB        *db.PrefixDB
	dataDB         *db.PrefixDB
	prevDataDB     *db.PrefixDB
	stateDB        *db.PrefixDB
	nameDB         *db.PrefixDB

	preferred         sync.Value[*Block]
	lastAcceptedBlock sync.Value[*Block]
	state             sync.Value[snow.State]

	verifiedBlocks sync.Map[ids.ID32, *Block]
	recentBlocks   *db.Cacher
	recentHeights  *db.Cacher
	recentData     *db.Cacher

	rpcTimeout time.Duration
	bb         *BlockBuilder
	txPool     *TxPool
}

func NewChain(
	name string,
	ctx *snow.Context,
	cfg *config.Config,
	gs *genesis.Genesis,
	baseDB database.Database,
	toEngine chan<- common.Message,
	transport http.RoundTripper,
) *blockChain {
	pdb := db.NewPrefixDB(baseDB, dbPrefix, 512)

	s := &blockChain{
		config:            cfg,
		genesis:           gs,
		db:                baseDB,
		rpcTimeout:        3 * time.Second,
		preferred:         sync.Value[*Block]{},
		lastAcceptedBlock: sync.Value[*Block]{},
		state:             sync.Value[snow.State]{},
		verifiedBlocks:    sync.Map[ids.ID32, *Block]{},
		blockDB:           pdb.With(blockDBPrefix),
		heightDB:          pdb.With(heightDBPrefix),
		lastAcceptedDB:    pdb.With(lastAcceptedDBPrefix),
		accountDB:         pdb.With(accountDBPrefix),
		ledgerDB:          pdb.With(ledgerDBPrefix),
		modelDB:           pdb.With(modelDBPrefix),
		dataDB:            pdb.With(dataDBPrefix),
		prevDataDB:        pdb.With(prevDataDBPrefix),
		stateDB:           pdb.With(stateDBPrefix),
		nameDB:            pdb.With(nameDBPrefix),
	}

	s.nameDB.SetHashKey(nameHashKey)
	s.ctx = NewContext(name, ctx, s, cfg, gs)

	s.preferred.Store(emptyBlock)
	s.lastAcceptedBlock.Store(emptyBlock)
	s.state.Store(0)

	txPool := NewTxPool(s.ctx, cfg.POSEndpoint, transport)
	s.bb = NewBlockBuilder(txPool, toEngine)
	s.txPool = txPool

	// this data will not change, so we can cache it
	s.recentBlocks = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(Block)
	})
	s.recentHeights = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(db.RawObject)
	})
	s.recentData = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(ld.DataInfo)
	})
	return s
}

func (bc *blockChain) Info() map[string]any {
	return map[string]any{
		"networkId": bc.ctx.NetworkID,
		"subnetId":  bc.ctx.SubnetID.String(),
		"nodeId":    bc.ctx.NodeID.String(),
		"builderId": bc.ctx.Builder(),
		"state":     bc.State().String(),
	}
}

func (bc *blockChain) DB() database.Database {
	return bc.db
}

func (bc *blockChain) Context() *Context {
	return bc.ctx
}

func (bc *blockChain) Bootstrap(ctx context.Context) error {
	var err error
	errp := erring.ErrPrefix("chain.BlockChain.Bootstrap: ")

	bc.genesisTxs, err = bc.genesis.ToTxs()
	if err != nil {
		logging.Log.Error("BlockChain.Bootstrap", zap.Error(err))
		return errp.ErrorIf(err)
	}

	genesisBlock, err := NewGenesisBlock(bc.ctx, bc.genesisTxs)
	if err != nil {
		logging.Log.Error("BlockChain.Bootstrap", zap.Error(err))
		return errp.ErrorIf(err)
	}

	if genesisBlock.Parent() != avaids.Empty ||
		genesisBlock.ID() == avaids.Empty ||
		genesisBlock.Height() != 0 ||
		genesisBlock.Timestamp2() != 0 {
		return errp.Errorf("invalid genesis block")
	}

	bc.genesisBlock = genesisBlock
	lastAcceptedID, err := database.GetID(bc.lastAcceptedDB, lastAcceptedKey)
	// create genesis block
	if err == database.ErrNotFound {
		logging.Log.Info("BlockChain.Bootstrap create genesis block",
			zap.Stringer("id", genesisBlock.ID()))
		bc.preferred.Store(genesisBlock)

		if err := genesisBlock.Accept(context.TODO()); err != nil {
			logging.Log.Error("BlockChain.Bootstrap", zap.Error(err))
			return errp.ErrorIf(err)
		}
		return nil
	}

	if err != nil {
		return errp.Errorf("load last_accepted: %v", err)
	}

	// verify genesis data
	genesisID, err := bc.GetBlockIDAtHeight(ctx, 0)
	if err != nil {
		return errp.Errorf("load genesis id: %v", err)
	}
	// not the one on blockchain, means that the genesis data changed
	if genesisID != genesisBlock.Hash() {
		return errp.Errorf("invalid genesis data, expected genesis id %s", genesisID)
	}

	// genesis block is the last accepted block.
	if lastAcceptedID == genesisBlock.ID() {
		logging.Log.Info("BlockChain.Bootstrap finished", zap.Stringer("id", lastAcceptedID))
		genesisBlock.InitState(genesisBlock, bc.db)
		bc.preferred.Store(genesisBlock)
		bc.lastAcceptedBlock.Store(genesisBlock)
		return nil
	}

	// load the last accepted block
	lastAcceptedBlock, err := bc.GetBlock(ctx, ids.ID32(lastAcceptedID))
	if err != nil {
		return errp.Errorf("load last accepted block: %v", err)
	}

	parent, err := bc.GetBlock(ctx, ids.ID32(lastAcceptedBlock.Parent()))
	if err != nil {
		return errp.Errorf("load last accepted block' parent: %v", err)
	}

	lastAcceptedBlock.InitState(parent, bc.db)
	lastAcceptedBlock.SetStatus(choices.Accepted)
	bc.preferred.Store(lastAcceptedBlock)
	bc.lastAcceptedBlock.Store(lastAcceptedBlock)

	// load latest fee config from chain.
	var di *ld.DataInfo
	feeConfigID := bc.genesis.Chain.FeeConfigID
	di, err = bc.LoadData(ctx, feeConfigID)
	if err != nil {
		return errp.Errorf("load last fee config: %v", err)
	}
	cfg, err := bc.genesis.Chain.AppendFeeConfig(di.Payload)
	if err != nil {
		return errp.Errorf("unmarshal fee config: %v", err)
	}

	for di.Version > 1 && cfg.StartHeight >= lastAcceptedBlock.ld.Height {
		di, err = bc.LoadPrevData(ctx, feeConfigID, di.Version-1)
		if err != nil {
			return errp.Errorf("load previous fee config: %v", err)
		}
		cfg, err = bc.genesis.Chain.AppendFeeConfig(di.Payload)
		if err != nil {
			return errp.Errorf("unmarshal fee config: %v", err)
		}
	}

	logging.Log.Info("BlockChain.Bootstrap finished",
		zap.Stringer("id", lastAcceptedBlock.ID()),
		zap.Uint64("height", lastAcceptedBlock.Height()),
		zap.Int("configs", len(bc.genesis.Chain.FeeConfigs)))
	return nil
}

func (bc *blockChain) HealthCheck(ctx context.Context) (any, error) {
	errp := erring.ErrPrefix("chain.BlockChain.HealthCheck: ")
	id, err := database.GetID(bc.lastAcceptedDB, lastAcceptedKey)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return map[string]string{
		"lastAccepted": id.String(),
	}, nil
}

func (bc *blockChain) TotalSupply(ctx context.Context) *big.Int {
	t := new(big.Int)
	if acc, err := bc.LoadAccount(ctx, ids.LDCAccount); err == nil {
		t.Sub(acc.MaxTotalSupply, acc.Balance)
	}
	return t
}

func (bc *blockChain) SetState(ctx context.Context, state snow.State) error {
	errp := erring.ErrPrefix("chain.BlockChain.SetState: ")
	switch state {
	case snow.Bootstrapping:
		bc.state.Store(state)
		return nil
	case snow.NormalOp:
		if bc.preferred.MustLoad() == emptyBlock {
			return errp.Errorf("bootstrap failed")
		}
		bc.state.Store(state)
		return nil
	default:
		return errp.ErrorIf(snow.ErrUnknownState)
	}
}

func (bc *blockChain) State() snow.State {
	return bc.state.MustLoad()
}

// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (bc *blockChain) LastAcceptedBlock(ctx context.Context) *Block {
	return bc.lastAcceptedBlock.MustLoad()
}

func (bc *blockChain) AddVerifiedBlock(blk *Block) {
	bc.verifiedBlocks.Store(blk.Hash(), blk)
}

func (bc *blockChain) GetVerifiedBlock(id ids.ID32) *Block {
	v, _ := bc.verifiedBlocks.Load(id)
	return v
}

func (bc *blockChain) SetLastAccepted(ctx context.Context, blk *Block) error {
	errp := erring.ErrPrefix("chain.BlockChain.SetLastAccepted: ")
	if parent := bc.lastAcceptedBlock.MustLoad(); parent.ID() != blk.Parent() {
		return errp.Errorf("invalid parent, expected %s:%d, got %s:%d",
			parent.ID(), parent.Height(), blk.Parent(), blk.Height())
	}

	id := blk.ID()
	height := blk.Height()
	preferred := bc.preferred.MustLoad()
	if preferred.ID() != id {
		logging.Log.Warn("BlockChain.SetLastAccepted accepting block in non-canonical chain",
			zap.Stringer("expected id", preferred.ID()),
			zap.Uint64("expected height", preferred.Height()),
			zap.Stringer("got id", id),
			zap.Uint64("got height", height))

		switch {
		case preferred.Height() <= height:
			if err := bc.setPreference(ctx, preferred, blk); err != nil {
				return errp.ErrorIf(err)
			}

		default:
			canonical, err := preferred.State().GetBlockIDAtHeight(height)
			if err != nil {
				return errp.ErrorIf(err)
			}
			if avaids.ID(canonical) != id {
				if err := bc.setPreference(ctx, preferred, blk); err != nil {
					return errp.ErrorIf(err)
				}
			}
		}
	}

	if err := database.PutID(bc.lastAcceptedDB, lastAcceptedKey, id); err != nil {
		return errp.ErrorIf(err)
	}

	bc.lastAcceptedBlock.Store(blk)
	bc.recentBlocks.SetObject(id[:], blk)

	if blk.LD().Builder == bc.ctx.Builder() {
		if err := bc.txPool.AcceptByBlock(ctx, blk.LD()); err != nil {
			return errp.ErrorIf(err)
		}
	}

	go func() {
		bc.verifiedBlocks.Range(func(key ids.ID32, b *Block) bool {
			if b.Height() < height {
				b.Free()
				bc.verifiedBlocks.Delete(key)
			}
			return true
		})
	}()
	return nil
}

func (bc *blockChain) PreferredBlock() *Block {
	return bc.preferred.MustLoad()
}

// SetPreference persists the VM of the currently preferred block into database.
// This should always be a block that has no children known to consensus.
func (bc *blockChain) SetPreference(ctx context.Context, id ids.ID32) error {
	errp := erring.ErrPrefix("chain.BlockChain.SetPreference: ")
	preferred := bc.preferred.MustLoad()
	if preferred.ID() == avaids.ID(id) {
		return nil
	}

	blk := bc.GetVerifiedBlock(id)
	if blk == nil {
		return errp.Errorf("block %s not verified", id)
	}

	return errp.ErrorIf(bc.setPreference(ctx, preferred, blk))
}

func (bc *blockChain) setPreference(ctx context.Context, preferred, blk *Block) error {
	if blk.Parent() != preferred.ID() {
		if err := bc.reorg(ctx, preferred, blk); err != nil {
			return err
		}
	}

	bc.preferred.Store(blk)
	bc.bb.HandlePreferenceBlock(ctx, blk.Height())
	logging.Log.Info("BlockChain.SetPreference", zap.Stringer("id", blk.ID()))
	return nil
}

// reorg takes two blocks, an old chain and a new chain and will reconstruct the blocks.
func (bc *blockChain) reorg(ctx context.Context, oldBlock, newBlock *Block) error {
	accepted := bc.lastAcceptedBlock.MustLoad()
	newChain, err := newBlock.AncestorBlocks(accepted.Height())
	if err != nil {
		return err
	}
	if newChain[0].ID() != accepted.ID() {
		return fmt.Errorf("reorg: new chain does not start with the last accepted block")
	}

	set := make(map[uint64]avaids.ID, len(newChain))
	for _, blk := range newChain {
		set[blk.Height()] = blk.ID()
	}

	// new chain is longer
	if set[oldBlock.Height()] == oldBlock.ID() {
		return nil
	}

	oldChain, err := oldBlock.AncestorBlocks(accepted.Height())
	if err != nil {
		return err
	}

	for len(oldChain) > 0 {
		blk := oldChain[0]
		oldChain = oldChain[1:]
		if set[blk.Height()] == blk.ID() {
			continue
		}
		blk.Reject(context.TODO())
	}
	oldBlock.Reject(context.TODO())
	return nil
}

func (bc *blockChain) IsBuilder() bool {
	return bc.preferred.MustLoad().FeeConfig().ValidBuilder(bc.ctx.Builder()) == nil
}

func (bc *blockChain) BuildBlock(ctx context.Context) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.BuildBlock: ")
	if s := bc.State(); s != snow.NormalOp {
		return nil, errp.Errorf("state not bootstrapped, expected %q, got %q", snow.NormalOp, s)
	}

	builder := bc.ctx.Builder()
	if err := bc.preferred.MustLoad().FeeConfig().ValidBuilder(builder); err != nil {
		return nil, errp.ErrorIf(err)
	}

	blk, err := bc.bb.Build(ctx, builder)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	id := blk.ID()
	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) ParseBlock(ctx context.Context, data []byte) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.ParseBlock: ")
	id := ids.ID32FromData(data)
	blk, err := bc.GetBlock(ctx, id)
	if err == nil {
		return blk, nil
	}

	blk = new(Block)
	if err := blk.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}

	if id != ids.ID32(blk.ID()) {
		return nil, errp.Errorf("blockChain.ParseBlock: invalid block id at %d, expected %s, got %s",
			blk.Height(), id, blk.ID())
	}

	parent, err := bc.GetBlock(ctx, ids.ID32(blk.Parent()))
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	blk.SetContext(bc.ctx)
	blk.InitState(parent, parent.State().VersionDB())

	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) GetBlock(ctx context.Context, id ids.ID32) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.GetBlock: ")
	if ids.ID32(bc.genesisBlock.ID()) == id {
		return bc.genesisBlock, nil
	}

	last := bc.lastAcceptedBlock.MustLoad()
	if ids.ID32(last.ID()) == id {
		return last, nil
	}

	if blk := bc.GetVerifiedBlock(id); blk != nil {
		return blk, nil
	}

	obj, err := bc.blockDB.LoadObject(id[:], bc.recentBlocks)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	blk := obj.(*Block)
	blk.SetContext(bc.ctx)

	if blk.Status() == choices.Unknown {
		if blk.Height() > last.Height() {
			blk.SetStatus(choices.Processing)
		} else {
			id, err := bc.GetBlockIDAtHeight(ctx, blk.Height())
			switch err {
			case nil:
				if id == ids.ID32(blk.ID()) {
					blk.SetStatus(choices.Accepted)
				} else {
					blk.SetStatus(choices.Rejected)
				}
			case database.ErrNotFound:
				blk.SetStatus(choices.Processing)
			default:
				return nil, errp.ErrorIf(err)
			}
		}
	}
	return blk, nil
}

func (bc *blockChain) GetBlockIDAtHeight(ctx context.Context, height uint64) (ids.ID32, error) {
	errp := erring.ErrPrefix("chain.BlockChain.GetBlockIDAtHeight: ")
	obj, err := bc.heightDB.LoadObject(database.PackUInt64(height), bc.recentHeights)
	if err != nil {
		return ids.ID32{}, errp.ErrorIf(err)
	}

	data := obj.(*db.RawObject)
	return ids.ID32FromBytes(*data)
}

func (bc *blockChain) GetBlockAtHeight(ctx context.Context, height uint64) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.GetBlockAtHeight: ")
	id, err := bc.GetBlockIDAtHeight(ctx, height)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return bc.GetBlock(ctx, id)
}

func (bc *blockChain) GetGenesisTxs() ld.Txs {
	return bc.genesisTxs
}

// PreVerifyPOSTxs pre-verify transactions from POS server
func (bc *blockChain) PreVerifyPOSTxs(ctx context.Context, txs ...*ld.Transaction) error {
	errp := erring.ErrPrefix("chain.BlockChain.PreVerifyPOSTxs: ")
	if len(txs) == 0 {
		return errp.Errorf("no tx")
	}

	for _, tx := range txs {
		if err := tx.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if len(txs) == 1 && txs[0].Tx.Type == ld.TypeTest {
		return errp.Errorf("TestTx should be in a batch transactions")
	}

	blk := bc.preferred.MustLoad()
	if err := blk.FeeConfig().ValidBuilder(bc.ctx.Builder()); err != nil {
		return errp.ErrorIf(err)
	}

	if err := blk.TryBuildTxs(txs...); err != nil {
		return errp.ErrorIf(err)
	}

	go bc.bb.SignalTxsReady()
	return nil
}

func (bc *blockChain) LoadTxsByIDsFromPOS(ctx context.Context, height uint64, txIDs []ids.ID32) (ld.Txs, error) {
	return bc.txPool.LoadByIDs(ctx, height, txIDs)
}

func (bc *blockChain) LoadAccount(ctx context.Context, id ids.Address) (*ld.Account, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadAccount: ")
	blk := bc.LastAcceptedBlock(ctx)
	acc, err := blk.State().LoadAccount(id)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	rt := &ld.Account{}
	if err = ld.Copy(rt, acc.LD()); err != nil {
		return nil, errp.ErrorIf(err)
	}
	rt.Height = blk.Height()
	rt.Timestamp = blk.Timestamp2()
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadModel(ctx context.Context, id ids.ModelID) (*ld.ModelInfo, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadModel: ")
	blk := bc.LastAcceptedBlock(ctx)
	mi, err := blk.State().LoadModel(id)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	rt := &ld.ModelInfo{}
	if err = ld.Copy(rt, mi); err != nil {
		return nil, errp.ErrorIf(err)
	}
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadData(ctx context.Context, id ids.DataID) (*ld.DataInfo, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadData: ")
	blk := bc.LastAcceptedBlock(ctx)
	di, err := blk.State().LoadData(id)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	rt := &ld.DataInfo{}
	if err = ld.Copy(rt, di); err != nil {
		return nil, errp.ErrorIf(err)
	}
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadPrevData(ctx context.Context, id ids.DataID, version uint64) (*ld.DataInfo, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadPrevData: ")
	if version == 0 {
		return nil, errp.Errorf("invalid version %d", version)
	}

	obj, err := bc.prevDataDB.LoadObject(id.VersionKey(version), bc.recentData)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	rt := obj.(*ld.DataInfo)
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadRawData(ctx context.Context, rawType string, key []byte) ([]byte, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadRawData: ")
	var pdb *db.PrefixDB
	switch rawType {
	case "block":
		pdb = bc.blockDB
	case "state":
		pdb = bc.stateDB
	case "account":
		pdb = bc.accountDB
	case "ledger":
		pdb = bc.ledgerDB
	case "model":
		pdb = bc.modelDB
	case "data":
		pdb = bc.dataDB
	case "prevdata":
		pdb = bc.prevDataDB
	case "name":
		pdb = bc.nameDB
	default:
		return nil, errp.Errorf("unknown type %q", rawType)
	}

	return errp.ErrorMap(pdb.Get(key))
}

func nameHashKey(key []byte) []byte {
	k := sha3.Sum224(key)
	return k[:]
}
