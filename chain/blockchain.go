// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"sync"
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
	lsync "github.com/ldclabs/ldvm/util/sync"
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
	DB() database.Database

	// global state
	HealthCheck() (interface{}, error)
	Bootstrap() error
	State() snow.State
	SetState(snow.State) error
	TotalSupply() *big.Int

	// blocks state
	BuildBlock() (*Block, error)
	ParseBlock([]byte) (*Block, error)
	GetBlockIDAtHeight(uint64) (ids.ID32, error)
	GetBlockAtHeight(uint64) (*Block, error)
	GetBlock(ids.ID32) (*Block, error)
	LastAcceptedBlock() *Block
	SetLastAccepted(*Block) error
	PreferredBlock() *Block
	SetPreference(ids.ID32) error
	AddVerifiedBlock(*Block)
	GetVerifiedBlock(ids.ID32) *Block

	// txs
	GetGenesisTxs() ld.Txs
	PreVerifyPdsTxs(...*ld.Transaction) (uint64, error)
	LoadTxsByIDsFromPds(uint64, []ids.ID32) (ld.Txs, error)

	LoadAccount(ids.Address) (*ld.Account, error)
	LoadModel(ids.ModelID) (*ld.ModelInfo, error)
	LoadData(ids.DataID) (*ld.DataInfo, error)
	LoadPrevData(ids.DataID, uint64) (*ld.DataInfo, error)
	LoadRawData(rawType string, key []byte) ([]byte, error)
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

	preferred         lsync.Value[*Block]
	lastAcceptedBlock lsync.Value[*Block]
	state             lsync.Value[snow.State]

	verifiedBlocks *sync.Map
	recentBlocks   *db.Cacher
	recentHeights  *db.Cacher
	recentData     *db.Cacher

	bb         *BlockBuilder
	txPool     *TxPool // Proposed transactions that haven't been put into a block yet
	rpcTimeout time.Duration
	builder    ids.StakeSymbol
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
		preferred:         lsync.Value[*Block]{},
		lastAcceptedBlock: lsync.Value[*Block]{},
		state:             lsync.Value[snow.State]{},
		verifiedBlocks:    new(sync.Map),
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

	txPool := NewTxPool(s.ctx, cfg.PdsEndpoint, transport)
	s.bb = NewBlockBuilder(txPool, toEngine)
	s.txPool = txPool
	s.preferred.Store(emptyBlock)
	s.lastAcceptedBlock.Store(emptyBlock)
	s.state.Store(0)
	s.builder = ids.Address(ctx.NodeID).ToStakeSymbol()

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

func (bc *blockChain) DB() database.Database {
	return bc.db
}

func (bc *blockChain) Context() *Context {
	return bc.ctx
}

