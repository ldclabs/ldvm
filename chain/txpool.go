// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/mailgun/holster/v4/collections"

	"github.com/ldclabs/ldvm/chain/transaction"
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
	Add(txs ...*ld.Transaction)
	Get(txID ids.ID) transaction.Transaction
	PopTxsBySize(askSize int) ld.Txs
	Reject(*ld.Transaction)
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
	txQueue    ld.Txs
	rejected   *collections.TTLMap
}

func (p *txPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	n := 0
	for _, tx := range p.txQueue {
		switch {
		case tx.IsBatched():
			for _, tx2 := range tx.Txs() {
				if tx2.Type != ld.TypeTest {
					n += 1
				}
			}
		default:
			n += 1
		}
	}
	return n
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
		if tx.ID == txID {
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

func (p *txPool) Get(txID ids.ID) transaction.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if v, ok := p.rejected.Get(string(txID[:])); ok {
		if tx, _ := transaction.NewTx(v.(*ld.Transaction), false); tx != nil {
			tx.SetStatus(choices.Rejected)
			return tx
		}
		return nil
	}

	if p.txQueueSet.Contains(txID) {
		for _, tx := range p.txQueue {
			if tx.ID == txID {
				if ntx, _ := transaction.NewTx(tx, false); ntx != nil {
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
func (p *txPool) Add(txs ...*ld.Transaction) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, tx := range txs {
		if tx.Type != ld.TypeTest && !p.has(tx.ID) {
			p.txQueueSet.Add(tx.ID)
			p.txQueue = append(p.txQueue, txs[i])
		}
	}
}

// Rejecte a tx that should not in pool.
func (p *txPool) Reject(tx *ld.Transaction) {
	p.Remove(tx.ID)
	p.rejected.Set(string(tx.ID[:]), tx, rejectedTxsTTL)
}

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
