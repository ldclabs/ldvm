// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/mailgun/holster/v4/collections"

	"github.com/ldclabs/ldvm/ld"
)

const (
	knownTxsCapacity = 1000000
	knownTxsTTL      = 600 // seconds
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
type TxPool interface {
	Len() int
	SetTxsHeight(height uint64, txIDs ...ids.ID)
	ClearTxs(txIDs ...ids.ID)
	AddRemote(txs ...*ld.Transaction)
	AddLocal(txs ...*ld.Transaction)
	GetHeight(txID ids.ID) int64
	PopTxsBySize(askSize int) ld.Txs
	Reject(tx *ld.Transaction)
	Clear()
}

// NewTxPool creates a new transaction pool.
func NewTxPool() *txPool {
	return &txPool{
		txQueueSet:     ids.NewSet(10000),
		txQueue:        make([]*ld.Transaction, 0, 10000),
		knownTxs:       &knownTxs{collections.NewTTLMap(knownTxsCapacity)},
		signalTxsReady: func() {},                   // initially noop
		gossipTx:       func(tx *ld.Transaction) {}, // initially noop
	}
}

type txPool struct {
	mu             sync.RWMutex
	txQueueSet     ids.Set
	txQueue        ld.Txs
	knownTxs       *knownTxs
	signalTxsReady func()
	gossipTx       func(tx *ld.Transaction)
}

type knownTxs struct {
	cache *collections.TTLMap
}

func (k *knownTxs) getHeight(txID ids.ID) int64 {
	if s, ok := k.cache.Get(string(txID[:])); ok {
		return s.(int64)
	}
	return -3
}

func (k *knownTxs) setHeight(txID ids.ID, height int64) {
	k.cache.Set(string(txID[:]), height, knownTxsTTL)
}

func (k *knownTxs) clear() {
	i := 100
	for i == 100 {
		i = k.cache.RemoveExpired(i)
	}
}

// Len returns the number of transactions in the pool.
func (p *txPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := 0
	for _, tx := range p.txQueue {
		switch {
		case tx.IsBatched():
			n += len(tx.Txs())
		default:
			n += 1
		}
	}
	return n
}

// GetHeight returns the height of block that transactions included in.
// -3: the transaction is not in the pool
// -2: the transaction is rejected
// -1: the transaction is in the pool, but not yet included in a block
// >= 0: the transaction is included in a block with the given height
func (p *txPool) GetHeight(txID ids.ID) int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.knownTxs.getHeight(txID)
}

// SetTxsHeight sets the height of block that transactions included in.
// It should be called only when block accepted.
func (p *txPool) SetTxsHeight(height uint64, txIDs ...ids.ID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, txID := range txIDs {
		p.knownTxs.setHeight(txID, int64(height))
	}
}

// ClearTxs removes transaction entities from the pool.
// but their ids and status can be retrieved.
func (p *txPool) ClearTxs(txIDs ...ids.ID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, txID := range txIDs {
		p.clear(txID)
	}
}

func (p *txPool) clear(txID ids.ID) {
	if !p.txQueueSet.Contains(txID) {
		return
	}

	for i, tx := range p.txQueue {
		if tx.ID == txID {
			p.txQueueSet.Remove(txID)
			n := copy(p.txQueue[i:], p.txQueue[i+1:])
			p.txQueue = p.txQueue[:i+n]
			return
		}
	}
}

func (p *txPool) knownTx(txID ids.ID) bool {
	if p.txQueueSet.Contains(txID) {
		return true
	}

	return p.knownTxs.getHeight(txID) >= -2
}

// AddRemote adds transaction entities from the network (API or AppGossip).
// The transaction already in the pool will not be added.
// The added transaction will be gossiped to the network.
// Transaction should be syntactic verified before adding.
func (p *txPool) AddRemote(txs ...*ld.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	added := 0
	for i, tx := range txs {
		if !p.knownTx(tx.ID) {
			added++
			p.knownTxs.setHeight(tx.ID, int64(-1))
			p.txQueueSet.Add(tx.ID)
			p.txQueue = append(p.txQueue, txs[i])

			go p.gossipTx(txs[i])
		}
	}

	if added > 0 {
		p.signalTxsReady()
	}
}

// AddLocal adds transaction entities from block build failing.
// So that the transaction can be process again.
func (p *txPool) AddLocal(txs ...*ld.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, tx := range txs {
		if p.knownTxs.getHeight(tx.ID) <= -2 {
			// transaction expired or rejected, ignore it.
			continue
		}

		if !p.txQueueSet.Contains(tx.ID) {
			p.txQueueSet.Add(tx.ID)
			p.txQueue = append(p.txQueue, txs[i])
		}
	}
}

// Reject removes a transaction from the pool, and sets the status to rejected.
func (p *txPool) Reject(tx *ld.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clear(tx.ID)
	p.knownTxs.setHeight(tx.ID, int64(-2))
}

// PopTxsBySize sorts transactions by priority and returns
// a batch of high priority transactions with the given size.
// The returned transactions are removed from the pool.
func (p *txPool) PopTxsBySize(askSize int) ld.Txs {
	if uint64(askSize) < 100 {
		return ld.Txs{}
	}

	rt := make(ld.Txs, 0, 64)
	p.mu.Lock()
	defer p.mu.Unlock()

	p.txQueue.Sort()
	total := 0
	n := 0
	for i, tx := range p.txQueue {
		total += tx.BytesSize()
		if total > askSize {
			break
		}
		n++
		p.txQueueSet.Remove(tx.ID)
		rt = append(rt, p.txQueue[i])
	}
	if n > 0 {
		n = copy(p.txQueue, p.txQueue[n:])
		p.txQueue = p.txQueue[:n]
	}
	return rt
}

func (p *txPool) Clear() {
	p.knownTxs.clear()
}
