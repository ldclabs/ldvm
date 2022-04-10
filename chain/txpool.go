// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
type TxPool interface {
	Add(txs ...*ld.Transaction) error
	Has(txID ids.ID) bool
	Get(txID ids.ID) *ld.Transaction
	Len() int
	PopBySize(askSize uint64) []*ld.Transaction

	MarkDropped(txID ids.ID)
	WasDropped(txID ids.ID) bool

	AddNode(nodeID ids.ShortID)
	HasNode(nodeID ids.ShortID) bool

	PopMiners(blockId ids.ID, height uint64) []ids.ShortID
	DecisionMintTx(blockId ids.ID) *ld.Transaction // should always return
}

type txPool struct {
	activeNodes ids.ShortSet // TODO: expired
}
