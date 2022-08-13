// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/ava-labs/avalanchego/ids"
	"go.uber.org/zap"

	"github.com/ldclabs/ldvm/ld"
)

type PushNetwork struct {
	vm *VM
}

func (v *VM) NewPushNetwork() {
	v.network = &PushNetwork{v}
}

func (n *PushNetwork) GossipTx(tx *ld.Transaction) {
	if n.vm.appSender == nil || tx == nil {
		return
	}

	var err error
	var data []byte

	// it should be a batch tx when txs length is greater than 1
	if tx.IsBatched() {
		data, err = tx.Txs().Marshal()
	} else {
		data, err = ld.Txs{tx}.Marshal()
	}

	if err != nil {
		n.vm.Log.Warn("PushNetwork.GossipTx marshal txs failed", zap.Error(err))
		return
	}

	n.vm.Log.Debug("PushNetwork.GossipTx SendAppGossip", zap.Int("bytes", len(data)))
	if err = n.vm.appSender.SendAppGossip(data); err != nil {
		n.vm.Log.Warn("PushNetwork.GossipTx SendAppGossip failed", zap.Error(err))
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
func (v *VM) AppGossip(nodeID ids.NodeID, msg []byte) error {
	txs := ld.Txs{}
	var err error
	var tx *ld.Transaction

	if err = txs.Unmarshal(msg); err == nil {
		v.Log.Info("LDVM.AppGossip",
			zap.Stringer("nodeID", nodeID),
			zap.Int("bytes", len(msg)),
			zap.Int("txs", len(txs)))

		if tx, err = txs.To(); err == nil {
			err = v.bc.AddRemoteTxs(tx)
		}
	}

	if err != nil {
		v.Log.Warn("LDVM.AppGossip",
			zap.Stringer("nodeID", nodeID),
			zap.Int("bytes", len(msg)),
			zap.Error(err))
	}
	return err
}
