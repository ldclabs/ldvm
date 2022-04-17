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
	acquaintedNodesCapacity = 60000
	acquaintedNodesTTL      = 600 // seconds
	rejectedTxsCapacity     = 100000
	rejectedTxsTTL          = 600
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
type TxPool interface {
	Len() int
	Has(txID ids.ID) bool
	Remove(txID ids.ID)
	Add(txs ...*ld.Transaction) error
	Get(txID ids.ID) *ld.Transaction
	PopTxsBySize(askSize int, threshold uint64) []*ld.Transaction
	Rejecte(*ld.Transaction)

	SeeNode(ids.ShortID)
	NodeAcquaintance(ids.ShortID) int

	ClearMintTxs(height uint64)
	SelectMiners(ids.ShortID, ids.ID, uint64) (*ld.Transaction, []ids.ShortID)
}

func NewTxPool() *txPool {
	return &txPool{
		txQueueSet:      ids.NewSet(1000),
		txQueue:         make([]*ld.Transaction, 0, 1000),
		mintTxs:         make([]*ld.Transaction, 0, 1000),
		mintTxsIDSet:    make(mintTxsIDSet, 1000),
		rejected:        collections.NewTTLMap(rejectedTxsCapacity),
		acquaintedNodes: collections.NewTTLMap(acquaintedNodesCapacity),
	}
}

type txPool struct {
	mu              sync.RWMutex
	txQueueSet      ids.Set
	txQueue         []*ld.Transaction
	mintTxs         []*ld.Transaction
	mintTxsIDSet    mintTxsIDSet
	rejected        *collections.TTLMap
	acquaintedNodes *collections.TTLMap
}

type mintTxsIDSet map[string]uint64

func (s mintTxsIDSet) has(data []byte) bool {
	_, ok := s[string(data)]
	return ok
}

func (s mintTxsIDSet) add(data []byte, height uint64) {
	s[string(data)] = height
}

func (s mintTxsIDSet) clear(height uint64) {
	for k, v := range s {
		if v < height {
			delete(s, k)
		}
	}
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

func (p *txPool) Get(txID ids.ID) *ld.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if v, ok := p.rejected.Get(string(txID[:])); ok {
		return v.(*ld.Transaction)
	}

	if p.txQueueSet.Contains(txID) {
		for _, tx := range p.txQueue {
			if tx.ID() == txID {
				return tx
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
		switch tx.Type {
		case ld.TypeMintFee:
			nid, _, _ := unwrapMintTxData(tx.Data)
			if p.NodeAcquaintance(nid) > 0 && !p.mintTxsIDSet.has(tx.Data) {
				tx.Status = choices.Unknown
				p.mintTxsIDSet.add(tx.Data, tx.Nonce)
				p.mintTxs = append(p.mintTxs, txs[i])
			}
		default:
			if !p.has(id) {
				tx.Status = choices.Unknown
				p.txQueueSet.Add(id)
				p.txQueue = append(p.txQueue, txs[i])
			}
		}
	}
	return nil
}

// Rejecte a tx that should not in pool.
func (p *txPool) Rejecte(tx *ld.Transaction) {
	id := tx.ID()
	tx.Status = choices.Rejected
	p.rejected.Set(string(id[:]), tx, rejectedTxsTTL)
}

func (p *txPool) SeeNode(id ids.ShortID) {
	p.acquaintedNodes.Increment(string(id[:]), 1, acquaintedNodesTTL)
}

func (p *txPool) NodeAcquaintance(id ids.ShortID) int {
	i, _, _ := p.acquaintedNodes.GetInt(string(id[:]))
	return i
}

func (p *txPool) ClearMintTxs(height uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := 0
	for i, tx := range p.mintTxs {
		if tx.Nonce >= height {
			p.mintTxs[n] = p.mintTxs[i]
			n++
		}
	}
	p.mintTxs = p.mintTxs[:n]
	p.mintTxsIDSet.clear(height)
}

func (p *txPool) SelectMiners(nodeID ids.ShortID, blkID ids.ID, height uint64) (
	*ld.Transaction, []ids.ShortID) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var mintTx *ld.Transaction
	var preferredMintTx *ld.Transaction

	mintTxs := make(map[ids.ShortID]uint64, len(p.mintTxs)/2)
	for i, tx := range p.mintTxs {
		nid, bid, _ := unwrapMintTxData(tx.Data)
		if bid != blkID {
			continue
		}

		if mintTx == nil {
			mintTx = p.mintTxs[i]
		}
		if preferredMintTx == nil && nid == nodeID {
			preferredMintTx = p.mintTxs[i]
		}
		mintTxs[tx.To] += uint64(p.NodeAcquaintance(nid))
	}

	if preferredMintTx == nil {
		preferredMintTx = mintTx
	}

	miners := make([]ids.ShortID, 0, len(mintTxs))
	for k := range mintTxs {
		miners = append(miners, k)
	}

	sort.SliceStable(miners, func(i, j int) bool {
		iv := mintTxs[miners[i]]
		jv := mintTxs[miners[j]]
		if iv == jv {
			return bytes.Compare(miners[i][:], miners[j][:]) == -1
		}
		return iv > jv
	})

	if len(miners) > ld.MaxMiners {
		miners = miners[:ld.MaxMiners]
	}

	sort.SliceStable(miners, func(i, j int) bool {
		return bytes.Compare(miners[i][:], miners[j][:]) == -1
	})
	return preferredMintTx, miners
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
