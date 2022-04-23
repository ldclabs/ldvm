// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/mailgun/holster/v4/collections"

	"github.com/ldclabs/ldvm/ld"
)

const (
	gossipedTxsCapacity = 100000
	gossipedTxsTTL      = 120 // seconds
	txsMsgMaxSize       = 4 * 1024 * 1024
)

type PushNetwork struct {
	mu          sync.Mutex
	vm          *VM
	ticker      *time.Ticker
	queue       []*ld.Transaction
	gossipedTxs *collections.TTLMap
}

func (v *VM) NewPushNetwork() {
	v.network = &PushNetwork{
		vm:          v,
		queue:       make([]*ld.Transaction, 0, 1000),
		gossipedTxs: collections.NewTTLMap(gossipedTxsCapacity),
	}
}

func (n *PushNetwork) seeTx(id ids.ID) {
	n.gossipedTxs.Set(string(id[:]), struct{}{}, gossipedTxsTTL)
}

func (n *PushNetwork) hasTx(id ids.ID) bool {
	_, ok := n.gossipedTxs.Get(string(id[:]))
	return ok
}

func (n *PushNetwork) sendTxs(txs ...*ld.Transaction) error {
	if len(txs) == 0 || n.vm.appSender == nil {
		return nil
	}

	data, err := ld.MarshalTxs(txs)
	if err != nil {
		n.vm.Log.Warn("PushNetwork marshal txs failed: %v", err)
		return err
	}

	n.vm.Log.Info("PushNetwork sendTxs %d bytes, %d txs", len(data), len(txs))
	if err = n.vm.appSender.SendAppGossip(data); err != nil {
		n.vm.Log.Warn("PushNetwork sendTxs failed: %v", err)
	}
	return err
}

func (n *PushNetwork) Start(du time.Duration) {
	n.ticker = time.NewTicker(du)
	for range n.ticker.C {
		n.mu.Lock()
		if len(n.queue) == 0 {
			n.mu.Unlock()
			continue
		}

		txs := make([]*ld.Transaction, 0, len(n.queue))
		size := 0
		for i, tx := range n.queue {
			size += len(tx.Bytes())
			if size > txsMsgMaxSize {
				break
			}
			txs = append(txs, n.queue[i])
		}
		l := copy(n.queue[:0], n.queue[len(txs):])
		n.queue = n.queue[:l]
		n.mu.Unlock()
		n.sendTxs(txs...)
	}
}

func (n *PushNetwork) Close() error {
	n.ticker.Stop()
	n.sendTxs(n.queue...)
	return nil
}

func (n *PushNetwork) GossipTx(tx *ld.Transaction) {
	if n.vm.appSender == nil {
		return
	}
	if id := tx.ID(); !n.hasTx(id) {
		n.seeTx(id)
		n.mu.Lock()
		n.queue = append(n.queue, tx)
		n.mu.Unlock()
	}
}

// AppGossip implements the common.VM AppHandler AppGossip interface
// This VM doesn't (currently) have any app-specific messages
//
// Notify this engine of a gossip message from [nodeID].
//
// The meaning of [msg] is application (VM) specific, and the VM defines how
// to react to this message.
//
// This message is not expected in response to any event, and it does not
// need to be responded to.
//
// A node may gossip the same message multiple times. That is,
// AppGossip([nodeID], [msg]) may be called multiple times.
func (v *VM) AppGossip(nodeID ids.ShortID, msg []byte) error {
	txs, err := ld.UnmarshalTxs(msg)
	if len(txs) > 0 {
		v.Log.Info("AppGossip from %s, %d bytes, %d txs", nodeID, len(msg), len(txs))
		rt := make([]*ld.Transaction, 0, len(txs))
		for i := range txs {
			id := txs[i].ID()
			if !v.network.hasTx(id) {
				v.network.seeTx(id)
				rt = append(rt, txs[i])
			}
		}
		v.state.AddTxs(true, rt...)
		return nil
	}

	v.Log.Warn("AppGossip %s, %d bytes, error: %v", nodeID, len(msg), err)
	return err
}