func (bc *blockChain) Bootstrap() error {
	errp := erring.ErrPrefix("chain.BlockChain.Bootstrap: ")
	txs, err := bc.genesis.ToTxs()
	if err != nil {
		logging.Log.Error("BlockChain.Bootstrap", zap.Error(err))
		return errp.ErrorIf(err)
	}

	genesisBlock, err := NewGenesisBlock(bc.ctx, txs)
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
	genesisID, err := bc.GetBlockIDAtHeight(0)
	if err != nil {
		return errp.Errorf("load genesis id: %v", err)
	}
	// not the one on blockchain, means that the genesis data changed
	if genesisID != genesisBlock.Hash() {
		return errp.Errorf("invalid genesis data, expected genesis id %s", genesisID)
	}

	bc.genesisTxs = txs
	// genesis block is the last accepted block.
	if lastAcceptedID == genesisBlock.ID() {
		logging.Log.Info("BlockChain.Bootstrap finished", zap.Stringer("id", lastAcceptedID))
		genesisBlock.InitState(genesisBlock, bc.db)
		bc.preferred.Store(genesisBlock)
		bc.lastAcceptedBlock.Store(genesisBlock)
		return nil
	}

	// load the last accepted block
	lastAcceptedBlock, err := bc.GetBlock(ids.ID32(lastAcceptedID))
	if err != nil {
		return errp.Errorf("load last accepted block: %v", err)
	}

	parent, err := bc.GetBlock(ids.ID32(lastAcceptedBlock.Parent()))
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
	di, err = bc.LoadData(feeConfigID)
	if err != nil {
		return errp.Errorf("load last fee config: %v", err)
	}
	cfg, err := bc.genesis.Chain.AppendFeeConfig(di.Payload)
	if err != nil {
		return errp.Errorf("unmarshal fee config: %v", err)
	}

	for di.Version > 1 && cfg.StartHeight >= lastAcceptedBlock.ld.Height {
		di, err = bc.LoadPrevData(feeConfigID, di.Version-1)
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

func (bc *blockChain) HealthCheck() (interface{}, error) {
	errp := erring.ErrPrefix("chain.BlockChain.HealthCheck: ")
	id, err := database.GetID(bc.lastAcceptedDB, lastAcceptedKey)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return map[string]string{
		"LastAccepted": id.String(),
	}, nil
}

func (bc *blockChain) TotalSupply() *big.Int {
	t := new(big.Int)
	if acc, err := bc.LoadAccount(ids.LDCAccount); err == nil {
		t.Sub(acc.MaxTotalSupply, acc.Balance)
	}
	return t
}

func (bc *blockChain) SetState(state snow.State) error {
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
func (bc *blockChain) LastAcceptedBlock() *Block {
	return bc.lastAcceptedBlock.MustLoad()
}

func (bc *blockChain) AddVerifiedBlock(blk *Block) {
	bc.verifiedBlocks.Store(blk.ID(), blk)
}

func (bc *blockChain) GetVerifiedBlock(id ids.ID32) *Block {
	if v, ok := bc.verifiedBlocks.Load(id); ok {
		return v.(*Block)
	}
	return nil
}

func (bc *blockChain) SetLastAccepted(blk *Block) error {
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
			if err := bc.setPreference(preferred, blk); err != nil {
				return errp.ErrorIf(err)
			}

		default:
			canonical, err := preferred.State().GetBlockIDAtHeight(height)
			if err != nil {
				return errp.ErrorIf(err)
			}
			if avaids.ID(canonical) != id {
				if err := bc.setPreference(preferred, blk); err != nil {
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

	if blk.LD().Builder == bc.builder {
		go bc.txPool.UpdateBuildStatus(blk.Height(), &TxsBuildStatus{
			Accepted: blk.LD().Txs,
		})
	}
	go func() {
		bc.verifiedBlocks.Range(func(key, value any) bool {
			if b, ok := value.(*Block); ok {
				if b.Height() < height {
					b.Free()
					bc.verifiedBlocks.Delete(key)
				}
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
func (bc *blockChain) SetPreference(id ids.ID32) error {
	errp := erring.ErrPrefix("chain.BlockChain.SetPreference: ")
	preferred := bc.preferred.MustLoad()
	if preferred.ID() == avaids.ID(id) {
		return nil
	}

	blk := bc.GetVerifiedBlock(id)
	if blk == nil {
		return errp.Errorf("block %s not verified", id)
	}

	return errp.ErrorIf(bc.setPreference(preferred, blk))
}

func (bc *blockChain) setPreference(preferred, blk *Block) error {
	if blk.Parent() != preferred.ID() {
		if err := bc.reorg(preferred, blk); err != nil {
			return err
		}
	}

	bc.preferred.Store(blk)
	bc.bb.HandlePreferenceBlock(blk.Height())
	logging.Log.Info("BlockChain.SetPreference", zap.Stringer("id", blk.ID()))
	return nil
}

// reorg takes two blocks, an old chain and a new chain and will reconstruct the blocks.
func (bc *blockChain) reorg(oldBlock, newBlock *Block) error {
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

func (bc *blockChain) BuildBlock() (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.BuildBlock: ")
	if s := bc.State(); s != snow.NormalOp {
		return nil, errp.Errorf("state not bootstrapped, expected %q, got %q", snow.NormalOp, s)
	}
	if err := bc.preferred.MustLoad().FeeConfig().ValidBuilder(bc.builder); err != nil {
		return nil, errp.ErrorIf(err)
	}

	blk, err := bc.bb.Build(bc.ctx)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	id := blk.ID()
	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) ParseBlock(data []byte) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.ParseBlock: ")
	id := ids.ID32FromData(data)
	blk, err := bc.GetBlock(id)
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

	parent, err := bc.GetBlock(ids.ID32(blk.Parent()))
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	blk.SetContext(bc.ctx)
	blk.InitState(parent, parent.State().VersionDB())

	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) GetBlock(id ids.ID32) (*Block, error) {
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
			id, err := bc.GetBlockIDAtHeight(blk.Height())
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

func (bc *blockChain) GetBlockIDAtHeight(height uint64) (ids.ID32, error) {
	errp := erring.ErrPrefix("chain.BlockChain.GetBlockIDAtHeight: ")
	obj, err := bc.heightDB.LoadObject(database.PackUInt64(height), bc.recentHeights)
	if err != nil {
		return ids.ID32{}, errp.ErrorIf(err)
	}

	data := obj.(*db.RawObject)
	return ids.ID32FromBytes(*data)
}

func (bc *blockChain) GetBlockAtHeight(height uint64) (*Block, error) {
	errp := erring.ErrPrefix("chain.BlockChain.GetBlockAtHeight: ")
	id, err := bc.GetBlockIDAtHeight(height)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return bc.GetBlock(id)
}

func (bc *blockChain) GetGenesisTxs() ld.Txs {
	return bc.genesisTxs
}

// PreVerifyPdsTxs pre-verify transactions from PDS server
func (bc *blockChain) PreVerifyPdsTxs(txs ...*ld.Transaction) (uint64, error) {
	errp := erring.ErrPrefix("chain.BlockChain.PreVerifyPdsTxs: ")
	if len(txs) == 0 {
		return 0, errp.Errorf("no tx")
	}

	for _, tx := range txs {
		if err := tx.SyntacticVerify(); err != nil {
			return 0, errp.ErrorIf(err)
		}
	}

	if len(txs) == 1 && txs[0].Tx.Type == ld.TypeTest {
		return 0, errp.Errorf("TestTx should be in a batch transactions")
	}

	blk := bc.preferred.MustLoad()
	if err := blk.FeeConfig().ValidBuilder(bc.builder); err != nil {
		return 0, errp.ErrorIf(err)
	}

	if err := blk.TryBuildTxs(txs...); err != nil {
		return 0, errp.ErrorIf(err)
	}
	bc.bb.SignalTxsReady()
	return blk.Height() + 1, nil
}

func (bc *blockChain) LoadTxsByIDsFromPds(height uint64, txIDs []ids.ID32) (ld.Txs, error) {
	return bc.txPool.LoadByIDs(height, txIDs)
}

func (bc *blockChain) LoadAccount(id ids.Address) (*ld.Account, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadAccount: ")
	blk := bc.LastAcceptedBlock()
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

func (bc *blockChain) LoadModel(id ids.ModelID) (*ld.ModelInfo, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadModel: ")
	blk := bc.LastAcceptedBlock()
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

func (bc *blockChain) LoadData(id ids.DataID) (*ld.DataInfo, error) {
	errp := erring.ErrPrefix("chain.BlockChain.LoadData: ")
	blk := bc.LastAcceptedBlock()
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

func (bc *blockChain) LoadPrevData(id ids.DataID, version uint64) (*ld.DataInfo, error) {
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

func (bc *blockChain) LoadRawData(rawType string, key []byte) ([]byte, error) {
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
