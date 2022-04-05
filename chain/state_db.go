// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
)

var _ StateDB = &stateDB{}

var (
	// Should never happen
	errPreferredBlock = fmt.Errorf("stateDB is not bootstrapped, no preferred block")
)

var (
	lastAcceptedKeyDBPrefix = []byte{'K'}
	heightDBPrefix          = []byte{'H'}
	blockDBPrefix           = []byte{'B'}
	accountDBPrefix         = []byte{'A'}
	modelDBPrefix           = []byte{'M'}
	dataDBPrefix            = []byte{'D'}
	nameDBPrefix            = []byte{'N'} // inverted index

	lastAcceptedKey = []byte("last_accepted_key")
)

// StateDB defines methods to manage state with Blocks and LastAcceptedIDs.
type StateDB interface {
	Bootstrap() error
	SetState(snow.State) error

	GetBlock(ids.ID) (*Block, error)
	ParseBlock([]byte) (*Block, error)
	LastAcceptedBlock() (*Block, error)
	SetLastAccepted(*Block) error
	PreferredBlock() *Block
	SetPreference(ids.ID) error
	GetBlockIDAtHeight(uint64) (ids.ID, error)

	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig
	AddTxs(...*ld.Transaction)
}

func NewState(
	ctx *snow.Context,
	db database.Database,
	gs *genesis.Genesis,
	cfg *config.Config,
) *stateDB {
	s := &stateDB{
		ctx:               ctx,
		db:                db,
		blockDB:           prefixdb.New(blockDBPrefix, db),
		heightDB:          prefixdb.New(heightDBPrefix, db),
		lastAcceptedKeyDB: prefixdb.New(lastAcceptedKey, db),
		genesis:           gs,
		config:            cfg,
	}
	return s
}

type stateDB struct {
	mu      sync.RWMutex
	ctx     *snow.Context
	genesis *genesis.Genesis
	config  *config.Config
	state   snow.State

	db                database.Database
	blockDB           database.Database
	heightDB          database.Database
	lastAcceptedKeyDB database.Database

	// Proposed transactions that haven't been put into a block yet
	txPool TxPool

	// prevents reorgs past this height,
	// should be preferred block or preferred block' child
	// accepted but may be not committed
	lastAcceptedBlock *Block
	// committed block, lastAcceptedKey's id is the preferred block
	preferred *Block

	// multiple accepted blocks may exist befoce SetPreference
	acceptedBlocks map[ids.ID]*Block
}

func (s *stateDB) Bootstrap() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	genesisLdBlock, err := s.genesis.ToBlock()
	if err != nil {
		return err
	}
	genesisBlock, err := NewBlock(genesisLdBlock)
	if err != nil {
		return err
	}
	if genesisBlock.Parent() != ids.Empty ||
		genesisBlock.ID() == ids.Empty ||
		genesisBlock.Height() != 0 ||
		genesisBlock.Timestamp().Unix() != 0 {
		return fmt.Errorf("invalid genesis block")
	}

	sb := newStateBlock(s, s.db)
	lastAcceptedID, err := database.GetID(s.lastAcceptedKeyDB, lastAcceptedKey)
	// create genesis block
	if err == database.ErrNotFound {
		genesisBlock.InitState(sb, nil)
		if err := genesisBlock.Verify(); err != nil {
			return fmt.Errorf("verify genesis block error: %v", err)
		}
		if err := genesisBlock.Accept(); err != nil {
			return fmt.Errorf("accept genesis block error: %v", err)
		}
		s.lastAcceptedBlock = genesisBlock // TODO
		return s.commitPreference(genesisBlock)
	}

	if err != nil {
		return fmt.Errorf("load last_accepted failed: %v", err)
	}

	// genesis block is the last accepted block.
	if lastAcceptedID == genesisBlock.ID() {
		genesisBlock.InitState(sb, genesisBlock)
		s.lastAcceptedBlock = genesisBlock
		s.preferred = genesisBlock
		return nil
	}

	// verify genesis data
	genesisID, err := s.GetBlockIDAtHeight(0)
	if err != nil {
		return fmt.Errorf("load genesis id failed: %v", err)
	}
	if genesisID != genesisBlock.ID() {
		return fmt.Errorf("invalid genesis data, expected genesis id %s", genesisID)
	}

	// load the last accepted block
	s.lastAcceptedBlock, err = s.GetBlock(lastAcceptedID)
	if err != nil {
		return fmt.Errorf("load last accepted block failed: %v", err)
	}

	s.lastAcceptedBlock.InitState(sb, s.lastAcceptedBlock)
	s.preferred = s.lastAcceptedBlock
	return nil
}

