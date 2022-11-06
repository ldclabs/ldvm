// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
)

type PushNetwork struct {
	vm *VM
}

func (v *VM) NewPushNetwork() {
	v.network = &PushNetwork{v}
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
// AppGossip(ctx context.Context, nodeID ids.NodeID, msg []byte) may be called multiple times.
func (v *VM) AppGossip(ctx context.Context, nodeID ids.NodeID, msg []byte) error {
	return nil
}
