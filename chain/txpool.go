// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
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
	SetTxsStatus(status choices.Status, txIDs ...ids.ID)
	ClearTxs(txIDs ...ids.ID)
	AddRemote(txs ...*ld.Transaction)
	AddLocal(txs ...*ld.Transaction)
	GetStatus(txID ids.ID) choices.Status
	PopTxsBySize(askSize int) ld.Txs
	Reject(tx *ld.Transaction)
}

// NewTxPool creates a new transaction pool.
func NewTxPool() *txPool {
	return &txPool{
		txQueueSet:     ids.NewSet(10000),
		txQueue:        make([]*ld.Transaction, 0, 10000),
		knownTxs:       collections.NewTTLMap(knownTxsCapacity),
		signalTxsReady: func() {},                   // initially noop
		gossipTx:       func(tx *ld.Transaction) {}, // initially noop
	}
}

type txPool struct {
	mu             sync.RWMutex
	txQueueSet     ids.Set
	txQueue        ld.Txs
	knownTxs       *collections.TTLMap
	signalTxsReady func()
	gossipTx       func(tx *ld.Transaction)
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

// GetStatus returns the status of a transaction by ID.
func (p *txPool) GetStatus(txID ids.ID) choices.Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if s, ok := p.knownTxs.Get(string(txID[:])); ok {
		return s.(choices.Status)
	}
	return choices.Unknown
}

// SetTxsStatus sets the status of a batchs transactions.
func (p *txPool) SetTxsStatus(status choices.Status, txIDs ...ids.ID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.setTxsStatus(status, txIDs...)
}

func (p *txPool) setTxsStatus(status choices.Status, txIDs ...ids.ID) {
	for _, txID := range txIDs {
		p.knownTxs.Set(string(txID[:]), status, knownTxsTTL)
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

func (p *txPool) has(txID ids.ID) bool {
	return p.txQueueSet.Contains(txID)
}

func (p *txPool) knownTx(txID ids.ID) bool {
	if p.txQueueSet.Contains(txID) {
		return true
	}

	_, ok := p.knownTxs.Get(string(txID[:]))
	return ok
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
			p.knownTxs.Set(string(tx.ID[:]), choices.Unknown, knownTxsTTL)
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
		if !p.txQueueSet.Contains(tx.ID) {
			p.knownTxs.Set(string(tx.ID[:]), choices.Unknown, knownTxsTTL)
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
	p.knownTxs.Set(string(tx.ID[:]), choices.Rejected, knownTxsTTL)
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