func (s *stateDB) SetState(state snow.State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch state {
	case snow.Bootstrapping:
		s.state = state
		return nil
	case snow.NormalOp:
		if err := s.verifyBootstrapped(nil); err != nil {
			return err
		}
		s.state = state
		return nil
	default:
		return snow.ErrUnknownState
	}
}

func (s *stateDB) verifyBootstrapped(prefer *Block) error {
	return nil
}

func (s *stateDB) IsBootstrapped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state == snow.NormalOp
}

// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (s *stateDB) LastAcceptedBlock() (*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastAcceptedBlock == nil {
		return nil, fmt.Errorf("stateDB not bootstrapped, no LastAcceptedBlock")
	}
	return s.lastAcceptedBlock, nil
}

func (s *stateDB) SetLastAccepted(blk *Block) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if blk.Parent() != s.preferred.ID() {
		return fmt.Errorf("invalid last accepted block parent: expected %v, got %v",
			s.preferred.ID(), blk.Parent())
	}
	if blk.Status() != choices.Accepted {
		return fmt.Errorf("invalid last accepted block status: expected %v, got %v",
			choices.Accepted, blk.Status())
	}

	s.acceptedBlocks[blk.ID()] = blk
	s.lastAcceptedBlock = blk
	return nil
}

func (s *stateDB) PreferredBlock() *Block {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.preferred == nil {
		panic(errPreferredBlock)
	}
	return s.preferred
}

// SetPreference persists the VM of the currently preferred block into database.
// This should always be a block that has no children known to consensus.
func (s *stateDB) SetPreference(id ids.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.preferred.ID() == id {
		return nil
	}

	if s.lastAcceptedBlock.ID() == id {
		return s.commitPreference(s.lastAcceptedBlock)
	}

	blk, err := s.GetBlock(id)
	if err != nil {
		return fmt.Errorf("SetPreference block %s error: %v", id, err)
	}

	if blk.Height() < s.preferred.Height() {
		// TODO: revert?
		// Should never happen
		return fmt.Errorf("SetPreference block %s error: can't revert block to %d", id, blk.Height())
	}

	if blk.Parent() == s.preferred.Parent() {
		// TODO: another preferred block?
		// Should never happen
		return fmt.Errorf("SetPreference block %s error: can't replace current preferred %d", id, s.preferred.ID())
	}

	if err := s.commitPreference(blk); err != nil {
		return err
	}

	s.lastAcceptedBlock = blk
	return err
}

func (s *stateDB) GetBlock(id ids.ID) (*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if blk, ok := s.acceptedBlocks[id]; ok && blk != nil {
		return blk, nil
	}

	data, err := s.blockDB.Get(id[:])
	if err != nil {
		return nil, err
	}

	return s.ParseBlock(data)
}

func (s *stateDB) ParseBlock(data []byte) (*Block, error) {
	blk, err := ParseBlock(data)
	if err != nil {
		return nil, err
	}
	blk.InitState(newStateBlock(s, s.db), s.preferred)
	return blk, nil
}

func (s *stateDB) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	return database.GetID(s.heightDB, database.PackUInt64(height))
}

func (s *stateDB) ChainConfig() *genesis.ChainConfig {
	return &s.genesis.ChainConfig
}

func (s *stateDB) FeeConfig() *genesis.FeeConfig {
	return &s.genesis.FeeConfig
}

func (s *stateDB) AddTxs(txs ...*ld.Transaction) {
	s.txPool.Add(txs...)
}

func (s *stateDB) commitPreference(blk *Block) error {
	parentId := ids.Empty
	if s.preferred != nil {
		parentId = s.preferred.ID()
	}

	if blk.Parent() != parentId {
		return fmt.Errorf("commitPreference parent error: expected %v, got %v",
			parentId, blk.Parent())
	}
	if blk.Status() != choices.Accepted {
		return fmt.Errorf("commitPreference block %s error: block not accepted", blk.ID())
	}
	if err := blk.State().Commit(); err != nil {
		return fmt.Errorf("commitPreference block %s error: %v", blk.ID(), err)
	}
	if err := database.PutID(s.lastAcceptedKeyDB, lastAcceptedKey, blk.ID()); err != nil {
		return err
	}
	s.preferred = blk

	for id, b := range s.acceptedBlocks {
		if b.Height() <= blk.Height() {
			delete(s.acceptedBlocks, id)
		}
	}
	return nil
}
