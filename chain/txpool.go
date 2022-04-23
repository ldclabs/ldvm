// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"sort"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/mailgun/holster/v4/collections"

	"github.com/ldclabs/ldvm/ld"
)

const (
	rejectedTxsCapacity = 100000
	rejectedTxsTTL      = 600
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
type TxPool interface {
	Len() int
	Has(txID ids.ID) bool
	Remove(txID ids.ID)
	Add(txs ...*ld.Transaction) error
	Get(txID ids.ID) Transaction
	PopTxsBySize(askSize int, threshold uint64) []*ld.Transaction
	Rejecte(*ld.Transaction)
}

func NewTxPool() *txPool {
	return &txPool{
		txQueueSet: ids.NewSet(1000),
		txQueue:    make([]*ld.Transaction, 0, 1000),
		rejected:   collections.NewTTLMap(rejectedTxsCapacity),
	}
}

type txPool struct {
	mu         sync.RWMutex
	txQueueSet ids.Set
	txQueue    []*ld.Transaction
	rejected   *collections.TTLMap
}

func (p *txPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.txQueue)
}

func (p *txPool) Has(txID ids.ID) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.has(txID)
}

func (p *txPool) Remove(txID ids.ID) {
	p.mu.RLock()
	if !p.txQueueSet.Contains(txID) {
		p.mu.RUnlock()
		return
	}
	p.mu.RUnlock()
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, tx := range p.txQueue {
		if tx.ID() == txID {
			p.txQueueSet.Remove(txID)
			n := copy(p.txQueue[i:], p.txQueue[i+1:])
			p.txQueue = p.txQueue[:i+n]
			return
		}
	}
}

func (p *txPool) has(txID ids.ID) bool {
	if p.txQueueSet.Contains(txID) {
		return true
	}
	_, ok := p.rejected.Get(string(txID[:]))
	return ok
}

func (p *txPool) Get(txID ids.ID) Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if v, ok := p.rejected.Get(string(txID[:])); ok {
		if tx, _ := NewTx(v.(*ld.Transaction), false); tx != nil {
			tx.SetStatus(choices.Rejected)
			return tx
		}
		return nil
	}

	if p.txQueueSet.Contains(txID) {
		for _, tx := range p.txQueue {
			if tx.ID() == txID {
				if ntx, _ := NewTx(tx, false); ntx != nil {
					ntx.SetStatus(choices.Unknown)
					return ntx
				}
				return nil
			}
		}
	}
	return nil
}

// txs should be syntactic verified before adding
func (p *txPool) Add(txs ...*ld.Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, tx := range txs {
		id := tx.ID()
		if !p.has(id) {
			p.txQueueSet.Add(id)
			p.txQueue = append(p.txQueue, txs[i])
		}
	}
	return nil
}

// Rejecte a tx that should not in pool.
func (p *txPool) Rejecte(tx *ld.Transaction) {
	id := tx.ID()
	p.rejected.Set(string(id[:]), tx, rejectedTxsTTL)
}

func (p *txPool) PopTxsBySize(askSize int, threshold uint64) []*ld.Transaction {
	rt := make([]*ld.Transaction, 0, 64)
	p.mu.Lock()
	defer p.mu.Unlock()

	now := uint64(time.Now().Unix())
	for _, tx := range p.txQueue {
		tx.SetPriority(threshold, now)
	}

	sort.SliceStable(p.txQueue, func(i, j int) bool {
		if p.txQueue[i].From == p.txQueue[j].From {
			return p.txQueue[i].Nonce < p.txQueue[j].Nonce
		}
		if p.txQueue[i].Priority == p.txQueue[j].Priority {
			return bytes.Compare(p.txQueue[i].Bytes(), p.txQueue[j].Bytes()) == -1
		}
		return p.txQueue[i].Priority > p.txQueue[j].Priority
	})

	total := 0
	n := 0
	for i, tx := range p.txQueue {
		total += len(tx.Bytes())
		if total > askSize {
			break
		}
		n++
		p.txQueueSet.Remove(tx.ID())
		rt = append(rt, p.txQueue[i])
	}
	if n > 0 {
		n = copy(p.txQueue, p.txQueue[n:])
		p.txQueue = p.txQueue[:n]
	}
	return rt
}
