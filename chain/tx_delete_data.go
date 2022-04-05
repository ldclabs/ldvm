// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxDeleteData struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxDeleteData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxDeleteData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxDeleteData) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxDeleteData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxDeleteData) SyntacticVerify() error {
	return nil
}

func (tx *TxDeleteData) Verify(blk *Block) error {
	return nil
}

func (tx *TxDeleteData) Accept(blk *Block) error {
	return nil
}
